package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/masterzen/winrm"
	"github.com/patrickmn/go-cache"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Connection pool for WinRM clients
type ConnectionPool struct {
	clients     map[string]*winrm.Client
	mutex       sync.Mutex
	maxSize     int
	idleTimeout time.Duration
	lastUsed    map[string]time.Time
}

// Add these missing variables and types
var (
	configFile     string
	config         Config
	connectionPool ConnectionPool
	discoveryCache *cache.Cache
	resultChannel  chan DiscoveryResult
	completedJobs  int32
	totalJobs      int32
	jobsMutex      sync.Mutex
	progressTicker *time.Ticker
	progressDone   chan bool
	getClient      = func(server ServerConfig) (*winrm.Client, error) {
		endpoint := winrm.NewEndpoint(
			server.Hostname,
			server.Port,
			server.UseHTTPS,
			config.SkipCertVerify,
			nil,
			nil,
			nil,
			30*time.Second,
		)

		client, err := winrm.NewClient(endpoint, server.Username, server.Password)
		if err != nil {
			return nil, err
		}

		return client, nil
	}
	resourceCtrl ResourceController
	workers      []*WorkerNode
)

// ResourceController manages system resources
type ResourceController struct {
	CPUThreshold    float64
	MemoryThreshold float64
	lastCheck       time.Time
	checkInterval   time.Duration
}

// waitForResources waits until system resources are below thresholds
func (rc *ResourceController) waitForResources() {
	// Don't check too frequently
	if time.Since(rc.lastCheck) < rc.checkInterval {
		return
	}

	for {
		cpuUsage, err := getCPUUsage()
		if err != nil {
			log.Printf("Warning: Failed to get CPU usage: %v", err)
			return
		}

		memUsage, err := getMemoryUsage()
		if err != nil {
			log.Printf("Warning: Failed to get memory usage: %v", err)
			return
		}

		if cpuUsage < rc.CPUThreshold && memUsage < rc.MemoryThreshold {
			break
		}

		log.Printf("Resource usage high (CPU: %.1f%%, Memory: %.1f%%), waiting before starting next batch...",
			cpuUsage, memUsage)
		time.Sleep(5 * time.Second)
	}

	rc.lastCheck = time.Now()
}

// Helper functions to get resource usage
func getCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return percentages[0], nil
}

func getMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return vmStat.UsedPercent, nil
}

// WorkerNode represents a worker node in the system
type WorkerNode struct {
	ID          string    `json:"id"`
	Hostname    string    `json:"hostname"`
	IPAddress   string    `json:"ip_address"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	JobsHandled int       `json:"jobs_handled"`
	currentJobs int32     // Used internally for load balancing
}

func init() {
	// Parse command line flags
	flag.StringVar(&configFile, "config", "config.json", "Path to configuration file")
	flag.Parse()

	// Initialize cache
	discoveryCache = cache.New(30*time.Minute, 10*time.Minute)

	// Initialize result channel
	resultChannel = make(chan DiscoveryResult, 100)

	// Initialize connection pool
	connectionPool = ConnectionPool{
		clients:     make(map[string]*winrm.Client),
		lastUsed:    make(map[string]time.Time),
		maxSize:     10,
		idleTimeout: 10 * time.Minute,
	}

	// Initialize progress channel
	progressDone = make(chan bool)
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Load configuration
	err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(config.OutputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize database connection
	err = initDatabase()
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
	}
	defer closeDatabase()

	// Load PowerShell script
	scriptContent, err := loadPowerShellScript()
	if err != nil {
		log.Fatalf("Failed to load PowerShell script: %v", err)
	}

	// Start metrics server
	// startMetricsServer() // Removed for simplification

	// Initialize tracing
	// initTracing() // Removed for simplification

	// Start API server
	startAPIServer()

	// Start progress reporting
	startProgressReporting()
	defer func() {
		progressTicker.Stop()
		progressDone <- true
	}()

	// Process servers
	processServers(scriptContent)

	// Wait for all results to be processed
	close(resultChannel)
	collectResults()

	log.Println("Server discovery completed successfully")
}

// Load PowerShell script from file
func loadPowerShellScript() (string, error) {
	scriptBytes, err := os.ReadFile(config.PowerShellScript)
	if err != nil {
		return "", fmt.Errorf("error reading PowerShell script: %w", err)
	}
	return string(scriptBytes), nil
}

// Execute discovery on a server
func executeDiscovery(server ServerConfig, scriptContent string) DiscoveryResult {
	serverKey := fmt.Sprintf("%s:%d", server.Hostname, server.Port)
	startTime := time.Now()

	// Check cache first
	if cachedResult, found := discoveryCache.Get(serverKey); found {
		log.Printf("Using cached result for %s", serverKey)
		result := cachedResult.(DiscoveryResult)
		result.Message = "Retrieved from cache"
		return result
	}

	result := DiscoveryResult{
		Server:    serverKey,
		StartTime: startTime,
	}

	// Get client from pool
	client, err := getClient(server)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Connection error: %v", err)
		result.EndTime = time.Now()
		return result
	}

	// Create a unique output directory for this server
	timestamp := time.Now().Format("20060102-150405")
	serverOutputDir := filepath.Join(config.OutputDir, server.Hostname+"-"+timestamp)

	// Execute the PowerShell script
	log.Printf("Executing discovery on %s", serverKey)

	// Create a buffer to capture output
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer

	// Execute the script
	command := fmt.Sprintf("$OutputFolder = \"%s\"; %s",
		strings.ReplaceAll(serverOutputDir, "\\", "\\\\"),
		scriptContent)

	exitCode, err := runCommand(client, command, &outputBuffer, &errorBuffer)

	result.EndTime = time.Now()

	if err != nil || exitCode != 0 {
		result.Success = false
		result.Error = fmt.Sprintf("Execution error (exit code %d): %v\n%s",
			exitCode, err, errorBuffer.String())
		return result
	}

	// Create local directory to store results
	if err := os.MkdirAll(serverOutputDir, 0755); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to create output directory: %v", err)
		return result
	}

	// Save output to file
	outputFile := filepath.Join(serverOutputDir, "execution_output.txt")
	if err := os.WriteFile(outputFile, outputBuffer.Bytes(), 0644); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to write output file: %v", err)
		return result
	}

	// Try to copy the ZIP file from the remote server
	zipFileName := fmt.Sprintf("ServerDiscovery-%s.zip", timestamp)
	remoteZipPath := fmt.Sprintf("$env:USERPROFILE\\Documents\\%s", zipFileName)
	localZipPath := filepath.Join(serverOutputDir, zipFileName)

	// Get the content of the ZIP file
	var zipBuffer bytes.Buffer
	_, err = runCommand(client, fmt.Sprintf("if (Test-Path %s) { [System.IO.File]::ReadAllBytes('%s') }",
		remoteZipPath, strings.ReplaceAll(remoteZipPath, "'", "''")), &zipBuffer, &errorBuffer)

	if err == nil && zipBuffer.Len() > 0 {
		// Save the ZIP file locally
		if err := os.WriteFile(localZipPath, zipBuffer.Bytes(), 0644); err != nil {
			log.Printf("Warning: Failed to save ZIP file for %s: %v", serverKey, err)
		} else {
			log.Printf("Successfully saved ZIP file for %s", serverKey)
		}
	} else {
		log.Printf("Warning: Could not retrieve ZIP file from %s: %v", serverKey, err)
	}

	result.Success = true
	result.Message = "Discovery completed successfully"
	result.OutputPath = serverOutputDir

	// Cache the result
	discoveryCache.Set(serverKey, result, cache.DefaultExpiration)

	return result
}

// Retrieve discovery files from remote server
func retrieveDiscoveryFiles(client *winrm.Client, serverKey, timestamp, localDir string) bool {
	remoteDir := fmt.Sprintf("$env:USERPROFILE\\Documents\\ServerDiscovery-%s", timestamp)

	// Get list of files in the remote directory
	var outputBuffer, errorBuffer bytes.Buffer
	listCommand := fmt.Sprintf("if (Test-Path %s) { Get-ChildItem -Path %s -File | Select-Object -ExpandProperty Name | ConvertTo-Json }",
		remoteDir, remoteDir)

	exitCode, err := runCommand(client, listCommand, &outputBuffer, &errorBuffer)
	if err != nil || exitCode != 0 {
		log.Printf("Warning: Could not list files in remote directory for %s: %v", serverKey, err)
		return false
	}

	// Parse file list
	var fileList []string
	if err := json.Unmarshal(outputBuffer.Bytes(), &fileList); err != nil {
		log.Printf("Warning: Could not parse file list for %s: %v", serverKey, err)
		return false
	}

	// Download each file
	success := true
	for _, fileName := range fileList {
		var fileBuffer, errBuffer bytes.Buffer
		remoteFilePath := fmt.Sprintf("%s\\%s", remoteDir, fileName)

		// Get file content with compression to reduce network transfer
		getFileCmd := fmt.Sprintf("if (Test-Path '%s') { [System.IO.File]::ReadAllBytes('%s') }",
			remoteFilePath, strings.ReplaceAll(remoteFilePath, "'", "''"))

		_, err := runCommand(client, getFileCmd, &fileBuffer, &errBuffer)
		if err != nil || fileBuffer.Len() == 0 {
			log.Printf("Warning: Could not retrieve file %s from %s: %v", fileName, serverKey, err)
			success = false
			continue
		}

		// Save the file locally
		localFilePath := filepath.Join(localDir, fileName)
		if err := os.WriteFile(localFilePath, fileBuffer.Bytes(), 0644); err != nil {
			log.Printf("Warning: Failed to save file %s for %s: %v", fileName, serverKey, err)
			success = false
		}
	}

	return success
}

// Collect and process results
func collectResults() {
	resultsFile := filepath.Join(config.OutputDir, "discovery_results.json")
	var results []DiscoveryResult

	// Process results as they arrive
	for result := range resultChannel {
		results = append(results, result)

		// Log the result
		if result.Success {
			log.Printf("Discovery on %s completed successfully in %v (region: %s)",
				result.Server, result.EndTime.Sub(result.StartTime), result.Region)
		} else {
			log.Printf("Discovery on %s failed: %s (region: %s)",
				result.Server, result.Error, result.Region)
		}

		// Store in database if enabled
		if config.DatabaseConfig.Enabled {
			go storeResultInDatabase(result)
		}
	}

	// Write results to file with compression
	compressedResultsFile := resultsFile + ".gz"
	file, err := os.Create(compressedResultsFile)
	if err != nil {
		log.Printf("Error creating compressed results file: %v", err)
		return
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	encoder := json.NewEncoder(gzWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Printf("Error encoding results: %v", err)
		return
	}

	log.Printf("Results written to %s", compressedResultsFile)
}

// Start progress reporting
func startProgressReporting() {
	progressTicker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-progressTicker.C:
				completed := atomic.LoadInt32(&completedJobs)
				total := atomic.LoadInt32(&totalJobs)
				if total > 0 {
					progress := float64(completed) / float64(total) * 100
					log.Printf("Progress: %.1f%% (%d/%d servers completed)",
						progress, completed, total)

					// Report resource usage
					cpuUsage, _ := getCPUUsage()
					memUsage, _ := getMemoryUsage()
					log.Printf("Resource usage - CPU: %.1f%%, Memory: %.1f%%",
						cpuUsage, memUsage)
				}
			case <-progressDone:
				return
			}
		}
	}()
}

// Add these missing functions
func executeWithRetry(server ServerConfig, scriptContent string) DiscoveryResult {
	// Simple implementation for now
	return executeDiscovery(server, scriptContent)
}

func determineBatchSize(region string) int {
	// Simple implementation for now
	return 5
}

func getLeastBusyWorker(workers []*WorkerNode) *WorkerNode {
	if len(workers) == 0 {
		return nil
	}

	leastBusy := workers[0]
	for _, worker := range workers {
		if worker.currentJobs < leastBusy.currentJobs {
			leastBusy = worker
		}
	}
	return leastBusy
}

// Process servers with worker pool
func processServers(scriptContent string) {
	// Process servers in batches
	totalJobs = int32(len(config.Servers))
	completedJobs = 0

	// Create a worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Concurrency)

	// Group servers by region for more efficient processing
	regionServers := make(map[string][]ServerConfig)
	for _, server := range config.Servers {
		regionServers[server.Region] = append(regionServers[server.Region], server)
	}

	// Process each region
	for region, servers := range regionServers {
		log.Printf("Processing %d servers in region %s", len(servers), region)

		// Get batch size for this region
		batchSize := determineBatchSize(region)

		// Process in batches
		for i := 0; i < len(servers); i += batchSize {
			end := i + batchSize
			if end > len(servers) {
				end = len(servers)
			}

			batch := servers[i:end]
			log.Printf("Processing batch of %d servers in region %s", len(batch), region)

			// Process batch
			for _, server := range batch {
				wg.Add(1)
				semaphore <- struct{}{} // Acquire semaphore

				go func(server ServerConfig) {
					defer wg.Done()
					defer func() { <-semaphore }() // Release semaphore

					// Execute discovery with retry
					result := executeWithRetry(server, scriptContent)
					resultChannel <- result

					// Update progress
					atomic.AddInt32(&completedJobs, 1)
				}(server)
			}

			// Check resource usage before starting next batch
			resourceCtrl.waitForResources()
		}
	}

	// Wait for all workers to complete
	wg.Wait()

	log.Println("All discovery jobs completed")
}

// ServerDiscoveryController handles server discovery operations
type ServerDiscoveryController struct {
	// Add any fields needed for the controller
}

// runDiscoveryScript runs the PowerShell discovery script on a Windows server
func (c *ServerDiscoveryController) runDiscoveryScript(server ServerConfig, outputDir string) (string, error) {
	// Create WinRM client
	client, err := getClient(server)
	if err != nil {
		return "", fmt.Errorf("failed to create WinRM client: %w", err)
	}

	// Create output directory for this server
	serverOutputDir := filepath.Join(outputDir, server.Hostname)
	if err := os.MkdirAll(serverOutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load PowerShell script
	scriptContent, err := loadPowerShellScript()
	if err != nil {
		return "", fmt.Errorf("failed to load PowerShell script: %w", err)
	}

	// Execute script on server
	result, err := executeScript(client, server.Hostname, scriptContent, serverOutputDir)
	if err != nil {
		return "", fmt.Errorf("failed to execute script: %w", err)
	}

	return result, nil
}

// RunDiscovery runs a discovery on a server
func (c *ServerDiscoveryController) RunDiscovery(w http.ResponseWriter, r *http.Request) {
	// Parse server ID from request
	vars := mux.Vars(r)
	serverIDStr := vars["id"]
	serverID, err := strconv.Atoi(serverIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid server ID: %s", serverIDStr), http.StatusBadRequest)
		return
	}

	// Get server configuration
	server, err := getServerByID(serverID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Server not found: %s", err), http.StatusNotFound)
		return
	}

	// Create discovery record
	discoveryID, err := createDiscoveryRecord(serverID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create discovery record: %s", err), http.StatusInternalServerError)
		return
	}

	// Run discovery in background
	go func() {
		startTime := time.Now()
		var outputPath string
		var success bool
		var message string

		// Run discovery script
		outputPath, err = c.runDiscoveryScript(server, config.OutputDir)
		if err != nil {
			log.Printf("Error running discovery script: %v", err)
			success = false
			message = fmt.Sprintf("Discovery failed: %v", err)
		} else {
			success = true
			message = "Discovery completed successfully"
		}

		// Update discovery record
		endTime := time.Now()
		result := DiscoveryResult{
			ID:         discoveryID,
			Server:     server.Hostname,
			Success:    success,
			Message:    message,
			StartTime:  startTime,
			EndTime:    endTime,
			OutputPath: outputPath,
			Error:      err.Error(),
			Region:     server.Region,
		}

		// Update discovery record
		updateDiscoveryStatus(discoveryID, success, message, outputPath)

		// Store result in database if enabled
		if config.DatabaseConfig.Enabled {
			err := storeResultInDatabase(result)
			if err != nil {
				log.Printf("Error storing result in database: %v", err)
			}
		}
	}()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Discovery started",
		"discovery_id": discoveryID,
	})
}

// getServerByID retrieves a server by its ID
func getServerByID(id int) (ServerConfig, error) {
	// In a real application, this would query the database
	// For now, we'll use mock data
	for _, server := range config.Servers {
		if server.ID == id {
			return server, nil
		}
	}
	return ServerConfig{}, fmt.Errorf("server with ID %d not found", id)
}

// createDiscoveryRecord creates a new discovery record in the database
func createDiscoveryRecord(serverID int) (int, error) {
	// In a real application, this would insert a record into the database
	// For now, we'll just return a mock ID
	return int(time.Now().Unix()), nil
}

// updateDiscoveryStatus updates the status of a discovery record
func updateDiscoveryStatus(id int, success bool, message, outputPath string) error {
	// In a real application, this would update a record in the database
	// For now, we'll just log the update
	log.Printf("Discovery %d: success=%v, message=%s, outputPath=%s", id, success, message, outputPath)
	return nil
}

// Load configuration from JSON file
func loadConfig() error {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if config.Concurrency <= 0 {
		config.Concurrency = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 600 // 10 minutes default timeout
	}
	if config.OutputDir == "" {
		config.OutputDir = "discovery_results"
	}
	if config.CacheTTL <= 0 {
		config.CacheTTL = 30 // 30 minutes default cache TTL
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 20
	}
	if config.PowerShellScript == "" {
		config.PowerShellScript = "Enhanced-ServerDiscovery.ps1"
	}
	if config.MetricsPort == 0 {
		config.MetricsPort = 9090
	}
	if config.TracingEndpoint == "" {
		config.TracingEndpoint = "localhost:4317" // Default OTLP gRPC endpoint
	}

	return nil
}

// Execute PowerShell script on a server
func executeScript(client *winrm.Client, hostname, scriptContent, outputDir string) (string, error) {
	// Create a unique output directory for this execution
	timestamp := time.Now().Format("20060102_150405")
	executionDir := filepath.Join(outputDir, fmt.Sprintf("%s_%s", hostname, timestamp))
	if err := os.MkdirAll(executionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create execution directory: %w", err)
	}

	// Create a temporary file for the script
	scriptFile := filepath.Join(executionDir, "discovery_script.ps1")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	// Prepare the command to execute the script
	// We'll use a base64 encoded command to avoid issues with special characters
	encodedScript := base64.StdEncoding.EncodeToString([]byte(scriptContent))
	command := fmt.Sprintf("powershell.exe -EncodedCommand %s", encodedScript)

	// Execute the command
	var stdout, stderr bytes.Buffer
	exitCode, err := runCommand(client, command, &stdout, &stderr)
	if err != nil {
		return "", fmt.Errorf("failed to execute script: %w", err)
	}

	// Check exit code
	if exitCode != 0 {
		return "", fmt.Errorf("script execution failed with exit code %d: %s", exitCode, stderr.String())
	}

	// Write output to files
	if err := os.WriteFile(filepath.Join(executionDir, "stdout.txt"), stdout.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write stdout file: %w", err)
	}
	if err := os.WriteFile(filepath.Join(executionDir, "stderr.txt"), stderr.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write stderr file: %w", err)
	}

	return executionDir, nil
}

// Run command on a server
func runCommand(client *winrm.Client, command string, stdout, stderr io.Writer) (int, error) {
	return client.Run(command, stdout, stderr)
}

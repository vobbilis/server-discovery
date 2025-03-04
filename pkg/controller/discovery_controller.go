package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/masterzen/winrm"
	"github.com/patrickmn/go-cache"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/vobbilis/codegen/server-discovery/pkg/database"
	"github.com/vobbilis/codegen/server-discovery/pkg/discovery"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// DiscoveryController handles server discovery operations
type DiscoveryController struct {
	config         models.Config
	connectionPool ConnectionPool
	discoveryCache *cache.Cache
	resultChannel  chan models.DiscoveryResult
	completedJobs  int32
	totalJobs      int32
	jobsMutex      sync.Mutex
	progressTicker *time.Ticker
	progressDone   chan bool
	db             *database.Database
	resourceCtrl   ResourceController
	workers        []*WorkerNode
}

// NewDiscoveryController creates a new discovery controller
func NewDiscoveryController(config *models.Config, db *database.Database) *DiscoveryController {
	return &DiscoveryController{
		config:         *config,
		db:             db,
		discoveryCache: cache.New(30*time.Minute, 10*time.Minute),
		resultChannel:  make(chan models.DiscoveryResult, 100),
		connectionPool: ConnectionPool{
			clients:     make(map[string]*winrm.Client),
			lastUsed:    make(map[string]time.Time),
			maxSize:     10,
			idleTimeout: 10 * time.Minute,
		},
		progressDone: make(chan bool),
	}
}

// ConnectionPool manages WinRM client connections
type ConnectionPool struct {
	clients     map[string]*winrm.Client
	mutex       sync.Mutex
	maxSize     int
	idleTimeout time.Duration
	lastUsed    map[string]time.Time
}

// ResourceController manages system resources
type ResourceController struct {
	CPUThreshold    float64
	MemoryThreshold float64
	lastCheck       time.Time
	checkInterval   time.Duration
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

// getClient creates a new WinRM client
func getClient(server models.ServerConfig) (*winrm.Client, error) {
	endpoint := winrm.NewEndpoint(
		server.Host,
		server.WinRMPort,
		server.WinRMHTTPS,
		server.WinRMInsecure,
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

// Load PowerShell script from file
func loadPowerShellScript(scriptPath string) (string, error) {
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("error reading PowerShell script: %w", err)
	}
	return string(scriptBytes), nil
}

// WindowsDiscoverer implements ServerDiscoverer for Windows servers
type WindowsDiscoverer struct {
	client        *winrm.Client
	scriptContent string
}

// LinuxDiscoverer implements ServerDiscoverer for Linux servers
type LinuxDiscoverer struct {
	sshConfig     models.SSHConfig
	scriptContent string
}

// ExecuteDiscovery for Windows servers
func (d *WindowsDiscoverer) ExecuteDiscovery(server models.ServerConfig, outputDir string) (models.DiscoveryResult, error) {
	result := models.DiscoveryResult{
		Status:      "running",
		LastChecked: time.Now(),
	}

	// Create output directory
	serverOutputDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", server.Host, time.Now().Format("20060102-150405")))
	if err := os.MkdirAll(serverOutputDir, 0755); err != nil {
		return result, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a temporary file for the script
	scriptFile := filepath.Join(serverOutputDir, "discovery_script.ps1")
	if err := os.WriteFile(scriptFile, []byte(d.scriptContent), 0644); err != nil {
		return result, fmt.Errorf("failed to write script file: %w", err)
	}

	// Execute script on server
	var outputBuffer, errorBuffer bytes.Buffer
	command := fmt.Sprintf("powershell.exe -EncodedCommand %s", base64.StdEncoding.EncodeToString([]byte(d.scriptContent)))

	exitCode, err := runCommand(d.client, command, &outputBuffer, &errorBuffer)
	if err != nil || exitCode != 0 {
		result.Status = "failed"
		result.Error = fmt.Sprintf("execution error (exit code %d): %v\n%s", exitCode, err, errorBuffer.String())
		return result, err
	}

	result.Status = "completed"
	return result, nil
}

// ExecuteDiscovery for Linux servers
func (d *LinuxDiscoverer) ExecuteDiscovery(server models.ServerConfig, outputDir string) (models.DiscoveryResult, error) {
	result := models.DiscoveryResult{
		Status:      "running",
		LastChecked: time.Now(),
	}

	// Execute Linux discovery
	_, err := discovery.RunLinuxDiscovery(d.sshConfig, outputDir)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Linux discovery failed: %v", err)
		return result, err
	}

	result.Status = "completed"
	return result, nil
}

// ParseDiscoveryOutput for Windows servers
func (d *WindowsDiscoverer) ParseDiscoveryOutput(outputPath string) (models.ServerDetails, error) {
	// Parse the JSON output file
	jsonFile := filepath.Join(outputPath, "server_details.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return models.ServerDetails{}, fmt.Errorf("failed to read JSON output: %w", err)
	}

	var details models.ServerDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return models.ServerDetails{}, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return details, nil
}

// ParseDiscoveryOutput for Linux servers
func (d *LinuxDiscoverer) ParseDiscoveryOutput(outputPath string) (models.ServerDetails, error) {
	// Parse the JSON output file
	jsonFile := filepath.Join(outputPath, "server_details.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return models.ServerDetails{}, fmt.Errorf("failed to read JSON output: %w", err)
	}

	var details models.ServerDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return models.ServerDetails{}, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return details, nil
}

// NewServerDiscoverer creates appropriate discoverer based on server type
func NewServerDiscoverer(server models.ServerConfig, scriptPath string) (discovery.ServerDiscoverer, error) {
	if server.UseWinRM {
		client, err := getClient(server)
		if err != nil {
			return nil, fmt.Errorf("failed to create WinRM client: %w", err)
		}
		scriptContent, err := loadPowerShellScript(scriptPath)
		if err != nil {
			return nil, err
		}
		return &WindowsDiscoverer{
			client:        client,
			scriptContent: scriptContent,
		}, nil
	}

	scriptContent, err := loadPowerShellScript(scriptPath)
	if err != nil {
		return nil, err
	}
	return &LinuxDiscoverer{
		sshConfig:     models.SSHConfig{},
		scriptContent: scriptContent,
	}, nil
}

// ExecuteDiscovery executes discovery on a server
func (c *DiscoveryController) ExecuteDiscovery(server models.ServerConfig, scriptContent string) models.DiscoveryResult {
	serverKey := fmt.Sprintf("%s:%d", server.Host, server.WinRMPort)

	// Check cache first
	if cachedResult, found := c.discoveryCache.Get(serverKey); found {
		log.Printf("Using cached result for %s", serverKey)
		result := cachedResult.(models.DiscoveryResult)
		result.Message = "Retrieved from cache"
		return result
	}

	// Create appropriate discoverer
	discoverer, err := NewServerDiscoverer(server, c.config.PowerShellScript)
	if err != nil {
		return models.DiscoveryResult{
			Server:    serverKey,
			Success:   false,
			Error:     fmt.Sprintf("Failed to create discoverer: %v", err),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
	}

	// Execute discovery
	result, err := discoverer.ExecuteDiscovery(server, c.config.OutputDir)
	if err != nil {
		log.Printf("Discovery failed for %s: %v", serverKey, err)
	} else {
		// Cache successful results
		c.discoveryCache.Set(serverKey, result, cache.DefaultExpiration)
	}

	return result
}

// Run command on a server
func runCommand(client *winrm.Client, command string, stdout, stderr io.Writer) (int, error) {
	return client.Run(command, stdout, stderr)
}

// StoreResultInDatabase stores a discovery result in the database
func (c *DiscoveryController) StoreResultInDatabase(result models.DiscoveryResult) error {
	// Create discovery result in database
	_, err := c.db.CreateDiscoveryResult(result)
	if err != nil {
		return fmt.Errorf("failed to store discovery result: %w", err)
	}
	return nil
}

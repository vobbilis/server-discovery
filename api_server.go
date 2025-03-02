package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// API Server instance
type APIServer struct {
	config APIServerConfig
	router *mux.Router
}

// Initialize API server
func NewAPIServer(config APIServerConfig) *APIServer {
	server := &APIServer{
		config: config,
		router: mux.NewRouter(),
	}

	// Register routes
	server.registerRoutes()

	return server
}

// Register API routes
func (s *APIServer) registerRoutes() {
	// API routes
	s.router.HandleFunc("/api/servers", getServersHandler).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}", getServerDetailsHandler).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/discoveries", getServerDiscoveriesHandler).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/discover", runServerDiscoveryHandler).Methods("POST")
	s.router.HandleFunc("/api/discoveries/{id}", getDiscoveryDetailsHandler).Methods("GET")
	s.router.HandleFunc("/api/query", executeQueryHandler).Methods("POST")
	s.router.HandleFunc("/api/stats", getStatsHandler).Methods("GET")
}

// Start the API server
func startAPIServer() {
	// Default configuration if not specified in config
	if config.APIServer.Port == 0 {
		config.APIServer.Port = 8080
	}
	if config.APIServer.AllowedOrigins == "" {
		config.APIServer.AllowedOrigins = "http://localhost:3000"
	}
	if config.APIServer.ReadTimeout == 0 {
		config.APIServer.ReadTimeout = 15
	}
	if config.APIServer.WriteTimeout == 0 {
		config.APIServer.WriteTimeout = 15
	}
	if config.APIServer.ShutdownTimeout == 0 {
		config.APIServer.ShutdownTimeout = 15
	}

	// Create router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/servers", getServersHandler).Methods("GET")
	router.HandleFunc("/api/servers/{id}", getServerDetailsHandler).Methods("GET")
	router.HandleFunc("/api/servers/{id}/discoveries", getServerDiscoveriesHandler).Methods("GET")
	router.HandleFunc("/api/servers/{id}/discover", runServerDiscoveryHandler).Methods("POST")
	router.HandleFunc("/api/discoveries/{id}", getDiscoveryDetailsHandler).Methods("GET")
	router.HandleFunc("/api/query", executeQueryHandler).Methods("POST")
	router.HandleFunc("/api/stats", getStatsHandler).Methods("GET")

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(config.APIServer.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.APIServer.Port),
		Handler:      c.Handler(router),
		ReadTimeout:  time.Duration(config.APIServer.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.APIServer.WriteTimeout) * time.Second,
	}

	// Start server
	log.Printf("API server listening on port %d", config.APIServer.Port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server error: %v", err)
		}
	}()
}

// Handler for getting system stats
func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Use mock stats instead of database stats
	stats := getMockStats()

	// Add servers to the stats
	if _, ok := stats["servers"]; !ok {
		servers, err := getAllServersWithDetails()
		if err != nil {
			log.Printf("Error getting servers: %v", err)
		} else {
			stats["servers"] = servers
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Basic validation for SQL queries
func isValidQuery(query string) bool {
	// Convert to uppercase for case-insensitive comparison
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Only allow SELECT statements
	if !strings.HasPrefix(upperQuery, "SELECT") {
		return false
	}

	// Disallow potentially harmful statements
	disallowedKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE",
		"GRANT", "REVOKE", "EXECUTE", "EXEC", "CALL",
	}

	for _, keyword := range disallowedKeywords {
		if strings.Contains(upperQuery, keyword) {
			return false
		}
	}

	return true
}

// Execute a custom SQL query
func executeCustomQuery(query string) ([]map[string]interface{}, error) {
	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare result
	var results []map[string]interface{}

	// Scan rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the slice
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle null values
			if val == nil {
				row[col] = nil
				continue
			}

			// Convert bytes to string for JSON encoding
			switch v := val.(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}

		results = append(results, row)
	}

	return results, nil
}

// Get system stats
func getSystemStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get server count
	var serverCount int
	err := db.QueryRow("SELECT COUNT(*) FROM server_discovery.servers").Scan(&serverCount)
	if err != nil {
		return nil, err
	}
	stats["serverCount"] = serverCount

	// Get discovery count
	var discoveryCount int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.discovery_results").Scan(&discoveryCount)
	if err != nil {
		return nil, err
	}
	stats["discoveryCount"] = discoveryCount

	// Get success rate
	var successCount int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.discovery_results WHERE success = true").Scan(&successCount)
	if err != nil {
		return nil, err
	}

	var successRate float64
	if discoveryCount > 0 {
		successRate = float64(successCount) / float64(discoveryCount) * 100
	}
	stats["successRate"] = successRate

	// Get regions
	rows, err := db.Query("SELECT region, COUNT(*) FROM server_discovery.servers GROUP BY region")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make(map[string]int)
	for rows.Next() {
		var region string
		var count int
		if err := rows.Scan(&region, &count); err != nil {
			return nil, err
		}
		regions[region] = count
	}
	stats["regions"] = regions

	return stats, nil
}

// Handler for running a discovery on a server
func runServerDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Starting server discovery for server ID: %s", r.URL.Path)

	// Extract server ID from the URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[ERROR] Invalid server ID: %s - %v", idStr, err)
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] Running discovery for server ID: %d", id)

	// Get the server details
	servers := getMockServers()
	var targetServer *Server
	for _, server := range servers {
		if server.ID == id {
			s := server
			targetServer = &s
			break
		}
	}

	if targetServer == nil {
		log.Printf("[ERROR] Server not found: ID %d", id)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	log.Printf("[DEBUG] Found server: %s (Port: %d, Region: %s)",
		targetServer.Hostname, targetServer.Port, targetServer.Region)

	// In a real implementation, we would trigger an actual discovery
	// For now, we'll just create a mock discovery result
	discoveryID := len(getMockServerDiscoveries(id)) + 1
	log.Printf("[DEBUG] Created new discovery ID: %d for server ID: %d", discoveryID, id)

	// Set OS name and version based on server type
	var osName, osVersion string
	if targetServer.Port == 22 {
		osName = "Ubuntu 20.04.4 LTS"
		osVersion = "20.04"
	} else {
		osName = "Windows Server 2019"
		osVersion = "10.0.17763"
	}

	// Create a new discovery result
	discovery := DiscoveryDetails{
		ID:             discoveryID,
		ServerID:       id,
		ServerHostname: targetServer.Hostname,
		ServerPort:     targetServer.Port,
		ServerRegion:   targetServer.Region,
		Success:        true,
		Message:        "Discovery completed successfully",
		StartTime:      time.Now().Add(-5 * time.Minute),
		EndTime:        time.Now(),
		OSName:         osName,
		OSVersion:      osVersion,
		CPUModel:       "Intel(R) Xeon(R) CPU E5-2670 0 @ 2.60GHz",
		CPUCount:       4,
		MemoryTotalGB:  16.0,
		DiskTotalGB:    256.0,
		DiskFreeGB:     128.0,
		LastBootTime:   time.Now().Add(-7 * 24 * time.Hour),
	}
	log.Printf("[DEBUG] Created discovery result for server %s: OS: %s, CPU: %s, Memory: %.2f GB",
		targetServer.Hostname, osName, discovery.CPUModel, discovery.MemoryTotalGB)

	// In a real implementation, we would save this to the database
	// For now, we'll just return it

	// Update the server's last discovery time
	targetServer.LastDiscovery = time.Now()
	targetServer.DiscoveryCount++
	log.Printf("[DEBUG] Updated server %s: LastDiscovery: %s, DiscoveryCount: %d",
		targetServer.Hostname, targetServer.LastDiscovery.Format(time.RFC3339), targetServer.DiscoveryCount)

	// Return the discovery result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
	log.Printf("[DEBUG] Completed discovery for server ID: %d", id)
}

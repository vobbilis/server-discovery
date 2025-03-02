package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Mock implementation for the servers endpoint
func getServersHandler(w http.ResponseWriter, r *http.Request) {
	mockServers := getMockServers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockServers)
}

// Mock implementation for the server details endpoint
func getServerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Getting server details for: %s", r.URL.Path)

	// Extract server ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	serverID := 1
	fmt.Sscanf(id, "%d", &serverID)
	log.Printf("[DEBUG] Parsed server ID: %d", serverID)

	mockServer := getMockServerWithDetails(serverID)
	log.Printf("[DEBUG] Retrieved server details for ID %d: Hostname: %s, OS: %s",
		serverID, mockServer.Hostname, mockServer.OSName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockServer)
	log.Printf("[DEBUG] Sent server details response for ID: %d", serverID)
}

// Mock implementation for the server discoveries endpoint
func getServerDiscoveriesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Getting server discoveries for: %s", r.URL.Path)

	// Extract server ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]

	serverID := 1
	fmt.Sscanf(id, "%d", &serverID)
	log.Printf("[DEBUG] Parsed server ID: %d", serverID)

	mockDiscoveries := getMockServerDiscoveries(serverID)
	log.Printf("[DEBUG] Retrieved %d discoveries for server ID %d", len(mockDiscoveries), serverID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockDiscoveries)
	log.Printf("[DEBUG] Sent server discoveries response for ID: %d", serverID)
}

// Mock implementation for the discovery details endpoint
func getDiscoveryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Getting discovery details for: %s", r.URL.Path)

	vars := mux.Vars(r)
	id := vars["id"]

	var discoveryID int
	fmt.Sscanf(id, "%d", &discoveryID)
	log.Printf("[DEBUG] Parsed discovery ID: %d", discoveryID)

	mockDiscovery := getMockDiscoveryDetails(discoveryID)
	log.Printf("[DEBUG] Retrieved discovery details for ID %d: Server: %s, Success: %t",
		discoveryID, mockDiscovery.ServerHostname, mockDiscovery.Success)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockDiscovery)
	log.Printf("[DEBUG] Sent discovery details response for ID: %d", discoveryID)
}

// Mock implementation for the query endpoint
func executeQueryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var request struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// Create mock query results
	mockResults := []map[string]interface{}{
		{
			"id":       1,
			"hostname": "win-server-01",
			"port":     5985,
			"region":   "us-east",
		},
		{
			"id":       2,
			"hostname": "win-server-02",
			"port":     5985,
			"region":   "us-west",
		},
		{
			"id":       3,
			"hostname": "win-server-03",
			"port":     5985,
			"region":   "eu-central",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockResults)
}

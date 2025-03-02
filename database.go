package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Database connection
var db *sql.DB

// Initialize database connection
func initDatabase() error {
	if !config.DatabaseConfig.Enabled {
		log.Println("Database is disabled in configuration")
		return nil
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseConfig.Host,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.Username,
		config.DatabaseConfig.Password,
		config.DatabaseConfig.Database)

	var err error
	db, err = sql.Open(config.DatabaseConfig.Type, connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

// Close database connection
func closeDatabase() {
	if db != nil {
		db.Close()
		log.Println("Database connection closed")
	}
}

// Store discovery result in database
func storeResultInDatabase(result DiscoveryResult) error {
	if db == nil || !config.DatabaseConfig.Enabled {
		return nil // Database not enabled, silently ignore
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert discovery result
	_, err = tx.Exec(`
		INSERT INTO server_discovery.discovery_results
		(server_id, success, message, start_time, end_time, output_path, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, result.ID, result.Success, result.Message, result.StartTime, result.EndTime, result.OutputPath, result.Error)
	if err != nil {
		return fmt.Errorf("failed to insert discovery result: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Stored discovery result for %s in database", result.Server)
	return nil
}

// Store server details from discovery output files
func storeServerDetails(tx *sql.Tx, serverID, discoveryID int, outputPath string) error {
	// Parse JSON data from the results file
	serverData, err := parseServerDetailsFromOutput(outputPath)
	if err != nil {
		return fmt.Errorf("failed to parse server details: %w", err)
	}

	// Convert complex data to JSON
	ipAddressesJSON, err := json.Marshal(serverData.IPAddresses)
	if err != nil {
		return err
	}

	installedSoftwareJSON, err := json.Marshal(serverData.InstalledSoftware)
	if err != nil {
		return err
	}

	runningServicesJSON, err := json.Marshal(serverData.RunningServices)
	if err != nil {
		return err
	}

	openPortsJSON, err := json.Marshal(serverData.OpenPorts)
	if err != nil {
		return err
	}

	// Insert server details
	_, err = tx.Exec(`
		INSERT INTO server_discovery.server_details
		(server_id, discovery_id, os_name, os_version, cpu_model, cpu_count,
		memory_total_gb, disk_total_gb, disk_free_gb, last_boot_time,
		ip_addresses, installed_software, running_services, open_ports)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, serverID, discoveryID, serverData.OSName, serverData.OSVersion,
		serverData.CPUModel, serverData.CPUCount, serverData.MemoryTotalGB,
		serverData.DiskTotalGB, serverData.DiskFreeGB, serverData.LastBootTime,
		ipAddressesJSON, installedSoftwareJSON, runningServicesJSON, openPortsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert server details: %w", err)
	}

	// Store server tags
	for _, tag := range serverData.Tags {
		_, err = tx.Exec(`
			INSERT INTO server_discovery.server_tags (server_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (server_id, key) DO UPDATE
			SET value = $3
		`, serverID, tag.Key, tag.Value)
		if err != nil {
			return fmt.Errorf("failed to insert server tag: %w", err)
		}
	}

	return nil
}

// Get all servers from database
func getAllServers() ([]ServerWithDetails, error) {
	if db == nil || !config.DatabaseConfig.Enabled {
		return nil, fmt.Errorf("database not enabled or initialized")
	}

	// Query for servers with their latest details
	rows, err := db.Query(`
		WITH latest_discovery AS (
			SELECT DISTINCT ON (server_id) 
				server_id, id AS discovery_id
			FROM server_discovery.discovery_results
			WHERE success = true
			ORDER BY server_id, end_time DESC
		)
		SELECT 
			s.id, s.hostname, s.port, s.region, 
			sd.os_name, sd.os_version, sd.cpu_model, sd.cpu_count,
			sd.memory_total_gb, sd.disk_total_gb, sd.disk_free_gb,
			sd.last_boot_time, sd.ip_addresses, sd.installed_software,
			sd.running_services, sd.open_ports
		FROM server_discovery.servers s
		LEFT JOIN latest_discovery ld ON s.id = ld.server_id
		LEFT JOIN server_discovery.server_details sd ON ld.discovery_id = sd.discovery_id
		ORDER BY s.hostname
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying servers: %w", err)
	}
	defer rows.Close()

	var servers []ServerWithDetails
	for rows.Next() {
		var server ServerWithDetails
		var ipAddressesJSON, softwareJSON, servicesJSON, portsJSON []byte

		err := rows.Scan(
			&server.ID, &server.Hostname, &server.Port, &server.Region,
			&server.OSName, &server.OSVersion, &server.CPUModel, &server.CPUCount,
			&server.MemoryTotalGB, &server.DiskTotalGB, &server.DiskFreeGB,
			&server.LastBootTime, &ipAddressesJSON, &softwareJSON,
			&servicesJSON, &portsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning server row: %w", err)
		}

		// Parse JSON fields
		if ipAddressesJSON != nil {
			json.Unmarshal(ipAddressesJSON, &server.IPAddresses)
		}
		if softwareJSON != nil {
			json.Unmarshal(softwareJSON, &server.InstalledSoftware)
		}
		if servicesJSON != nil {
			json.Unmarshal(servicesJSON, &server.RunningServices)
		}
		if portsJSON != nil {
			json.Unmarshal(portsJSON, &server.OpenPorts)
		}

		// Get tags for this server
		server.Tags, err = getServerTags(server.ID)
		if err != nil {
			log.Printf("Error getting tags for server %d: %v", server.ID, err)
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// Get tags for a server
func getServerTags(serverID int) ([]Tag, error) {
	rows, err := db.Query(`
		SELECT key, value
		FROM server_discovery.server_tags
		WHERE server_id = $1
	`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.Key, &tag.Value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Insert open ports for a discovery
func insertOpenPorts(tx *sql.Tx, discoveryID int, ports []Port) error {
	for _, port := range ports {
		_, err := tx.Exec(`
			INSERT INTO server_discovery.open_ports (
				discovery_id, local_port, local_ip, remote_port, remote_ip, 
				state, description, process_id, process_name
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, discoveryID, port.LocalPort, port.LocalIP, port.RemotePort, port.RemoteIP,
			port.State, port.Description, port.ProcessID, port.ProcessName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get open ports for a discovery
func getOpenPorts(discoveryID int) ([]Port, error) {
	rows, err := db.Query(`
		SELECT local_port, local_ip, remote_port, remote_ip, state, description, process_id, process_name
		FROM server_discovery.open_ports
		WHERE discovery_id = $1
	`, discoveryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []Port
	for rows.Next() {
		var port Port
		var localIP, remoteIP, description, processName sql.NullString
		var remotePort, processID sql.NullInt64
		err := rows.Scan(
			&port.LocalPort, &localIP, &remotePort, &remoteIP,
			&port.State, &description, &processID, &processName,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if localIP.Valid {
			port.LocalIP = localIP.String
		}
		if remotePort.Valid {
			port.RemotePort = int(remotePort.Int64)
		}
		if remoteIP.Valid {
			port.RemoteIP = remoteIP.String
		}
		if description.Valid {
			port.Description = description.String
		}
		if processID.Valid {
			port.ProcessID = int(processID.Int64)
		}
		if processName.Valid {
			port.ProcessName = processName.String
		}

		ports = append(ports, port)
	}

	return ports, nil
}

// Parse server details from discovery output files
func parseServerDetailsFromOutput(outputPath string) (ServerDetails, error) {
	// In a real implementation, this would parse JSON files from the output directory
	// For now, return mock data
	return getMockServerDetails(), nil
}

// Get mock server details
func getMockServerDetails() ServerDetails {
	return ServerDetails{
		OSName:        "Windows Server 2019",
		OSVersion:     "10.0.17763",
		CPUModel:      "Intel(R) Xeon(R) CPU E5-2670 0 @ 2.60GHz",
		CPUCount:      4,
		MemoryTotalGB: 16.0,
		DiskTotalGB:   256.0,
		DiskFreeGB:    128.0,
		LastBootTime:  time.Now().Add(-7 * 24 * time.Hour),
		IPAddresses: []IPAddress{
			{
				IPAddress:     "192.168.1.100",
				InterfaceName: "Ethernet",
			},
			{
				IPAddress:     "10.0.0.100",
				InterfaceName: "Internal",
			},
		},
		InstalledSoftware: []Software{
			{
				Name:        "Microsoft SQL Server 2019",
				Version:     "15.0.2000.5",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2023-01-15",
			},
			{
				Name:        "Microsoft .NET Framework 4.8",
				Version:     "4.8.03761",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2023-01-15",
			},
		},
		RunningServices: []Service{
			{
				Name:        "MSSQLSERVER",
				DisplayName: "SQL Server (MSSQLSERVER)",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT Service\\MSSQLSERVER",
			},
			{
				Name:        "SQLTELEMETRY",
				DisplayName: "SQL Server CEIP service (MSSQLSERVER)",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT Service\\SQLTELEMETRY",
			},
		},
		OpenPorts: []Port{
			{
				LocalPort:   1433,
				State:       "LISTENING",
				Description: "Microsoft SQL Server",
			},
			{
				LocalPort:   3389,
				State:       "LISTENING",
				Description: "Remote Desktop",
			},
		},
		Tags: []Tag{
			{
				Key:   "env",
				Value: "production",
			},
			{
				Key:   "role",
				Value: "database",
			},
		},
	}
}

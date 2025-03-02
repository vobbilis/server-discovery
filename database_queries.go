package main

import (
	"database/sql"
	"encoding/json"
)

// Get all servers
func getAllServersWithDetails() ([]Server, error) {
	query := `
		SELECT s.id, s.hostname, s.port, s.region, 
		       COUNT(DISTINCT dr.id) as discovery_count,
		       MAX(dr.end_time) as last_discovery
		FROM server_discovery.servers s
		LEFT JOIN server_discovery.discovery_results dr ON s.id = dr.server_id
		GROUP BY s.id, s.hostname, s.port, s.region
		ORDER BY s.hostname
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var server Server
		var lastDiscovery sql.NullTime
		var discoveryCount int

		err := rows.Scan(
			&server.ID,
			&server.Hostname,
			&server.Port,
			&server.Region,
			&discoveryCount,
			&lastDiscovery,
		)
		if err != nil {
			return nil, err
		}

		server.DiscoveryCount = discoveryCount
		if lastDiscovery.Valid {
			server.LastDiscovery = lastDiscovery.Time
		}

		// Get tags for this server
		tags, err := getServerTags(server.ID)
		if err != nil {
			return nil, err
		}
		server.Tags = tags

		servers = append(servers, server)
	}

	return servers, nil
}

// Get server with details
func getServerWithDetails(serverID int) (*ServerWithDetails, error) {
	// Get server info
	query := `
		SELECT s.id, s.hostname, s.port, s.region
		FROM server_discovery.servers s
		WHERE s.id = $1
	`

	var server ServerWithDetails
	err := db.QueryRow(query, serverID).Scan(
		&server.ID,
		&server.Hostname,
		&server.Port,
		&server.Region,
	)
	if err != nil {
		return nil, err
	}

	// Get tags
	tags, err := getServerTags(serverID)
	if err != nil {
		return nil, err
	}
	server.Tags = tags

	// Get latest details
	query = `
		SELECT sd.os_name, sd.os_version, sd.cpu_model, sd.cpu_count,
		       sd.memory_total_gb, sd.disk_total_gb, sd.disk_free_gb,
		       sd.last_boot_time, sd.ip_addresses, sd.installed_software,
		       sd.running_services, sd.open_ports
		FROM server_discovery.server_details sd
		JOIN server_discovery.discovery_results dr ON sd.discovery_id = dr.id
		WHERE sd.server_id = $1
		ORDER BY dr.end_time DESC
		LIMIT 1
	`

	var ipAddressesJSON, softwareJSON, servicesJSON, portsJSON sql.NullString

	err = db.QueryRow(query, serverID).Scan(
		&server.OSName,
		&server.OSVersion,
		&server.CPUModel,
		&server.CPUCount,
		&server.MemoryTotalGB,
		&server.DiskTotalGB,
		&server.DiskFreeGB,
		&server.LastBootTime,
		&ipAddressesJSON,
		&softwareJSON,
		&servicesJSON,
		&portsJSON,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Parse JSON fields
	if ipAddressesJSON.Valid {
		err = json.Unmarshal([]byte(ipAddressesJSON.String), &server.IPAddresses)
		if err != nil {
			return nil, err
		}
	}

	if softwareJSON.Valid {
		err = json.Unmarshal([]byte(softwareJSON.String), &server.InstalledSoftware)
		if err != nil {
			return nil, err
		}
	}

	if servicesJSON.Valid {
		err = json.Unmarshal([]byte(servicesJSON.String), &server.RunningServices)
		if err != nil {
			return nil, err
		}
	}

	if portsJSON.Valid {
		err = json.Unmarshal([]byte(portsJSON.String), &server.OpenPorts)
		if err != nil {
			return nil, err
		}
	}

	return &server, nil
}

// Get server discoveries
func getServerDiscoveries(serverID int) ([]DiscoveryResult, error) {
	query := `
		SELECT id, success, message, error_message, start_time, end_time, output_path
		FROM server_discovery.discovery_results
		WHERE server_id = $1
		ORDER BY end_time DESC
	`

	rows, err := db.Query(query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discoveries []DiscoveryResult
	for rows.Next() {
		var discovery DiscoveryResult
		var errorMsg sql.NullString
		var outputPath sql.NullString

		err := rows.Scan(
			&discovery.ID,
			&discovery.Success,
			&discovery.Message,
			&errorMsg,
			&discovery.StartTime,
			&discovery.EndTime,
			&outputPath,
		)
		if err != nil {
			return nil, err
		}

		if errorMsg.Valid {
			discovery.Error = errorMsg.String
		}

		if outputPath.Valid {
			discovery.OutputPath = outputPath.String
		}

		discoveries = append(discoveries, discovery)
	}

	return discoveries, nil
}

// Get discovery details
func getDiscoveryDetails(discoveryID int) (*DiscoveryDetails, error) {
	// Get discovery info
	query := `
		SELECT dr.id, dr.server_id, dr.success, dr.message, dr.error_message, 
		       dr.start_time, dr.end_time, dr.output_path,
		       s.hostname, s.port, s.region
		FROM server_discovery.discovery_results dr
		JOIN server_discovery.servers s ON dr.server_id = s.id
		WHERE dr.id = $1
	`

	var details DiscoveryDetails
	var errorMsg sql.NullString
	var outputPath sql.NullString

	err := db.QueryRow(query, discoveryID).Scan(
		&details.ID,
		&details.ServerID,
		&details.Success,
		&details.Message,
		&errorMsg,
		&details.StartTime,
		&details.EndTime,
		&outputPath,
		&details.ServerHostname,
		&details.ServerPort,
		&details.ServerRegion,
	)
	if err != nil {
		return nil, err
	}

	if errorMsg.Valid {
		details.Error = errorMsg.String
	}

	if outputPath.Valid {
		details.OutputPath = outputPath.String
	}

	// Get server details for this discovery
	if details.Success {
		query = `
			SELECT sd.os_name, sd.os_version, sd.cpu_model, sd.cpu_count,
			       sd.memory_total_gb, sd.disk_total_gb, sd.disk_free_gb,
			       sd.last_boot_time, sd.ip_addresses, sd.installed_software,
			       sd.running_services, sd.open_ports
			FROM server_discovery.server_details sd
			WHERE sd.discovery_id = $1
		`

		var ipAddressesJSON, softwareJSON, servicesJSON, portsJSON sql.NullString

		err = db.QueryRow(query, discoveryID).Scan(
			&details.OSName,
			&details.OSVersion,
			&details.CPUModel,
			&details.CPUCount,
			&details.MemoryTotalGB,
			&details.DiskTotalGB,
			&details.DiskFreeGB,
			&details.LastBootTime,
			&ipAddressesJSON,
			&softwareJSON,
			&servicesJSON,
			&portsJSON,
		)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		// Parse JSON fields
		if ipAddressesJSON.Valid {
			err = json.Unmarshal([]byte(ipAddressesJSON.String), &details.IPAddresses)
			if err != nil {
				return nil, err
			}
		}

		if softwareJSON.Valid {
			err = json.Unmarshal([]byte(softwareJSON.String), &details.InstalledSoftware)
			if err != nil {
				return nil, err
			}
		}

		if servicesJSON.Valid {
			err = json.Unmarshal([]byte(servicesJSON.String), &details.RunningServices)
			if err != nil {
				return nil, err
			}
		}

		if portsJSON.Valid {
			err = json.Unmarshal([]byte(portsJSON.String), &details.OpenPorts)
			if err != nil {
				return nil, err
			}
		}
	}

	return &details, nil
}

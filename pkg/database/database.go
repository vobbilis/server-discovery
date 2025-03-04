package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// Database represents a connection to the PostgreSQL database
type Database struct {
	db *sqlx.DB
}

// NewDatabase creates a new database connection
func NewDatabase(config *models.DatabaseConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// GetAllServers retrieves all servers from the database
func (d *Database) GetAllServers() ([]models.ServerWithDetails, error) {
	rows, err := d.db.Queryx(`
		SELECT 
			s.id, 
			s.hostname,
			s.ip,
			s.os_type,
			s.region,
			s.status,
			s.last_checked,
			COALESCE(m.cpu_usage, 0) as cpu_usage,
			COALESCE(m.memory_total, 0) as memory_total,
			COALESCE(m.memory_used, 0) as memory_used,
			COALESCE(m.disk_total, 0) as disk_total,
			COALESCE(m.disk_used, 0) as disk_used,
			COALESCE(m.load_average, 0) as load_average,
			COALESCE(m.process_count, 0) as process_count
		FROM server_discovery.servers s
		LEFT JOIN server_discovery.server_metrics m ON s.id = m.server_id
		ORDER BY s.hostname
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying servers: %w", err)
	}
	defer rows.Close()

	var servers []models.ServerWithDetails
	for rows.Next() {
		var server models.ServerWithDetails
		var metrics models.ServerMetrics
		err := rows.Scan(
			&server.ID,
			&server.Hostname,
			&server.IP,
			&server.OSType,
			&server.Region,
			&server.Status,
			&server.LastChecked,
			&metrics.CPUUsage,
			&metrics.MemoryTotal,
			&metrics.MemoryUsed,
			&metrics.DiskTotal,
			&metrics.DiskUsed,
			&metrics.LoadAverage,
			&metrics.ProcessCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning server row: %w", err)
		}

		server.Metrics = &metrics

		// Get tags for this server
		tags, err := d.GetServerTags(server.ID)
		if err != nil {
			log.Printf("Warning: Error getting tags for server %d: %v", server.ID, err)
		} else {
			server.Tags = tags
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// GetServerDetails retrieves detailed information about a specific server
func (d *Database) GetServerDetails(serverID string) (*models.ServerDetails, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			s.id,
			s.hostname,
			s.ip,
			s.os_type,
			s.status,
			s.last_checked,
			s.region,
			sd.os_name,
			sd.os_version,
			sd.cpu_model,
			sd.cpu_count,
			sd.memory_total_gb,
			sd.disk_total_gb,
			sd.disk_free_gb,
			sd.last_boot_time
		FROM server_discovery.servers s
		LEFT JOIN server_discovery.server_details sd ON s.id = sd.server_id
		WHERE s.id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server details: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("server not found")
	}

	var details models.ServerDetails
	err = rows.Scan(
		&details.ID,
		&details.Hostname,
		&details.IP,
		&details.OSType,
		&details.Status,
		&details.LastChecked,
		&details.Region,
		&details.OSName,
		&details.OSVersion,
		&details.CPUModel,
		&details.CPUCount,
		&details.MemoryTotalGB,
		&details.DiskTotalGB,
		&details.DiskFreeGB,
		&details.LastBootTime,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning server details: %v", err)
	}

	// Get server metrics
	metrics, err := d.getServerMetrics(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server metrics: %v", err)
	} else {
		details.Metrics = metrics
	}

	// Get server services
	services, err := d.getServerServices(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server services: %v", err)
	} else {
		details.Services = services
	}

	// Get server IP addresses
	ipAddresses, err := d.GetServerIPAddresses(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server IP addresses: %v", err)
	} else {
		details.IPAddresses = ipAddresses
	}

	// Get server open ports
	openPorts, err := d.GetServerOpenPorts(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server open ports: %v", err)
	} else {
		details.OpenPorts = openPorts
	}

	// Get server installed software
	software, err := d.GetServerInstalledSoftware(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server installed software: %v", err)
	} else {
		details.InstalledSoftware = software
	}

	// Get server filesystems
	filesystems, err := d.GetServerFilesystems(serverID)
	if err != nil {
		log.Printf("Warning: Error getting server filesystems: %v", err)
	} else {
		details.Filesystems = filesystems
	}

	// Get server tags
	tags, err := d.GetServerTags(id)
	if err != nil {
		log.Printf("Warning: Error getting server tags: %v", err)
	} else {
		details.Tags = tags
	}

	return &details, nil
}

// GetServerDiscoveries retrieves discovery history for a specific server
func (d *Database) GetServerDiscoveries(serverID string) ([]models.DiscoveryResult, error) {
	rows, err := d.db.Queryx(`
		SELECT id, server_id, success, message, start_time, end_time, status
		FROM server_discovery.discovery_results
		WHERE server_id = $1
		ORDER BY end_time DESC
	`, serverID)
	if err != nil {
		return nil, fmt.Errorf("error querying discoveries: %v", err)
	}
	defer rows.Close()

	var discoveries []models.DiscoveryResult
	for rows.Next() {
		var d models.DiscoveryResult
		err := rows.Scan(
			&d.ID,
			&d.ServerID,
			&d.Success,
			&d.Message,
			&d.StartTime,
			&d.EndTime,
			&d.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning discovery row: %v", err)
		}
		discoveries = append(discoveries, d)
	}

	return discoveries, nil
}

// CreateDiscoveryResult creates a new discovery result in the database
func (d *Database) CreateDiscoveryResult(result models.DiscoveryResult) (int, error) {
	var id int
	log.Printf("[DEBUG] Creating discovery result: %+v", result)
	err := d.db.QueryRowx(`
		INSERT INTO server_discovery.discovery_results (
			server_id, success, message, start_time, end_time, output_path, error, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, result.ServerID, result.Success, result.Message, result.StartTime,
		result.EndTime, result.OutputPath, result.Error, result.Status).Scan(&id)
	if err != nil {
		log.Printf("[ERROR] Failed to create discovery result: %v", err)
		return 0, fmt.Errorf("failed to create discovery result: %w", err)
	}
	log.Printf("[DEBUG] Created discovery result with ID: %d", id)
	return id, nil
}

// GetAllDiscoveries retrieves all discovery results from the database
func (d *Database) GetAllDiscoveries() ([]models.DiscoveryResult, error) {
	rows, err := d.db.Queryx(`
		SELECT id, server_id, success, message, start_time, end_time, output_path, error, status
		FROM server_discovery.discovery_results
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying discovery results: %w", err)
	}
	defer rows.Close()

	var results []models.DiscoveryResult
	for rows.Next() {
		var d models.DiscoveryResult
		var outputPath, errorMsg sql.NullString
		err := rows.Scan(
			&d.ID,
			&d.ServerID,
			&d.Success,
			&d.Message,
			&d.StartTime,
			&d.EndTime,
			&outputPath,
			&errorMsg,
			&d.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning discovery result row: %w", err)
		}

		if outputPath.Valid {
			d.OutputPath = outputPath.String
		}
		if errorMsg.Valid {
			d.Error = errorMsg.String
		}

		results = append(results, d)
	}

	return results, nil
}

// GetDiscoveryByID retrieves a single discovery result by its ID
func (d *Database) GetDiscoveryByID(id int) (*models.DiscoveryResult, error) {
	var result models.DiscoveryResult
	err := d.db.QueryRowx(`
		SELECT id, server_id, success, message, start_time, end_time, output_path, error, status
		FROM server_discovery.discovery_results
		WHERE id = $1
	`, id).StructScan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error querying discovery result: %w", err)
	}

	return &result, nil
}

// Helper functions

func (d *Database) getServerMetrics(serverID string) (*models.ServerMetrics, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	var metrics models.ServerMetrics
	err = d.db.QueryRowx(`
		SELECT 
			cpu_usage,
			memory_total,
			memory_used,
			disk_total,
			disk_used,
			load_average,
			process_count
		FROM server_discovery.server_metrics
		WHERE server_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, id).Scan(
		&metrics.CPUUsage,
		&metrics.MemoryTotal,
		&metrics.MemoryUsed,
		&metrics.DiskTotal,
		&metrics.DiskUsed,
		&metrics.LoadAverage,
		&metrics.ProcessCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting server metrics: %v", err)
	}
	return &metrics, nil
}

func (d *Database) getServerServices(serverID string) ([]models.Service, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			id,
			server_id,
			service_name,
			service_status,
			service_description,
			port,
			created_at,
			updated_at
		FROM server_discovery.server_services
		WHERE server_id = $1
		ORDER BY service_name
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server services: %v", err)
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		err := rows.Scan(
			&service.ID,
			&service.ServerID,
			&service.Name,
			&service.Status,
			&service.Description,
			&service.Port,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning service row: %v", err)
		}
		services = append(services, service)
	}
	return services, nil
}

func (d *Database) GetServerIPAddresses(serverID string) ([]models.IPAddress, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			ip_address,
			interface_name
		FROM server_discovery.ip_addresses
		WHERE discovery_id IN (
			SELECT id FROM server_discovery.discovery_results
			WHERE server_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY ip_address
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server IP addresses: %v", err)
	}
	defer rows.Close()

	var ipAddresses []models.IPAddress
	for rows.Next() {
		var ip models.IPAddress
		err := rows.Scan(
			&ip.IPAddress,
			&ip.InterfaceName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning IP address row: %v", err)
		}
		ipAddresses = append(ipAddresses, ip)
	}
	return ipAddresses, nil
}

// GetServerOpenPorts retrieves open ports for a server
func (d *Database) GetServerOpenPorts(serverID string) ([]models.Port, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			local_port,
			local_ip,
			CASE WHEN remote_port IS NULL THEN NULL ELSE remote_port END as remote_port,
			CASE WHEN remote_ip IS NULL THEN NULL ELSE remote_ip END as remote_ip,
			state,
			CASE WHEN description IS NULL THEN '' ELSE description END as description,
			process_id,
			CASE WHEN process_name IS NULL THEN '' ELSE process_name END as process_name
		FROM server_discovery.open_ports
		WHERE discovery_id IN (
			SELECT id FROM server_discovery.discovery_results
			WHERE server_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY local_port
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server open ports: %v", err)
	}
	defer rows.Close()

	var ports []models.Port
	for rows.Next() {
		var port models.Port
		var remotePort sql.NullInt64
		var remoteIP sql.NullString
		var processID sql.NullInt64
		err := rows.Scan(
			&port.LocalPort,
			&port.LocalIP,
			&remotePort,
			&remoteIP,
			&port.State,
			&port.Description,
			&processID,
			&port.ProcessName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning port row: %v", err)
		}

		if remotePort.Valid {
			port.RemotePort = int(remotePort.Int64)
		}
		if remoteIP.Valid {
			port.RemoteIP = remoteIP.String
		}
		if processID.Valid {
			pid := int(processID.Int64)
			port.ProcessID = &pid
		}

		ports = append(ports, port)
	}
	return ports, nil
}

func (d *Database) GetServerInstalledSoftware(serverID string) ([]models.Software, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			name,
			version,
			install_date
		FROM server_discovery.installed_software
		WHERE discovery_id IN (
			SELECT id FROM server_discovery.discovery_results
			WHERE server_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY name
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server installed software: %v", err)
	}
	defer rows.Close()

	var software []models.Software
	for rows.Next() {
		var s models.Software
		err := rows.Scan(
			&s.Name,
			&s.Version,
			&s.InstallDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning software row: %v", err)
		}
		software = append(software, s)
	}
	return software, nil
}

// GetServerFilesystems retrieves filesystem information for a server
func (d *Database) GetServerFilesystems(serverID string) ([]models.Filesystem, error) {
	// Convert serverID string to integer
	id, err := strconv.Atoi(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %v", err)
	}

	rows, err := d.db.Queryx(`
		SELECT 
			device,
			mount_point,
			fs_type,
			total_bytes,
			used_bytes,
			free_bytes,
			used_percent,
			total_inodes,
			used_inodes,
			free_inodes
		FROM server_discovery.filesystems
		WHERE discovery_id IN (
			SELECT id FROM server_discovery.discovery_results
			WHERE server_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY mount_point
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying server filesystems: %v", err)
	}
	defer rows.Close()

	var filesystems []models.Filesystem
	for rows.Next() {
		var fs models.Filesystem
		err := rows.Scan(
			&fs.Device,
			&fs.MountPoint,
			&fs.FSType,
			&fs.TotalBytes,
			&fs.UsedBytes,
			&fs.FreeBytes,
			&fs.UsedPercent,
			&fs.TotalInodes,
			&fs.UsedInodes,
			&fs.FreeInodes,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning filesystem row: %v", err)
		}
		filesystems = append(filesystems, fs)
	}
	return filesystems, nil
}

// GetServerTags retrieves tags for a server
func (d *Database) GetServerTags(serverID int) ([]models.Tag, error) {
	var tags []models.Tag
	query := `
		SELECT id, server_id, tag_name, tag_value, created_at, updated_at
		FROM server_discovery.server_tags
		WHERE server_id = $1
	`
	err := d.db.Select(&tags, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("error querying server tags: %w", err)
	}
	return tags, nil
}

// GetAllServerTags retrieves all unique tags from all servers
func (d *Database) GetAllServerTags() ([]models.Tag, error) {
	var tags []models.Tag
	query := `
		SELECT DISTINCT ON (tag_name, tag_value) id, server_id, tag_name, tag_value, created_at, updated_at
		FROM server_discovery.server_tags
		ORDER BY tag_name, tag_value
	`
	err := d.db.Select(&tags, query)
	if err != nil {
		return nil, fmt.Errorf("error querying server tags: %w", err)
	}
	return tags, nil
}

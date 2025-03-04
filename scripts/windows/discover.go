package windows

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

type Server struct {
	ID       int
	IP       string
	Hostname string
	OSType   string
	Status   string
}

// ServerWithDetails represents a server with additional details
type ServerWithDetails struct {
	ID                int                `json:"id"`
	Hostname          string             `json:"hostname"`
	Port              int                `json:"port"`
	Region            string             `json:"region"`
	Tags              []models.Tag       `json:"tags"`
	OSName            string             `json:"os_name"`
	OSVersion         string             `json:"os_version"`
	CPUModel          string             `json:"cpu_model"`
	CPUCount          int                `json:"cpu_count"`
	MemoryTotalGB     float64            `json:"memory_total_gb"`
	DiskTotalGB       float64            `json:"disk_total_gb"`
	DiskFreeGB        float64            `json:"disk_free_gb"`
	LastBootTime      time.Time          `json:"last_boot_time"`
	IPAddresses       []models.IPAddress `json:"ip_addresses"`
	InstalledSoftware []models.Software  `json:"installed_software"`
	RunningServices   []models.Service   `json:"running_services"`
	OpenPorts         []models.Port      `json:"open_ports"`
	DiscoveryCount    int                `json:"discovery_count"`
	LastDiscovery     time.Time          `json:"last_discovery"`
}

// Tag represents a key-value tag for a server
type Tag struct {
	TagName  string `json:"tag_name"`
	TagValue string `json:"tag_value"`
}

// IPAddress represents a network interface and its IP address
type IPAddress struct {
	IPAddress     string `json:"ip_address"`
	InterfaceName string `json:"interface_name"`
}

// Software represents installed software on a server
type Software struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Vendor      string `json:"vendor"`
	InstallDate string `json:"install_date"`
}

// Port represents an open network port on a server
type Port struct {
	LocalPort   int    `json:"local_port"`
	LocalIP     string `json:"local_ip,omitempty"`
	RemotePort  int    `json:"remote_port,omitempty"`
	RemoteIP    string `json:"remote_ip,omitempty"`
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
	ProcessID   int    `json:"process_id,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
}

func DiscoverServers() {
	// Connect to PostgreSQL
	connStr := "host=localhost port=5433 user=postgres password=postgres dbname=server_discovery sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Get all Windows servers
	rows, err := db.Query(`
		SELECT id, ip, hostname, os_type, status 
		FROM servers 
		WHERE os_type LIKE 'Windows Server%'
	`)
	if err != nil {
		log.Fatalf("Failed to query Windows servers: %v", err)
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var s Server
		err := rows.Scan(&s.ID, &s.IP, &s.Hostname, &s.OSType, &s.Status)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		servers = append(servers, s)
	}

	log.Printf("Found %d Windows servers", len(servers))

	// Process each server
	for _, server := range servers {
		processServer(db, server)
	}
}

func processServer(db *sql.DB, server Server) {
	log.Printf("Processing server %s (%s)", server.Hostname, server.IP)

	// Get existing server details from database
	var cpuModel, osVersion string
	var cpuCount int
	var memoryTotal, diskTotal float64
	err := db.QueryRow(`
		SELECT 
			cpu_model, 
			cpu_cores,
			memory_total,
			disk_total,
			os_version
		FROM server_details 
		WHERE server_id = $1
	`, server.ID).Scan(&cpuModel, &cpuCount, &memoryTotal, &diskTotal, &osVersion)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting server details: %v", err)
		updateServerStatus(db, server.ID, "error", err.Error())
		return
	}

	// If no existing details, create new ones with simulated values
	if err == sql.ErrNoRows {
		// Use different CPU models for Windows servers
		cpuModels := []string{
			"Intel(R) Xeon(R) E5-2680 v4 @ 2.40GHz",
			"Intel(R) Xeon(R) E5-2690 v4 @ 2.60GHz",
			"Intel(R) Xeon(R) Gold 6248R CPU @ 3.00GHz",
			"Intel(R) Xeon(R) Platinum 8280 CPU @ 2.70GHz",
		}
		cpuModel = cpuModels[server.ID%len(cpuModels)]

		cpuCount = 4 + (server.ID % 60)                // 4-64 cores
		memoryTotal = 16.0 + float64(server.ID%32)*8.0 // 16-272 GB
		diskTotal = 256.0 + float64(server.ID%8)*256.0 // 256-2048 GB

		// Use different Windows Server versions
		osTypes := []string{
			"Windows Server 2019 Datacenter",
			"Windows Server 2019 Standard",
			"Windows Server 2016 Datacenter",
			"Windows Server 2016 Standard",
		}
		osVersion = osTypes[server.ID%len(osTypes)]

		_, err = db.Exec(`
			INSERT INTO server_details (
				server_id, 
				cpu_model, 
				cpu_cores, 
				memory_total, 
				disk_total, 
				os_version, 
				created_at, 
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			server.ID,
			cpuModel,
			cpuCount,
			memoryTotal,
			diskTotal,
			osVersion,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			log.Printf("Error inserting server details: %v", err)
			updateServerStatus(db, server.ID, "error", err.Error())
			return
		}
	}

	// Get running services
	rows, err := db.Query(`
		SELECT service_name, status, last_checked
		FROM server_services
		WHERE server_id = $1
		ORDER BY last_checked DESC
		LIMIT 10
	`, server.ID)
	if err != nil {
		log.Printf("Error getting services: %v", err)
		updateServerStatus(db, server.ID, "error", err.Error())
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		var lastChecked time.Time
		err := rows.Scan(&s.Name, &s.Status, &lastChecked)
		if err != nil {
			log.Printf("Error scanning service row: %v", err)
			continue
		}
		services = append(services, s)
	}

	// Update metrics with simulated values based on historical data
	err = updateServerMetrics(db, server.ID, cpuCount, memoryTotal, diskTotal)
	if err != nil {
		log.Printf("Error updating metrics: %v", err)
		updateServerStatus(db, server.ID, "error", err.Error())
		return
	}

	log.Printf("Successfully processed server %s", server.Hostname)
	updateServerStatus(db, server.ID, "online", "")
}

func updateServerMetrics(db *sql.DB, serverID, cpuCount int, memoryTotal, diskTotal float64) error {
	// Get last metrics
	var lastCPUUsage, lastMemoryUsage, lastDiskUsage float64
	err := db.QueryRow(`
		SELECT cpu_usage, memory_usage, disk_usage
		FROM server_metrics
		WHERE server_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`, serverID).Scan(&lastCPUUsage, &lastMemoryUsage, &lastDiskUsage)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Simulate slight changes in metrics
	cpuUsage := simulateMetric(lastCPUUsage, 40, 80)       // Windows servers typically have higher CPU usage
	memoryUsage := simulateMetric(lastMemoryUsage, 50, 85) // And higher memory usage
	diskUsage := simulateMetric(lastDiskUsage, 40, 90)     // And more variable disk usage

	// Insert new metrics
	_, err = db.Exec(`
		INSERT INTO server_metrics (server_id, cpu_usage, memory_usage, disk_usage, recorded_at)
		VALUES ($1, $2, $3, $4, $5)`,
		serverID,
		cpuUsage,
		memoryUsage,
		diskUsage,
		time.Now(),
	)
	return err
}

func simulateMetric(lastValue float64, min, max float64) float64 {
	if lastValue == 0 {
		// Initial value if no history
		return min + (max-min)*0.5
	}

	// Simulate small change (-5% to +5%)
	change := (float64(time.Now().UnixNano()%100) - 50) * 0.1
	newValue := lastValue + change

	// Keep within bounds
	if newValue < min {
		newValue = min
	}
	if newValue > max {
		newValue = max
	}
	return newValue
}

func updateServerStatus(db *sql.DB, serverID int, status string, errorMsg string) {
	_, err := db.Exec(`
		UPDATE servers 
		SET status = $1, last_checked = $2, last_error = $3
		WHERE id = $4`,
		status, time.Now(), errorMsg, serverID)
	if err != nil {
		log.Printf("Error updating server status: %v", err)
	}
}

func updateServerDetails(db *sql.DB, serverID int, details map[string]interface{}) error {
	// Start transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Update server metrics
	if cpu, ok := details["cpu"].(map[string]interface{}); ok {
		_, err = tx.Exec(`
			INSERT INTO server_metrics (server_id, cpu_usage, memory_usage, disk_usage, recorded_at)
			VALUES ($1, $2, $3, $4, $5)`,
			serverID,
			cpu["usage"],
			getMemoryUsage(details),
			getDiskUsage(details),
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert metrics: %v", err)
		}
	}

	// Update services
	if services, ok := details["services"].([]models.Service); ok {
		for _, service := range services {
			_, err = tx.Exec(`
				INSERT INTO server_services (server_id, service_name, status, last_checked)
				VALUES ($1, $2, $3, $4)`,
				serverID,
				service.Name,
				service.Status,
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("failed to insert service: %v", err)
			}
		}
	}

	// Update server status
	_, err = tx.Exec(`
		UPDATE servers 
		SET status = 'online', last_checked = $1 
		WHERE id = $2`,
		time.Now(), serverID)
	if err != nil {
		return fmt.Errorf("failed to update server status: %v", err)
	}

	return tx.Commit()
}

func getMemoryUsage(details map[string]interface{}) float64 {
	if mem, ok := details["memory"].(map[string]interface{}); ok {
		total := mem["total"].(float64)
		used := mem["used"].(float64)
		if total > 0 {
			return (used / total) * 100
		}
	}
	return 0
}

func getDiskUsage(details map[string]interface{}) float64 {
	if disk, ok := details["disk"].(map[string]interface{}); ok {
		if drives, ok := disk["drives"].([]map[string]interface{}); ok && len(drives) > 0 {
			drive := drives[0]
			total := drive["total"].(float64)
			used := drive["used"].(float64)
			if total > 0 {
				return (used / total) * 100
			}
		}
	}
	return 0
}

// getMockServerWithDetails returns mock server details for testing
func getMockServerWithDetails(id int) *ServerWithDetails {
	server := &ServerWithDetails{
		ID:       id,
		Hostname: fmt.Sprintf("WIN-SERVER-%d", id),
		Port:     3389,
		Region:   "us-east-1",
		Tags: []models.Tag{
			{TagName: "Environment", TagValue: "Production"},
			{TagName: "Role", TagValue: "Application Server"},
		},
		OSName:        "Windows Server 2019",
		OSVersion:     "10.0.17763",
		CPUModel:      "Intel(R) Xeon(R) CPU E5-2680 v3 @ 2.50GHz",
		CPUCount:      4,
		MemoryTotalGB: 16.0,
		DiskTotalGB:   500.0,
		DiskFreeGB:    350.0,
		LastBootTime:  time.Now().Add(-24 * time.Hour),
		IPAddresses: []models.IPAddress{
			{
				IPAddress:     fmt.Sprintf("192.168.1.%d", id),
				InterfaceName: "Ethernet0",
			},
		},
		InstalledSoftware: []models.Software{
			{
				Name:        "Microsoft .NET Framework 4.8",
				Version:     "4.8.03761",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2023-01-15",
			},
			{
				Name:        "Windows Server 2019 Standard",
				Version:     "10.0.17763",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2023-01-01",
			},
		},
		RunningServices: []models.Service{
			{
				Name:        "W32Time",
				DisplayName: "Windows Time",
				Status:      "Running",
				Description: "Windows Time Service",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\LocalService",
			},
			{
				Name:        "WinRM",
				DisplayName: "Windows Remote Management (WS-Management)",
				Status:      "Running",
				Description: "Windows Remote Management Service",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\NetworkService",
			},
		},
		OpenPorts: []models.Port{
			{
				LocalPort:   3389,
				State:       "LISTENING",
				Description: "Remote Desktop",
				ProcessName: "TermService",
			},
			{
				LocalPort:   445,
				State:       "LISTENING",
				Description: "SMB",
				ProcessName: "System",
			},
		},
		DiscoveryCount: 1,
		LastDiscovery:  time.Now(),
	}
	return server
}

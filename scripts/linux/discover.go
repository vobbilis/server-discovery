package linux

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type LinuxServer struct {
	ID       int
	IP       string
	Hostname string
	OSType   string
	Status   string
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

	// Get all Linux servers
	rows, err := db.Query(`
		SELECT id, ip, hostname, os_type, status 
		FROM servers 
		WHERE os_type NOT LIKE 'Windows Server%'
	`)
	if err != nil {
		log.Fatalf("Failed to query Linux servers: %v", err)
	}
	defer rows.Close()

	var servers []LinuxServer
	for rows.Next() {
		var s LinuxServer
		err := rows.Scan(&s.ID, &s.IP, &s.Hostname, &s.OSType, &s.Status)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		servers = append(servers, s)
	}

	log.Printf("Found %d Linux servers", len(servers))

	// Create a worker pool
	workerCount := 10
	serverChan := make(chan LinuxServer)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for server := range serverChan {
				discoverLinuxServerDetails(db, server)
			}
		}()
	}

	// Send servers to workers
	for _, server := range servers {
		serverChan <- server
	}
	close(serverChan)

	// Wait for all workers to finish
	wg.Wait()
}

func discoverLinuxServerDetails(db *sql.DB, server LinuxServer) {
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
		updateLinuxServerStatus(db, server.ID, "error", err.Error())
		return
	}

	// If no existing details, create new ones with simulated values
	if err == sql.ErrNoRows {
		// Use different CPU models for Linux servers
		cpuModels := []string{
			"AMD EPYC 7763 64-Core Processor",
			"AMD EPYC 7542 32-Core Processor",
			"Intel(R) Xeon(R) Platinum 8380 CPU @ 2.30GHz",
			"Intel(R) Xeon(R) Gold 6330 CPU @ 2.00GHz",
		}
		cpuModel = cpuModels[server.ID%len(cpuModels)]

		cpuCount = 8 + (server.ID % 56)                // 8-64 cores
		memoryTotal = 32.0 + float64(server.ID%32)*8.0 // 32-288 GB
		diskTotal = 512.0 + float64(server.ID%8)*512.0 // 512-4096 GB

		// Use different Linux distributions and versions
		osTypes := []string{
			"Ubuntu 22.04 LTS",
			"CentOS 7.9",
			"Red Hat Enterprise Linux 8.6",
			"SUSE Linux Enterprise Server 15 SP4",
			"Debian 11.6",
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
			updateLinuxServerStatus(db, server.ID, "error", err.Error())
			return
		}
	}

	// Update metrics with simulated values based on historical data
	err = updateLinuxServerMetrics(db, server.ID, cpuCount, memoryTotal, diskTotal)
	if err != nil {
		log.Printf("Error updating metrics: %v", err)
		updateLinuxServerStatus(db, server.ID, "error", err.Error())
		return
	}

	log.Printf("Successfully processed server %s", server.Hostname)
	updateLinuxServerStatus(db, server.ID, "online", "")
}

func updateLinuxServerMetrics(db *sql.DB, serverID, cpuCount int, memoryTotal, diskTotal float64) error {
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
	cpuUsage := simulateLinuxMetric(lastCPUUsage, 30, 70)       // Linux servers typically have lower CPU usage
	memoryUsage := simulateLinuxMetric(lastMemoryUsage, 40, 75) // And lower memory usage
	diskUsage := simulateLinuxMetric(lastDiskUsage, 30, 80)     // And more controlled disk usage

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

func simulateLinuxMetric(lastValue float64, min, max float64) float64 {
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

func updateLinuxServerStatus(db *sql.DB, serverID int, status string, errorMsg string) {
	_, err := db.Exec(`
		UPDATE servers 
		SET status = $1, last_checked = $2, last_error = $3
		WHERE id = $4`,
		status, time.Now(), errorMsg, serverID)
	if err != nil {
		log.Printf("Error updating server status: %v", err)
	}
}

// Package main provides a tool for generating test data in the server discovery database
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

// Database Configuration Constants
const (
	// DatabaseName is the name of the PostgreSQL database used for server discovery
	DatabaseName = "server_discovery"

	// SchemaName is the name of the schema where all tables are created
	SchemaName = "server_discovery"

	// Default connection parameters
	DefaultHost     = "localhost"
	DefaultPort     = 5433 // Note: Using port 5433 for test database
	DefaultUser     = "postgres"
	DefaultPassword = "postgres"
)

// getConnectionString returns the database connection string based on the provided parameters
func getConnectionString(host string, port int, dbname, user, password string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func main() {
	// Use the configured connection string
	connStr := getConnectionString(DefaultHost, DefaultPort, DatabaseName, DefaultUser, DefaultPassword)

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE SCHEMA IF NOT EXISTS server_discovery;
		SET search_path TO server_discovery, public;

		CREATE TABLE IF NOT EXISTS server_discovery.servers (
			id SERIAL PRIMARY KEY,
			hostname VARCHAR(255) NOT NULL,
			region VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ip VARCHAR(50) NOT NULL,
			os_type VARCHAR(50) DEFAULT 'windows',
			status VARCHAR(50) NOT NULL,
			last_checked TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS server_discovery.server_services (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
			service_name VARCHAR(255) NOT NULL,
			service_status VARCHAR(50) NOT NULL,
			service_description TEXT,
			port INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS server_discovery.discovery_results (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
			success BOOLEAN NOT NULL,
			message TEXT,
			error TEXT,
			start_time TIMESTAMP WITH TIME ZONE,
			end_time TIMESTAMP WITH TIME ZONE,
			output_path TEXT,
			os_name VARCHAR(255),
			os_version VARCHAR(255),
			cpu_model VARCHAR(255),
			cpu_count INTEGER,
			memory_total_gb NUMERIC(10, 2),
			disk_total_gb NUMERIC(10, 2),
			disk_free_gb NUMERIC(10, 2),
			last_boot_time TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS server_discovery.open_ports (
			id SERIAL PRIMARY KEY,
			discovery_id INTEGER REFERENCES server_discovery.discovery_results(id) ON DELETE CASCADE,
			local_port INTEGER NOT NULL,
			local_ip VARCHAR(50),
			remote_port INTEGER,
			remote_ip VARCHAR(50),
			state VARCHAR(50),
			description VARCHAR(255),
			process_id INTEGER,
			process_name VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	log.Println("Successfully created schema and tables")

	// Configuration for test data
	config := struct {
		totalServers       int
		windowsPercentage  float64
		linuxDistributions []string
		windowsVersions    []string
		regions            []string
		services           []struct {
			name string
			port int
		}
		subnets []string
	}{
		totalServers:      500, // Generate 500 servers
		windowsPercentage: 0.4, // 40% Windows servers
		regions: []string{
			"us-west",
			"us-east",
			"eu-central",
		},
		linuxDistributions: []string{
			"Ubuntu 22.04 LTS",
			"CentOS 7",
			"Red Hat Enterprise Linux 8",
			"Debian 11",
			"Amazon Linux 2",
			"SUSE Linux Enterprise 15",
		},
		windowsVersions: []string{
			"Windows Server 2022",
			"Windows Server 2019",
			"Windows Server 2016",
			"Windows Server 2012 R2",
		},
		services: []struct {
			name string
			port int
		}{
			{"SSH", 22},
			{"HTTP", 80},
			{"HTTPS", 443},
			{"MySQL", 3306},
			{"PostgreSQL", 5432},
			{"MongoDB", 27017},
			{"Redis", 6379},
			{"SMTP", 25},
			{"DNS", 53},
			{"LDAP", 389},
		},
		subnets: []string{
			"10.0.0",
			"172.16.0",
			"192.168.1",
			"192.168.2",
			"10.10.0",
		},
	}

	// Start transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	startTime := time.Now()
	log.Printf("Starting to generate %d servers...", config.totalServers)

	// Prepare statements
	stmtServer, err := tx.Prepare(`
		INSERT INTO server_discovery.servers (ip, hostname, region, os_type, status, last_checked)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)
	if err != nil {
		log.Fatalf("Failed to prepare server statement: %v", err)
	}

	stmtService, err := tx.Prepare(`
		INSERT INTO server_discovery.server_services (server_id, service_name, service_status, service_description, port)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		log.Fatalf("Failed to prepare service statement: %v", err)
	}

	stmtDiscovery, err := tx.Prepare(`
		INSERT INTO server_discovery.discovery_results (
			server_id, success, message, start_time, end_time,
			os_name, os_version, cpu_model, cpu_count,
			memory_total_gb, disk_total_gb, disk_free_gb, last_boot_time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`)
	if err != nil {
		log.Fatalf("Failed to prepare discovery statement: %v", err)
	}

	stmtOpenPorts, err := tx.Prepare(`
		INSERT INTO server_discovery.open_ports (
			discovery_id, local_port, local_ip, state, description, process_name
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		log.Fatalf("Failed to prepare open ports statement: %v", err)
	}

	// Generate servers
	for i := 0; i < config.totalServers; i++ {
		// Generate basic server info
		subnet := config.subnets[rand.Intn(len(config.subnets))]
		ip := fmt.Sprintf("%s.%d", subnet, rand.Intn(254)+1)
		hostname := fmt.Sprintf("server-%s-%d", subnet, i)
		region := config.regions[rand.Intn(len(config.regions))]

		// Determine OS
		var osType string
		if rand.Float64() < config.windowsPercentage {
			osType = config.windowsVersions[rand.Intn(len(config.windowsVersions))]
		} else {
			osType = config.linuxDistributions[rand.Intn(len(config.linuxDistributions))]
		}

		// Random status (mostly online)
		status := "online"
		if rand.Float64() < 0.05 { // 5% chance of being offline
			status = "offline"
		}

		lastChecked := time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)

		// Insert server
		var serverID int
		err := stmtServer.QueryRow(
			ip,
			hostname,
			region,
			osType,
			status,
			lastChecked,
		).Scan(&serverID)
		if err != nil {
			log.Fatalf("Failed to insert server: %v", err)
		}

		// Add discovery result
		discoveryStart := lastChecked.Add(-time.Duration(rand.Intn(60)) * time.Minute)
		discoveryEnd := discoveryStart.Add(time.Duration(rand.Intn(300)) * time.Second)
		var discoveryID int
		err = stmtDiscovery.QueryRow(
			serverID,
			true,
			"Discovery completed successfully",
			discoveryStart,
			discoveryEnd,
			osType,
			"1.0",
			"Intel(R) Xeon(R) CPU @ 2.20GHz",
			8,
			32.0,
			500.0,
			350.0,
			discoveryStart.Add(-time.Duration(rand.Intn(720))*time.Hour),
		).Scan(&discoveryID)
		if err != nil {
			log.Fatalf("Failed to insert discovery result: %v", err)
		}

		// Add random services (3-8 services per server)
		numServices := rand.Intn(6) + 3
		usedPorts := make(map[int]bool)
		for j := 0; j < numServices; j++ {
			service := config.services[rand.Intn(len(config.services))]
			if !usedPorts[service.port] {
				_, err = stmtService.Exec(
					serverID,
					service.name,
					"running",
					fmt.Sprintf("%s service", service.name),
					service.port,
				)
				if err != nil {
					log.Fatalf("Failed to insert service: %v", err)
				}
				usedPorts[service.port] = true

				// Add corresponding open port
				_, err = stmtOpenPorts.Exec(
					discoveryID,
					service.port,
					ip,
					"LISTENING",
					fmt.Sprintf("%s service port", service.name),
					fmt.Sprintf("%s-service", service.name),
				)
				if err != nil {
					log.Fatalf("Failed to insert open port: %v", err)
				}
			}
		}

		if (i+1)%50 == 0 {
			log.Printf("Generated %d servers...", i+1)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("Successfully generated %d servers in %v", config.totalServers, duration)

	// Verify the data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.servers").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count servers: %v", err)
	}
	log.Printf("Total servers in database: %d", count)

	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.server_services").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count services: %v", err)
	}
	log.Printf("Total services in database: %d", count)

	// Print some statistics
	var windowsCount int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.servers WHERE os_type LIKE 'Windows%'").Scan(&windowsCount)
	if err != nil {
		log.Fatalf("Failed to count Windows servers: %v", err)
	}
	log.Printf("Windows servers: %d (%.1f%%)", windowsCount, float64(windowsCount)/float64(config.totalServers)*100)

	log.Println("\nSample of generated servers:")
	rows, err := db.Query(`
		SELECT s.ip, s.hostname, s.os_type, s.status, COUNT(ss.id) as service_count
		FROM server_discovery.servers s
		LEFT JOIN server_discovery.server_services ss ON s.id = ss.server_id
		GROUP BY s.id, s.ip, s.hostname, s.os_type, s.status
		LIMIT 5
	`)
	if err != nil {
		log.Fatalf("Failed to query sample servers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			ip, hostname, osType, status string
			serviceCount                 int
		)
		if err := rows.Scan(&ip, &hostname, &osType, &status, &serviceCount); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		log.Printf("Server: %s (%s) - OS: %s, Status: %s, Services: %d",
			hostname, ip, osType, status, serviceCount)
	}
}

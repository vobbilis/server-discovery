// Package main provides database setup and test data generation utilities.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
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

// TestLoadDatabaseWithServers populates the database with diverse server configurations
func TestLoadDatabaseWithServers(t *testing.T) {
	// Use the configured connection string
	connStr := getConnectionString(DefaultHost, DefaultPort, DatabaseName, DefaultUser, DefaultPassword)

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

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
		t.Fatalf("Failed to create tables: %v", err)
	}

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
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	startTime := time.Now()
	t.Logf("Starting to generate %d servers...", config.totalServers)

	// Prepare statements
	stmtServer, err := tx.Prepare(`
		INSERT INTO server_discovery.servers (ip, hostname, region, os_type, status, last_checked)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)
	if err != nil {
		t.Fatalf("Failed to prepare server statement: %v", err)
	}

	stmtService, err := tx.Prepare(`
		INSERT INTO server_discovery.server_services (server_id, service_name, service_status, service_description, port)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		t.Fatalf("Failed to prepare service statement: %v", err)
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
		t.Fatalf("Failed to prepare discovery statement: %v", err)
	}

	stmtOpenPorts, err := tx.Prepare(`
		INSERT INTO server_discovery.open_ports (
			discovery_id, local_port, local_ip, state, description, process_name
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		t.Fatalf("Failed to prepare open ports statement: %v", err)
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
			t.Fatalf("Failed to insert server: %v", err)
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
			t.Fatalf("Failed to insert discovery result: %v", err)
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
					t.Fatalf("Failed to insert service: %v", err)
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
					t.Fatalf("Failed to insert open port: %v", err)
				}
			}
		}

		if (i+1)%50 == 0 {
			t.Logf("Generated %d servers...", i+1)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	duration := time.Since(startTime)
	t.Logf("Successfully generated %d servers in %v", config.totalServers, duration)

	// Verify the data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.servers").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count servers: %v", err)
	}
	t.Logf("Total servers in database: %d", count)

	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.server_services").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count services: %v", err)
	}
	t.Logf("Total services in database: %d", count)

	err = db.QueryRow("SELECT COUNT(*) FROM server_metrics").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count metrics: %v", err)
	}
	t.Logf("Total metrics in database: %d", count)

	// Print some statistics
	var windowsCount int
	err = db.QueryRow("SELECT COUNT(*) FROM server_discovery.servers WHERE os_type LIKE 'Windows%'").Scan(&windowsCount)
	if err != nil {
		t.Fatalf("Failed to count Windows servers: %v", err)
	}
	t.Logf("Windows servers: %d (%.1f%%)", windowsCount, float64(windowsCount)/float64(config.totalServers)*100)

	t.Log("\nSample of generated servers:")
	rows, err := db.Query(`
		SELECT s.ip, s.hostname, s.os_type, s.status, 
			   COUNT(ss.id) as service_count,
			   sm.cpu_usage, sm.memory_usage, sm.disk_usage
		FROM server_discovery.servers s
		LEFT JOIN server_discovery.server_services ss ON s.id = ss.server_id
		LEFT JOIN server_metrics sm ON s.id = sm.server_id
		GROUP BY s.id, s.ip, s.hostname, s.os_type, s.status, sm.cpu_usage, sm.memory_usage, sm.disk_usage
		LIMIT 5
	`)
	if err != nil {
		t.Fatalf("Failed to query sample servers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			ip, hostname, osType, status string
			serviceCount                 int
			cpu, memory, disk            float64
		)
		if err := rows.Scan(&ip, &hostname, &osType, &status, &serviceCount, &cpu, &memory, &disk); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		t.Logf("Server: %s (%s) - OS: %s, Status: %s, Services: %d, CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%",
			hostname, ip, osType, status, serviceCount, cpu, memory, disk)
	}
}

type testConfig struct {
	totalServers       int
	windowsPercentage  float64
	subnets            []string
	regions            []string
	services           []serviceConfig
	windowsVersions    []string
	linuxDistributions []string
}

type serviceConfig struct {
	name string
	port int
}

func getTestConfig() testConfig {
	return testConfig{
		totalServers:      100,
		windowsPercentage: 0.3,
		subnets: []string{
			"192.168.1",
			"192.168.2",
			"10.0.1",
			"10.0.2",
		},
		regions: []string{
			"us-west",
			"us-east",
			"eu-central",
			"ap-southeast",
		},
		services: []serviceConfig{
			{name: "http", port: 80},
			{name: "https", port: 443},
			{name: "ssh", port: 22},
			{name: "mysql", port: 3306},
			{name: "postgresql", port: 5432},
			{name: "redis", port: 6379},
			{name: "mongodb", port: 27017},
			{name: "elasticsearch", port: 9200},
			{name: "prometheus", port: 9090},
			{name: "grafana", port: 3000},
		},
		windowsVersions: []string{
			"Windows Server 2019",
			"Windows Server 2016",
			"Windows Server 2012 R2",
		},
		linuxDistributions: []string{
			"Ubuntu 20.04",
			"CentOS 8",
			"Red Hat Enterprise Linux 8",
			"Amazon Linux 2",
			"Debian 10",
		},
	}
}

// TestLoadDatabase is an alternative implementation for database population
// that drops and recreates the schema each time.
// Note: This implementation uses a different approach from TestLoadDatabaseWithServers
// and should not be used in conjunction with it.
func TestLoadDatabase(t *testing.T) {
	// Use the configured connection string
	connStr := getConnectionString(DefaultHost, DefaultPort, DatabaseName, DefaultUser, DefaultPassword)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create schema and tables
	_, err = db.Exec(fmt.Sprintf(`
		DROP SCHEMA IF EXISTS %s CASCADE;
		CREATE SCHEMA %s;

		CREATE TABLE %s.servers (
			id SERIAL PRIMARY KEY,
			ip VARCHAR(255) NOT NULL,
			hostname VARCHAR(255) NOT NULL,
			region VARCHAR(50) NOT NULL,
			os_type VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			last_checked TIMESTAMP NOT NULL
		);

		CREATE TABLE %s.server_services (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES %s.servers(id),
			service_name VARCHAR(255) NOT NULL,
			service_status VARCHAR(50) NOT NULL,
			service_description TEXT,
			port INTEGER NOT NULL,
			UNIQUE(server_id, port)
		);

		CREATE TABLE %s.discovery_results (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES %s.servers(id),
			success BOOLEAN NOT NULL,
			message TEXT,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			os_name VARCHAR(255),
			os_version VARCHAR(50),
			cpu_model VARCHAR(255),
			cpu_count INTEGER,
			memory_total_gb FLOAT,
			disk_total_gb FLOAT,
			disk_free_gb FLOAT,
			last_boot_time TIMESTAMP
		);

		CREATE TABLE %s.open_ports (
			id SERIAL PRIMARY KEY,
			discovery_id INTEGER REFERENCES %s.discovery_results(id),
			local_port INTEGER NOT NULL,
			local_ip VARCHAR(255) NOT NULL,
			remote_port INTEGER,
			remote_ip VARCHAR(255),
			state VARCHAR(50) NOT NULL,
			description TEXT,
			process_id INTEGER,
			process_name VARCHAR(255)
		);
	`, SchemaName, SchemaName, SchemaName, SchemaName, SchemaName, SchemaName, SchemaName, SchemaName))
	if err != nil {
		t.Fatalf("Failed to create schema and tables: %v", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Get test configuration
	config := getTestConfig()

	// ... existing code for preparing statements and generating data ...

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	t.Logf("Successfully generated %d servers with services and discovery results", config.totalServers)
}

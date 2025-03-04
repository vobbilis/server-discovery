package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func setupTestDB(t *testing.T) {
	var err error
	// Connect to PostgreSQL test database in Docker
	// Using port 5433 as specified in docker-compose.test.yml
	connStr := "host=localhost port=5433 user=postgres password=postgres dbname=server_discovery_test sslmode=disable"

	// Allow overriding connection string through environment variable for CI/CD
	if envConnStr := os.Getenv("TEST_DB_CONN"); envConnStr != "" {
		connStr = envConnStr
	}

	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Verify connection
	err = testDB.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create schema and tables
	_, err = testDB.Exec(`
		DROP SCHEMA IF EXISTS server_discovery CASCADE;
		CREATE SCHEMA server_discovery;

		CREATE TABLE server_discovery.servers (
			id SERIAL PRIMARY KEY,
			hostname VARCHAR(255) NOT NULL,
			region VARCHAR(50),
			os_type VARCHAR(50) DEFAULT 'windows',
			discovery_count INTEGER DEFAULT 0,
			last_discovery TIMESTAMP WITH TIME ZONE
		);

		CREATE TABLE server_discovery.server_details (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
			os_name VARCHAR(255),
			os_version VARCHAR(100),
			os_type VARCHAR(50),
			kernel_version VARCHAR(100),
			package_manager VARCHAR(50),
			init_system VARCHAR(50),
			selinux_status VARCHAR(50),
			firewall_status VARCHAR(50),
			cpu_model VARCHAR(255),
			cpu_count INTEGER,
			memory_total_gb NUMERIC(10,2),
			disk_total_gb NUMERIC(10,2),
			disk_free_gb NUMERIC(10,2),
			last_boot_time TIMESTAMP WITH TIME ZONE
		);

		CREATE TABLE server_discovery.discoveries (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
			success BOOLEAN NOT NULL,
			message TEXT,
			start_time TIMESTAMP WITH TIME ZONE,
			end_time TIMESTAMP WITH TIME ZONE,
			output_path TEXT,
			error TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}
}

func cleanupTestDB(t *testing.T) {
	if testDB != nil {
		// Drop the test schema
		_, err := testDB.Exec("DROP SCHEMA IF EXISTS server_discovery CASCADE")
		if err != nil {
			t.Errorf("Failed to drop test schema: %v", err)
		}

		if err := testDB.Close(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	}
}

func TestDatabaseOperations(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	t.Run("Insert and Retrieve Server", func(t *testing.T) {
		// Insert test server
		server := ServerConfig{
			Hostname: "test-server",
			Region:   "us-west-1",
			OSType:   "linux",
		}

		result, err := testDB.Exec(`
			INSERT INTO server_discovery.servers (hostname, region, os_type)
			VALUES ($1, $2, $3)
			RETURNING id
		`, server.Hostname, server.Region, server.OSType)
		if err != nil {
			t.Fatalf("Failed to insert server: %v", err)
		}

		serverID, err := result.LastInsertId()
		if err != nil {
			// PostgreSQL doesn't support LastInsertId, we need to use RETURNING
			var id int64
			err = testDB.QueryRow(`
				SELECT id FROM server_discovery.servers 
				WHERE hostname = $1
			`, server.Hostname).Scan(&id)
			if err != nil {
				t.Fatalf("Failed to get server ID: %v", err)
			}
			serverID = id
		}

		// Retrieve and verify server
		var retrieved ServerConfig
		err = testDB.QueryRow(`
			SELECT hostname, region, os_type
			FROM server_discovery.servers WHERE id = $1
		`, serverID).Scan(&retrieved.Hostname, &retrieved.Region, &retrieved.OSType)
		if err != nil {
			t.Fatalf("Failed to retrieve server: %v", err)
		}

		if retrieved.Hostname != server.Hostname {
			t.Errorf("Expected hostname %s, got %s", server.Hostname, retrieved.Hostname)
		}
		if retrieved.Region != server.Region {
			t.Errorf("Expected region %s, got %s", server.Region, retrieved.Region)
		}
		if retrieved.OSType != server.OSType {
			t.Errorf("Expected OS type %s, got %s", server.OSType, retrieved.OSType)
		}
	})

	t.Run("Store and Retrieve Discovery Result", func(t *testing.T) {
		// Insert test server first
		var serverID int64
		err := testDB.QueryRow(`
			INSERT INTO server_discovery.servers (hostname, region, os_type)
			VALUES ($1, $2, $3)
			RETURNING id
		`, "test-server", "us-west-1", "linux").Scan(&serverID)
		if err != nil {
			t.Fatalf("Failed to insert server: %v", err)
		}

		// Create test discovery result
		discoveryResult := DiscoveryResult{
			Server:     "test-server",
			Success:    true,
			Message:    "Test discovery completed",
			StartTime:  time.Now(),
			EndTime:    time.Now().Add(time.Minute),
			OutputPath: "/tmp/test-output",
		}

		// Store discovery result
		_, err = testDB.Exec(`
			INSERT INTO server_discovery.discoveries (
				server_id, success, message, start_time, end_time, output_path
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, serverID, discoveryResult.Success, discoveryResult.Message,
			discoveryResult.StartTime, discoveryResult.EndTime, discoveryResult.OutputPath)
		if err != nil {
			t.Fatalf("Failed to insert discovery result: %v", err)
		}

		// Retrieve and verify discovery result
		var retrieved DiscoveryResult
		err = testDB.QueryRow(`
			SELECT success, message, output_path
			FROM server_discovery.discoveries WHERE server_id = $1
		`, serverID).Scan(&retrieved.Success, &retrieved.Message, &retrieved.OutputPath)
		if err != nil {
			t.Fatalf("Failed to retrieve discovery result: %v", err)
		}

		if retrieved.Success != discoveryResult.Success {
			t.Errorf("Expected success %v, got %v", discoveryResult.Success, retrieved.Success)
		}
		if retrieved.Message != discoveryResult.Message {
			t.Errorf("Expected message %s, got %s", discoveryResult.Message, retrieved.Message)
		}
	})

	t.Run("Store and Retrieve Server Details", func(t *testing.T) {
		// Insert test server first
		var serverID int64
		err := testDB.QueryRow(`
			INSERT INTO server_discovery.servers (hostname, region, os_type)
			VALUES ($1, $2, $3)
			RETURNING id
		`, "test-server", "us-west-1", "linux").Scan(&serverID)
		if err != nil {
			t.Fatalf("Failed to insert server: %v", err)
		}

		// Create test server details
		details := ServerDetails{
			OSName:         "Ubuntu",
			OSVersion:      "20.04",
			OSType:         "linux",
			KernelVersion:  "5.4.0",
			PackageManager: "apt",
			InitSystem:     "systemd",
			CPUModel:       "Intel(R) Xeon(R)",
			CPUCount:       4,
			MemoryTotalGB:  16.0,
			DiskTotalGB:    100.0,
			DiskFreeGB:     50.0,
		}

		// Store server details
		_, err = testDB.Exec(`
			INSERT INTO server_discovery.server_details (
				server_id, os_name, os_version, os_type, kernel_version,
				package_manager, init_system, cpu_model, cpu_count,
				memory_total_gb, disk_total_gb, disk_free_gb
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, serverID, details.OSName, details.OSVersion, details.OSType,
			details.KernelVersion, details.PackageManager, details.InitSystem,
			details.CPUModel, details.CPUCount, details.MemoryTotalGB,
			details.DiskTotalGB, details.DiskFreeGB)
		if err != nil {
			t.Fatalf("Failed to insert server details: %v", err)
		}

		// Retrieve and verify server details
		var retrieved ServerDetails
		err = testDB.QueryRow(`
			SELECT os_name, os_version, os_type, kernel_version,
				   package_manager, init_system, cpu_model, cpu_count,
				   memory_total_gb, disk_total_gb, disk_free_gb
			FROM server_discovery.server_details WHERE server_id = $1
		`, serverID).Scan(
			&retrieved.OSName, &retrieved.OSVersion, &retrieved.OSType,
			&retrieved.KernelVersion, &retrieved.PackageManager,
			&retrieved.InitSystem, &retrieved.CPUModel, &retrieved.CPUCount,
			&retrieved.MemoryTotalGB, &retrieved.DiskTotalGB, &retrieved.DiskFreeGB)
		if err != nil {
			t.Fatalf("Failed to retrieve server details: %v", err)
		}

		if retrieved.OSName != details.OSName {
			t.Errorf("Expected OS name %s, got %s", details.OSName, retrieved.OSName)
		}
		if retrieved.OSType != details.OSType {
			t.Errorf("Expected OS type %s, got %s", details.OSType, retrieved.OSType)
		}
		if retrieved.CPUCount != details.CPUCount {
			t.Errorf("Expected CPU count %d, got %d", details.CPUCount, retrieved.CPUCount)
		}
	})
}

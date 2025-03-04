package scripts

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"

	_ "github.com/lib/pq"
)

// Common ports and their descriptions
var commonPorts = map[int]string{
	20:    "FTP (Data)",
	21:    "FTP (Control)",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	88:    "Kerberos",
	110:   "POP3",
	123:   "NTP",
	135:   "MSRPC",
	137:   "NetBIOS Name Service",
	138:   "NetBIOS Datagram Service",
	139:   "NetBIOS Session Service",
	143:   "IMAP",
	389:   "LDAP",
	443:   "HTTPS",
	445:   "SMB",
	464:   "Kerberos Change/Set password",
	465:   "SMTP over SSL",
	500:   "ISAKMP/IKE",
	514:   "Syslog",
	587:   "SMTP (Submission)",
	636:   "LDAPS",
	993:   "IMAPS",
	995:   "POP3S",
	1433:  "Microsoft SQL Server",
	1434:  "Microsoft SQL Monitor",
	1521:  "Oracle Database",
	3306:  "MySQL",
	3389:  "RDP",
	5060:  "SIP",
	5222:  "XMPP",
	5432:  "PostgreSQL",
	5985:  "WinRM HTTP",
	5986:  "WinRM HTTPS",
	8080:  "HTTP Alternate",
	8443:  "HTTPS Alternate",
	49152: "Windows RPC",
}

// Port represents an open network port on a server
type Port struct {
	LocalPort   int
	LocalIP     string
	RemotePort  int
	RemoteIP    string
	State       string
	Description string
	ProcessID   int
	ProcessName string
}

func main() {
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

	// Get all servers
	rows, err := db.Query(`
		SELECT id, hostname, os_type 
		FROM servers
	`)
	if err != nil {
		log.Fatalf("Failed to query servers: %v", err)
	}
	defer rows.Close()

	var servers []struct {
		ID       int
		Hostname string
		OSType   string
	}

	for rows.Next() {
		var s struct {
			ID       int
			Hostname string
			OSType   string
		}
		err := rows.Scan(&s.ID, &s.Hostname, &s.OSType)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		servers = append(servers, s)
	}

	log.Printf("Found %d servers", len(servers))

	// Process servers concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent goroutines

	for _, server := range servers {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(server struct {
			ID       int
			Hostname string
			OSType   string
		}) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Generate and insert port information
			err := generateAndInsertPorts(db, server.ID, server.OSType)
			if err != nil {
				log.Printf("Error processing server %s (ID: %d): %v", server.Hostname, server.ID, err)
				return
			}
			log.Printf("Successfully processed server %s (ID: %d)", server.Hostname, server.ID)
		}(server)
	}

	wg.Wait()
	log.Println("Port information update completed")
}

func generateAndInsertPorts(db *sql.DB, serverID int, osType string) error {
	// Delete existing ports for this server
	_, err := db.Exec("DELETE FROM server_ports WHERE server_id = $1", serverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing ports: %w", err)
	}

	// Generate appropriate ports based on OS type
	var ports []Port

	// Common ports for all servers
	ports = append(ports, Port{
		LocalPort:   22,
		LocalIP:     "0.0.0.0",
		State:       "LISTENING",
		Description: commonPorts[22],
		ProcessID:   rand.Intn(1000) + 1,
		ProcessName: "sshd",
	})

	// Add OS-specific ports
	if isWindowsServer(osType) {
		// Windows-specific ports
		ports = append(ports,
			Port{
				LocalPort:   3389,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[3389],
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "TermService",
			},
			Port{
				LocalPort:   445,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[445],
				ProcessID:   4,
				ProcessName: "System",
			},
			Port{
				LocalPort:   135,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[135],
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "svchost.exe",
			},
		)

		// Add some established connections for Windows
		ports = append(ports,
			Port{
				LocalPort:   49152 + rand.Intn(1000),
				LocalIP:     fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255)),
				RemotePort:  443,
				RemoteIP:    fmt.Sprintf("20.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255)),
				State:       "ESTABLISHED",
				Description: "Windows RPC",
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "svchost.exe",
			},
		)
	} else {
		// Linux-specific ports
		ports = append(ports,
			Port{
				LocalPort:   80,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[80],
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "nginx",
			},
			Port{
				LocalPort:   443,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[443],
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "nginx",
			},
		)

		// Add some established connections for Linux
		ports = append(ports,
			Port{
				LocalPort:   32768 + rand.Intn(28000),
				LocalIP:     fmt.Sprintf("10.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255)),
				RemotePort:  443,
				RemoteIP:    fmt.Sprintf("151.101.%d.%d", rand.Intn(255), rand.Intn(255)),
				State:       "ESTABLISHED",
				Description: "Outbound HTTPS",
				ProcessID:   rand.Intn(1000) + 1,
				ProcessName: "curl",
			},
		)
	}

	// Insert the ports
	for _, port := range ports {
		_, err = db.Exec(`
			INSERT INTO server_ports (
				server_id, local_port, local_ip, remote_port, remote_ip,
				state, description, process_id, process_name
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			serverID,
			port.LocalPort,
			port.LocalIP,
			port.RemotePort,
			port.RemoteIP,
			port.State,
			port.Description,
			port.ProcessID,
			port.ProcessName,
		)
		if err != nil {
			return fmt.Errorf("failed to insert port: %w", err)
		}
	}

	return nil
}

func isWindowsServer(osType string) bool {
	return osType == "Windows Server 2012 R2" ||
		osType == "Windows Server 2016" ||
		osType == "Windows Server 2019" ||
		osType == "Windows Server 2022"
}

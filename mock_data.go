package main

import (
	"fmt"
	"time"
)

// Create a map of common ports and their descriptions
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

// Create mock server data
func getMockServers() []Server {
	return []Server{
		{
			ID:             1,
			Hostname:       "win-server-01",
			Port:           5985,
			Region:         "us-east",
			Tags:           []Tag{{Key: "env", Value: "production"}, {Key: "role", Value: "web"}},
			DiscoveryCount: 5,
			LastDiscovery:  time.Now().Add(-24 * time.Hour),
		},
		{
			ID:             2,
			Hostname:       "win-server-02",
			Port:           5985,
			Region:         "us-west",
			Tags:           []Tag{{Key: "env", Value: "staging"}, {Key: "role", Value: "app"}},
			DiscoveryCount: 3,
			LastDiscovery:  time.Now().Add(-48 * time.Hour),
		},
		{
			ID:             3,
			Hostname:       "win-server-03",
			Port:           5985,
			Region:         "eu-central",
			Tags:           []Tag{{Key: "env", Value: "development"}, {Key: "role", Value: "db"}},
			DiscoveryCount: 2,
			LastDiscovery:  time.Now().Add(-72 * time.Hour),
		},
		{
			ID:             4,
			Hostname:       "linux-server-01",
			Port:           22,
			Region:         "ap-south",
			Tags:           []Tag{{Key: "env", Value: "production"}, {Key: "role", Value: "web"}, {Key: "os", Value: "linux"}},
			DiscoveryCount: 4,
			LastDiscovery:  time.Now().Add(-36 * time.Hour),
		},
		{
			ID:             5,
			Hostname:       "linux-server-02",
			Port:           22,
			Region:         "us-east",
			Tags:           []Tag{{Key: "env", Value: "staging"}, {Key: "role", Value: "database"}, {Key: "os", Value: "linux"}},
			DiscoveryCount: 2,
			LastDiscovery:  time.Now().Add(-60 * time.Hour),
		},
	}
}

// Create mock server details
func getMockServerWithDetails(id int) ServerWithDetails {
	// Customize the server details based on the ID
	var hostname string
	region := "us-east"
	isLinux := false

	// Set hostname and determine if it's a Linux server based on ID
	if id >= 4 {
		hostname = fmt.Sprintf("linux-server-%02d", id-3)
		isLinux = true
	} else {
		hostname = fmt.Sprintf("win-server-%02d", id)
	}

	// Different regions based on ID
	switch id {
	case 2:
		region = "us-west"
	case 3:
		region = "eu-central"
	case 4:
		region = "ap-south"
	case 5:
		region = "us-east"
	}

	// Create IP addresses based on the server ID
	ipAddresses := []IPAddress{
		{
			IPAddress: fmt.Sprintf("192.168.1.%d", 100+id),
			InterfaceName: func() string {
				if isLinux {
					return "eth0"
				}
				return "Ethernet"
			}(),
		},
		{
			IPAddress: fmt.Sprintf("10.0.0.%d", 100+id),
			InterfaceName: func() string {
				if isLinux {
					return "ens3"
				}
				return "Internal"
			}(),
		},
	}

	// Create installed software based on the server ID
	installedSoftware := []Software{
		{
			Name:        "Microsoft Windows Server 2019",
			Version:     "10.0.17763",
			InstallDate: "2022-01-01",
		},
		{
			Name:        "Microsoft .NET Framework 4.8",
			Version:     "4.8.03761",
			InstallDate: "2022-01-01",
		},
		{
			Name:        "Microsoft Visual C++ 2015-2019 Redistributable (x64)",
			Version:     "14.29.30139.0",
			InstallDate: "2022-01-01",
		},
	}

	// For Linux servers, use different software
	if isLinux {
		installedSoftware = []Software{
			{
				Name:        "Ubuntu 20.04.4 LTS",
				Version:     "20.04",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "OpenSSH",
				Version:     "8.2p1",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "Python",
				Version:     "3.8.10",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "nginx",
				Version:     "1.18.0",
				InstallDate: "2022-01-15",
			},
		}
	}

	// Create running services based on the server ID
	var runningServices []Service
	if isLinux {
		runningServices = []Service{
			{
				Name:        "ssh.service",
				DisplayName: "OpenSSH Server",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "root",
			},
			{
				Name:        "nginx.service",
				DisplayName: "Nginx Web Server",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "www-data",
			},
			{
				Name:        "cron.service",
				DisplayName: "Regular background program processing daemon",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "root",
			},
			{
				Name:        "systemd-timesyncd.service",
				DisplayName: "Network Time Synchronization",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "systemd-timesync",
			},
			{
				Name:        "ufw.service",
				DisplayName: "Uncomplicated firewall",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "root",
			},
		}
	} else if id == 2 { // Database server
		runningServices = []Service{
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
			{
				Name:        "SQLWriter",
				DisplayName: "SQL Server VSS Writer",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT Service\\SQLWriter",
			},
			{
				Name:        "Winmgmt",
				DisplayName: "Windows Management Instrumentation",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
		}
	} else { // Default Windows services
		runningServices = []Service{
			{
				Name:        "wuauserv",
				DisplayName: "Windows Update",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
			{
				Name:        "LanmanServer",
				DisplayName: "Server",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
			{
				Name:        "W32Time",
				DisplayName: "Windows Time",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\LocalService",
			},
			{
				Name:        "WinRM",
				DisplayName: "Windows Remote Management",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\NetworkService",
			},
			{
				Name:        "Spooler",
				DisplayName: "Print Spooler",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
		}
	}

	// Create different open ports based on server role
	var openPorts []Port

	// Web server
	if id == 1 {
		openPorts = []Port{
			{
				LocalPort:   80,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[80],
				ProcessID:   4,
				ProcessName: "httpd.exe",
			},
			{
				LocalPort:   443,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[443],
				ProcessID:   4,
				ProcessName: "httpd.exe",
			},
			{
				LocalPort:   22,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[22],
				ProcessID:   1024,
				ProcessName: "sshd.exe",
			},
			{
				LocalPort:   5985,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[5985],
				ProcessID:   4,
				ProcessName: "svchost.exe",
			},
			// Add some established connections
			{
				LocalPort:   49152,
				LocalIP:     "192.168.1.101",
				RemotePort:  443,
				RemoteIP:    "20.81.111.85",
				State:       "ESTABLISHED",
				Description: "Windows RPC",
				ProcessID:   980,
				ProcessName: "svchost.exe",
			},
			{
				LocalPort:   49153,
				LocalIP:     "192.168.1.101",
				RemotePort:  80,
				RemoteIP:    "104.18.12.129",
				State:       "ESTABLISHED",
				Description: "Windows RPC",
				ProcessID:   4,
				ProcessName: "System",
			},
		}
	} else if id == 2 { // Database server
		openPorts = []Port{
			{LocalPort: 1433, State: "LISTENING", Description: commonPorts[1433]},
			{LocalPort: 3306, State: "LISTENING", Description: commonPorts[3306]},
			{LocalPort: 5432, State: "LISTENING", Description: commonPorts[5432]},
			{LocalPort: 5985, State: "LISTENING", Description: commonPorts[5985]},
		}
	} else if id == 3 { // Domain controller
		openPorts = []Port{
			{LocalPort: 53, State: "LISTENING", Description: commonPorts[53]},
			{LocalPort: 88, State: "LISTENING", Description: commonPorts[88]},
			{LocalPort: 389, State: "LISTENING", Description: commonPorts[389]},
			{LocalPort: 445, State: "LISTENING", Description: commonPorts[445]},
			{LocalPort: 5985, State: "LISTENING", Description: commonPorts[5985]},
		}
	} else if id == 4 { // Linux web server
		openPorts = []Port{
			{
				LocalPort:   80,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[80],
				ProcessID:   1234,
				ProcessName: "nginx",
			},
			{
				LocalPort:   443,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[443],
				ProcessID:   1234,
				ProcessName: "nginx",
			},
			{
				LocalPort:   22,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[22],
				ProcessID:   987,
				ProcessName: "sshd",
			},
			// Add some established connections
			{
				LocalPort:   56789,
				LocalIP:     "192.168.1.104",
				RemotePort:  443,
				RemoteIP:    "151.101.1.69",
				State:       "ESTABLISHED",
				Description: "Ephemeral",
				ProcessID:   2345,
				ProcessName: "curl",
			},
			{
				LocalPort:   45678,
				LocalIP:     "192.168.1.104",
				RemotePort:  80,
				RemoteIP:    "172.217.167.78",
				State:       "ESTABLISHED",
				Description: "Ephemeral",
				ProcessID:   3456,
				ProcessName: "wget",
			},
		}
	} else if id == 5 { // Linux database server
		openPorts = []Port{
			{
				LocalPort:   3306,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[3306],
				ProcessID:   1122,
				ProcessName: "mysqld",
			},
			{
				LocalPort:   5432,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[5432],
				ProcessID:   1133,
				ProcessName: "postgres",
			},
			{
				LocalPort:   22,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[22],
				ProcessID:   987,
				ProcessName: "sshd",
			},
			// Add some established connections
			{
				LocalPort:   56790,
				LocalIP:     "192.168.1.105",
				RemotePort:  3306,
				RemoteIP:    "10.0.0.15",
				State:       "ESTABLISHED",
				Description: "Ephemeral",
				ProcessID:   1122,
				ProcessName: "mysqld",
			},
		}
	} else { // Default
		openPorts = []Port{
			{LocalPort: 135, State: "LISTENING", Description: commonPorts[135]},
			{LocalPort: 139, State: "LISTENING", Description: commonPorts[139]},
			{LocalPort: 445, State: "LISTENING", Description: commonPorts[445]},
			{LocalPort: 3389, State: "LISTENING", Description: commonPorts[3389]},
			{LocalPort: 5985, State: "LISTENING", Description: commonPorts[5985]},
		}
	}

	var port int
	if id == 4 || id == 5 {
		port = 22 // SSH port for Linux servers
	} else {
		port = 5985 // WinRM port for Windows servers
	}

	var tags []Tag
	if id >= 4 {
		tags = []Tag{{Key: "env", Value: "production"}, {Key: "role", Value: "web"}, {Key: "os", Value: "linux"}}
	} else {
		tags = []Tag{{Key: "env", Value: "production"}, {Key: "role", Value: "web"}}
	}

	var osName, osVersion string
	if id >= 4 {
		osName = "Ubuntu 20.04.4 LTS"
		osVersion = "20.04"
	} else {
		osName = "Windows Server 2019"
		osVersion = "10.0.17763"
	}

	return ServerWithDetails{
		ID:                id,
		Hostname:          hostname,
		Port:              port,
		Region:            region,
		Tags:              tags,
		OSName:            osName,
		OSVersion:         osVersion,
		CPUModel:          "Intel(R) Xeon(R) CPU E5-2670 0 @ 2.60GHz",
		CPUCount:          4,
		MemoryTotalGB:     16.0,
		DiskTotalGB:       256.0,
		DiskFreeGB:        128.0,
		LastBootTime:      time.Now().Add(-7 * 24 * time.Hour),
		IPAddresses:       ipAddresses,
		InstalledSoftware: installedSoftware,
		RunningServices:   runningServices,
		OpenPorts:         openPorts,
		DiscoveryCount:    id,
		LastDiscovery:     time.Now().Add(-24 * time.Hour),
	}
}

// Create mock server discoveries
func getMockServerDiscoveries(serverID int) []DiscoveryDetails {
	// Create 3-5 mock discoveries for this server
	count := 3 + (serverID % 3)
	discoveries := make([]DiscoveryDetails, count)

	for i := 0; i < count; i++ {
		// Create a discovery with ID based on server ID and index
		discoveryID := serverID*10 + i + 1

		// Create the discovery with different timestamps
		timeOffset := time.Duration(-(i+1)*24) * time.Hour
		startTime := time.Now().Add(timeOffset)
		endTime := startTime.Add(30 * time.Minute)

		// Set success status (make some discoveries fail for realism)
		success := true
		message := "Discovery completed successfully"
		if i == count-1 && serverID%2 == 0 {
			success = false
			message = "Connection timeout"
		}

		// Create the discovery details
		discovery := getMockDiscoveryDetails(discoveryID)
		discovery.ServerID = serverID
		discovery.Success = success
		discovery.Message = message
		discovery.StartTime = startTime
		discovery.EndTime = endTime

		discoveries[i] = discovery
	}

	return discoveries
}

// Helper function to get output path
func getOutputPath(success bool, serverID, index int) string {
	if success {
		return fmt.Sprintf("/discovery_results/win-server-%02d/%d", serverID, index+1)
	}
	return ""
}

// Helper function to get region
func getRegion(serverID int) string {
	if serverID == 2 {
		return "us-west"
	} else if serverID == 3 {
		return "eu-central"
	}
	return "us-east"
}

// Create mock discovery details
func getMockDiscoveryDetails(id int) DiscoveryDetails {
	// Get server ID (assuming discovery ID maps to server ID for simplicity)
	serverID := (id % 5) + 1
	if serverID == 0 {
		serverID = 5
	}

	// Get server details
	var serverHostname string
	var serverPort int
	var serverRegion string
	isLinux := serverID >= 4

	if isLinux {
		serverHostname = fmt.Sprintf("linux-server-%02d", serverID-3)
		serverPort = 22
	} else {
		serverHostname = fmt.Sprintf("win-server-%02d", serverID)
		serverPort = 5985
	}

	// Set region based on server ID
	switch serverID {
	case 2:
		serverRegion = "us-west"
	case 3:
		serverRegion = "eu-central"
	case 4:
		serverRegion = "ap-south"
	case 5:
		serverRegion = "us-east"
	default:
		serverRegion = "us-east"
	}

	// Create IP addresses based on the server ID
	ipAddresses := []IPAddress{
		{
			IPAddress: fmt.Sprintf("192.168.1.%d", 100+serverID),
			InterfaceName: func() string {
				if isLinux {
					return "eth0"
				}
				return "Ethernet"
			}(),
		},
		{
			IPAddress: fmt.Sprintf("10.0.0.%d", 100+serverID),
			InterfaceName: func() string {
				if isLinux {
					return "ens3"
				}
				return "Internal"
			}(),
		},
	}

	// Create installed software based on the server ID
	var installedSoftware []Software
	if isLinux {
		installedSoftware = []Software{
			{
				Name:        "Ubuntu 20.04.4 LTS",
				Version:     "20.04",
				Vendor:      "Canonical",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "OpenSSH",
				Version:     "8.2p1",
				Vendor:      "OpenBSD",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "Python",
				Version:     "3.8.10",
				Vendor:      "Python Software Foundation",
				InstallDate: "2022-01-01",
			},
		}

		// Add web server software for the web server
		if serverID == 4 {
			installedSoftware = append(installedSoftware, Software{
				Name:        "nginx",
				Version:     "1.18.0",
				Vendor:      "Nginx, Inc.",
				InstallDate: "2022-01-15",
			})
		}

		// Add database software for the database server
		if serverID == 5 {
			installedSoftware = append(installedSoftware,
				Software{
					Name:        "MySQL Server",
					Version:     "8.0.28",
					Vendor:      "Oracle Corporation",
					InstallDate: "2022-01-15",
				},
				Software{
					Name:        "PostgreSQL",
					Version:     "14.2",
					Vendor:      "PostgreSQL Global Development Group",
					InstallDate: "2022-02-01",
				})
		}
	} else {
		installedSoftware = []Software{
			{
				Name:        "Microsoft Windows Server 2019",
				Version:     "10.0.17763",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "Microsoft .NET Framework 4.8",
				Version:     "4.8.03761",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2022-01-01",
			},
			{
				Name:        "Microsoft Visual C++ 2015-2019 Redistributable (x64)",
				Version:     "14.29.30139.0",
				Vendor:      "Microsoft Corporation",
				InstallDate: "2022-01-01",
			},
		}
	}

	// Create running services based on the server ID
	var runningServices []Service
	if isLinux {
		runningServices = []Service{
			{
				Name:        "ssh.service",
				DisplayName: "OpenSSH Server",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "root",
			},
			{
				Name:        "cron.service",
				DisplayName: "Regular background program processing daemon",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "root",
			},
			{
				Name:        "systemd-timesyncd.service",
				DisplayName: "Network Time Synchronization",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "systemd-timesync",
			},
		}

		// Add web server services for the web server
		if serverID == 4 {
			runningServices = append(runningServices, Service{
				Name:        "nginx.service",
				DisplayName: "Nginx Web Server",
				Status:      "Running",
				StartType:   "enabled",
				Account:     "www-data",
			})
		}

		// Add database services for the database server
		if serverID == 5 {
			runningServices = append(runningServices,
				Service{
					Name:        "mysql.service",
					DisplayName: "MySQL Database Server",
					Status:      "Running",
					StartType:   "enabled",
					Account:     "mysql",
				},
				Service{
					Name:        "postgresql.service",
					DisplayName: "PostgreSQL Database Server",
					Status:      "Running",
					StartType:   "enabled",
					Account:     "postgres",
				})
		}
	} else {
		// Windows services
		runningServices = []Service{
			{
				Name:        "wuauserv",
				DisplayName: "Windows Update",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
			{
				Name:        "LanmanServer",
				DisplayName: "Server",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "LocalSystem",
			},
			{
				Name:        "W32Time",
				DisplayName: "Windows Time",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\LocalService",
			},
			{
				Name:        "WinRM",
				DisplayName: "Windows Remote Management",
				Status:      "Running",
				StartType:   "Automatic",
				Account:     "NT AUTHORITY\\NetworkService",
			},
		}
	}

	// Create open ports based on the server ID
	var openPorts []Port
	if isLinux {
		// Basic Linux ports
		openPorts = []Port{
			{
				LocalPort:   22,
				LocalIP:     "0.0.0.0",
				State:       "LISTENING",
				Description: commonPorts[22],
				ProcessID:   987,
				ProcessName: "sshd",
			},
		}

		// Add web server ports for the web server
		if serverID == 4 {
			openPorts = append(openPorts,
				Port{
					LocalPort:   80,
					LocalIP:     "0.0.0.0",
					State:       "LISTENING",
					Description: commonPorts[80],
					ProcessID:   1234,
					ProcessName: "nginx",
				},
				Port{
					LocalPort:   443,
					LocalIP:     "0.0.0.0",
					State:       "LISTENING",
					Description: commonPorts[443],
					ProcessID:   1234,
					ProcessName: "nginx",
				})
		}

		// Add database ports for the database server
		if serverID == 5 {
			openPorts = append(openPorts,
				Port{
					LocalPort:   3306,
					LocalIP:     "0.0.0.0",
					State:       "LISTENING",
					Description: commonPorts[3306],
					ProcessID:   1122,
					ProcessName: "mysqld",
				},
				Port{
					LocalPort:   5432,
					LocalIP:     "0.0.0.0",
					State:       "LISTENING",
					Description: commonPorts[5432],
					ProcessID:   1133,
					ProcessName: "postgres",
				})
		}
	} else {
		// Windows ports
		openPorts = []Port{
			{LocalPort: 135, State: "LISTENING", Description: commonPorts[135]},
			{LocalPort: 139, State: "LISTENING", Description: commonPorts[139]},
			{LocalPort: 445, State: "LISTENING", Description: commonPorts[445]},
			{LocalPort: 3389, State: "LISTENING", Description: commonPorts[3389]},
			{LocalPort: 5985, State: "LISTENING", Description: commonPorts[5985]},
		}
	}

	// Set OS name and version based on server type
	var osName, osVersion string
	if isLinux {
		osName = "Ubuntu 20.04.4 LTS"
		osVersion = "20.04"
	} else {
		osName = "Windows Server 2019"
		osVersion = "10.0.17763"
	}

	return DiscoveryDetails{
		ID:                id,
		ServerID:          serverID,
		ServerHostname:    serverHostname,
		ServerPort:        serverPort,
		ServerRegion:      serverRegion,
		Success:           true,
		Message:           "Discovery completed successfully",
		StartTime:         time.Now().Add(-24 * time.Hour),
		EndTime:           time.Now().Add(-24*time.Hour + 30*time.Minute),
		OSName:            osName,
		OSVersion:         osVersion,
		CPUModel:          "Intel(R) Xeon(R) CPU E5-2670 0 @ 2.60GHz",
		CPUCount:          4,
		MemoryTotalGB:     16.0,
		DiskTotalGB:       256.0,
		DiskFreeGB:        128.0,
		LastBootTime:      time.Now().Add(-7 * 24 * time.Hour),
		IPAddresses:       ipAddresses,
		InstalledSoftware: installedSoftware,
		RunningServices:   runningServices,
		OpenPorts:         openPorts,
	}
}

// Create mock query results
func getMockQueryResults() []map[string]interface{} {
	return []map[string]interface{}{
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
}

// Create mock system stats
func getMockStats() map[string]interface{} {
	// Get the number of mock servers we're actually providing
	mockServers := getMockServers()
	serverCount := len(mockServers)

	// Calculate discovery count (assuming each server has 2-5 discoveries)
	discoveryCount := 0
	for _, server := range mockServers {
		discoveryCount += server.DiscoveryCount
	}

	// Calculate success rate (90-95% success is realistic)
	successCount := int(float64(discoveryCount) * 0.93)
	successRate := float64(successCount) / float64(discoveryCount) * 100

	// Create region distribution based on our mock servers
	regions := make(map[string]int)
	for _, server := range mockServers {
		regions[server.Region] = regions[server.Region] + 1
	}

	// Debug log
	fmt.Printf("Mock stats: %+v\n", map[string]interface{}{
		"serverCount":    serverCount,
		"discoveryCount": discoveryCount,
		"successRate":    successRate,
		"regions":        regions,
	})

	return map[string]interface{}{
		"serverCount":    serverCount,
		"discoveryCount": discoveryCount,
		"successRate":    successRate,
		"regions":        regions,
		"lastDiscovery":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"servers":        mockServers,
		"recentDiscoveries": []map[string]interface{}{
			{
				"id":             1,
				"serverHostname": "win-server-01",
				"success":        true,
				"endTime":        time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":             2,
				"serverHostname": "win-server-02",
				"success":        true,
				"endTime":        time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":             3,
				"serverHostname": "linux-server-01",
				"success":        true,
				"endTime":        time.Now().Add(-36 * time.Hour).Format(time.RFC3339),
			},
		},
	}
}

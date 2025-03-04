// Package models contains the server discovery models
package models

import "time"

// Service represents a running service on a server
type Service struct {
	ID          int       `json:"id" db:"id"`
	ServerID    int       `json:"server_id" db:"server_id"`
	Name        string    `json:"name" db:"service_name"`
	Status      string    `json:"status" db:"service_status"`
	Description string    `json:"description" db:"service_description"`
	Port        int       `json:"port" db:"port"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ServerDetails represents detailed information about a server
type ServerDetails struct {
	ID                int            `json:"id" db:"id"`
	Hostname          string         `json:"hostname" db:"hostname"`
	IP                string         `json:"ip" db:"ip"`
	OSType            string         `json:"os_type" db:"os_type"`
	Status            string         `json:"status" db:"status"`
	LastChecked       time.Time      `json:"last_checked" db:"last_checked"`
	Region            string         `json:"region" db:"region"`
	OSName            string         `json:"os_name" db:"os_name"`
	OSVersion         string         `json:"os_version" db:"os_version"`
	CPUModel          string         `json:"cpu_model" db:"cpu_model"`
	CPUCount          int            `json:"cpu_count" db:"cpu_count"`
	MemoryTotalGB     float64        `json:"memory_total_gb" db:"memory_total_gb"`
	DiskTotalGB       float64        `json:"disk_total_gb" db:"disk_total_gb"`
	DiskFreeGB        float64        `json:"disk_free_gb" db:"disk_free_gb"`
	LastBootTime      time.Time      `json:"last_boot_time" db:"last_boot_time"`
	Metrics           *ServerMetrics `json:"metrics,omitempty"`
	Services          []Service      `json:"services,omitempty"`
	IPAddresses       []IPAddress    `json:"ip_addresses,omitempty"`
	OpenPorts         []Port         `json:"open_ports,omitempty"`
	Filesystems       []Filesystem   `json:"filesystems,omitempty"`
	InstalledSoftware []Software     `json:"installed_software,omitempty" db:"installed_software"`
	Tags              []Tag          `json:"tags,omitempty"`
}

// ServerMetrics represents server performance metrics
type ServerMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryTotal  int64   `json:"memory_total"`
	MemoryUsed   int64   `json:"memory_used"`
	DiskTotal    int64   `json:"disk_total"`
	DiskUsed     int64   `json:"disk_used"`
	LoadAverage  float64 `json:"load_average"`
	ProcessCount int     `json:"process_count"`
}

// Tag represents a key-value tag for a server
type Tag struct {
	ID        int       `json:"id" db:"id"`
	ServerID  int       `json:"server_id" db:"server_id"`
	TagName   string    `json:"tag_name" db:"tag_name"`
	TagValue  string    `json:"tag_value" db:"tag_value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ServerWithDetails represents a server with its details
type ServerWithDetails struct {
	ID          int            `json:"id" db:"id"`
	Hostname    string         `json:"hostname" db:"hostname"`
	IP          string         `json:"ip" db:"ip"`
	OSType      string         `json:"os_type" db:"os_type"`
	Region      string         `json:"region" db:"region"`
	Status      string         `json:"status" db:"status"`
	LastChecked time.Time      `json:"last_checked" db:"last_checked"`
	Metrics     *ServerMetrics `json:"metrics,omitempty"`
	Tags        []Tag          `json:"tags,omitempty"`
}

// IPAddress represents an IP address and its interface
type IPAddress struct {
	IPAddress     string `json:"ip_address"`
	InterfaceName string `json:"interface_name"`
}

// Port represents an open network port discovered during server scanning.
// This is different from the Port field in the Service struct:
// - Service.Port represents a port number that a service is configured to use
// - Port struct represents an actually discovered open network port with its full details
type Port struct {
	// LocalPort is the port number on the local machine that was found to be open
	LocalPort int `json:"local_port" db:"local_port"`

	// LocalIP is the IP address on the local machine that the port is bound to
	LocalIP string `json:"local_ip" db:"local_ip"`

	// RemotePort is the port number on the remote end of an established connection
	// This may be empty for listening ports with no current connections
	RemotePort int `json:"remote_port" db:"remote_port"`

	// RemoteIP is the IP address of the remote end of an established connection
	// This may be empty for listening ports with no current connections
	RemoteIP string `json:"remote_ip" db:"remote_ip"`

	// State indicates the current state of the port (e.g., "LISTENING", "ESTABLISHED")
	State string `json:"state" db:"state"`

	// Description provides additional information about what is using this port
	Description string `json:"description" db:"description"`

	// ProcessID is the ID of the process that has this port open
	ProcessID *int `json:"process_id" db:"process_id"`

	// ProcessName is the name of the process that has this port open
	ProcessName string `json:"process_name" db:"process_name"`
}

// Software represents installed software
type Software struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	InstallDate string `json:"install_date"`
}

// Filesystem represents a mounted filesystem
type Filesystem struct {
	MountPoint  string  `json:"mount_point"`
	Device      string  `json:"device"`
	FSType      string  `json:"fs_type"`
	TotalBytes  int64   `json:"total_bytes"`
	UsedBytes   int64   `json:"used_bytes"`
	FreeBytes   int64   `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
	TotalInodes int64   `json:"total_inodes,omitempty"`
	UsedInodes  int64   `json:"used_inodes,omitempty"`
	FreeInodes  int64   `json:"free_inodes,omitempty"`
}

// Config represents the main configuration for the application
type Config struct {
	Database         DatabaseConfig `json:"database"`
	Server           ServerConfig   `json:"server"`
	SSH              SSHConfig      `json:"ssh"`
	API              APIConfig      `json:"api"`
	PowerShellScript string         `json:"powershell_script"`
	OutputDir        string         `json:"output_dir"`
	Concurrency      int            `json:"concurrency"`
	Servers          []ServerConfig `json:"servers"`
	DatabaseConfig   DatabaseConfig `json:"database_config"`
	SkipCertVerify   bool           `json:"skip_cert_verify"`
	Timeout          int            `json:"timeout"`
	CacheTTL         int            `json:"cache_ttl"`
	BatchSize        int            `json:"batch_size"`
	MetricsPort      int            `json:"metrics_port"`
	TracingEndpoint  string         `json:"tracing_endpoint"`
}

// APIConfig represents API server configuration
type APIConfig struct {
	Port            int           `json:"port"`
	AllowedOrigins  string        `json:"allowed_origins"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
	Enabled  bool   `json:"enabled"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	ID             int    `json:"id"`
	Host           string `json:"host"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PrivateKeyPath string `json:"private_key_path"`
	UseWinRM       bool   `json:"use_winrm"`
	WinRMPort      int    `json:"winrm_port"`
	WinRMHTTPS     bool   `json:"winrm_https"`
	WinRMInsecure  bool   `json:"winrm_insecure"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Region         string `json:"region"`
}

// SSHConfig represents SSH connection configuration
type SSHConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PrivateKeyPath string `json:"private_key_path"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// DiscoveryResult represents the result of a server discovery operation
type DiscoveryResult struct {
	ID          int       `json:"id"`
	ServerID    int       `json:"server_id"`
	Server      string    `json:"server"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	LastChecked time.Time `json:"last_checked"`
	OutputPath  string    `json:"output_path,omitempty"`
	Error       string    `json:"error,omitempty"`
	Region      string    `json:"region,omitempty"`
}

// DiscoveryRequest represents a request to discover a server
type DiscoveryRequest struct {
	ServerID int    `json:"server_id"`
	IP       string `json:"ip"`
	OSType   string `json:"os_type"`
}

// DiscoveryResponse represents the response from a discovery operation
type DiscoveryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// DiscoveryStats represents statistics about the discovery process
type DiscoveryStats struct {
	TotalServers       int     `json:"total_servers"`
	ProcessedServers   int     `json:"processed_servers"`
	SuccessfulScans    int     `json:"successful_scans"`
	FailedScans        int     `json:"failed_scans"`
	AverageTimePerScan float64 `json:"average_time_per_scan"`
	StartTime          string  `json:"start_time"`
	EndTime            string  `json:"end_time"`
}

// User represents an active user on the server
type User struct {
	Username  string        `json:"username"`
	Terminal  string        `json:"terminal"`
	Host      string        `json:"host"`
	LoginTime time.Time     `json:"login_time"`
	IDLE      time.Duration `json:"idle"`
	CPU       string        `json:"cpu"`
	What      string        `json:"what"`
}

// LoadInfo represents system load information
type LoadInfo struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// NetworkInterface represents a network interface on the server
type NetworkInterface struct {
	Name        string   `json:"name"`
	MACAddress  string   `json:"mac_address"`
	IPAddresses []string `json:"ip_addresses"`
	State       string   `json:"state"`
	Speed       int      `json:"speed"`
	MTU         int      `json:"mtu"`
}

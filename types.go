package main

import (
	"time"
)

// Server represents a server in the system
type Server struct {
	ID             int       `json:"id"`
	Hostname       string    `json:"hostname"`
	Port           int       `json:"port"`
	Region         string    `json:"region"`
	Tags           []Tag     `json:"tags"`
	DiscoveryCount int       `json:"discovery_count"`
	LastDiscovery  time.Time `json:"last_discovery"`
}

// ServerWithDetails represents a server with additional details
type ServerWithDetails struct {
	ID                int         `json:"id"`
	Hostname          string      `json:"hostname"`
	Port              int         `json:"port"`
	Region            string      `json:"region"`
	Tags              []Tag       `json:"tags"`
	OSName            string      `json:"os_name"`
	OSVersion         string      `json:"os_version"`
	CPUModel          string      `json:"cpu_model"`
	CPUCount          int         `json:"cpu_count"`
	MemoryTotalGB     float64     `json:"memory_total_gb"`
	DiskTotalGB       float64     `json:"disk_total_gb"`
	DiskFreeGB        float64     `json:"disk_free_gb"`
	LastBootTime      time.Time   `json:"last_boot_time"`
	IPAddresses       []IPAddress `json:"ip_addresses"`
	InstalledSoftware []Software  `json:"installed_software"`
	RunningServices   []Service   `json:"running_services"`
	OpenPorts         []Port      `json:"open_ports"`
	DiscoveryCount    int         `json:"discovery_count"`
	LastDiscovery     time.Time   `json:"last_discovery"`
}

// DiscoveryDetails represents the detailed results of a server discovery
type DiscoveryDetails struct {
	ID                int         `json:"id"`
	ServerID          int         `json:"server_id"`
	ServerHostname    string      `json:"server_hostname"`
	ServerPort        int         `json:"server_port"`
	ServerRegion      string      `json:"server_region"`
	Success           bool        `json:"success"`
	Message           string      `json:"message"`
	StartTime         time.Time   `json:"start_time"`
	EndTime           time.Time   `json:"end_time"`
	OSName            string      `json:"os_name"`
	OSVersion         string      `json:"os_version"`
	CPUModel          string      `json:"cpu_model"`
	CPUCount          int         `json:"cpu_count"`
	MemoryTotalGB     float64     `json:"memory_total_gb"`
	DiskTotalGB       float64     `json:"disk_total_gb"`
	DiskFreeGB        float64     `json:"disk_free_gb"`
	LastBootTime      time.Time   `json:"last_boot_time"`
	IPAddresses       []IPAddress `json:"ip_addresses"`
	InstalledSoftware []Software  `json:"installed_software"`
	RunningServices   []Service   `json:"running_services"`
	OpenPorts         []Port      `json:"open_ports"`
	Error             string      `json:"error,omitempty"`
	OutputPath        string      `json:"output_path,omitempty"`
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

// Service represents a running service on a server
type Service struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	StartType   string `json:"start_type"`
	Account     string `json:"account"`
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

// Tag represents a key-value tag for a server
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Result of a discovery operation
type DiscoveryResult struct {
	ID          int       `json:"id,omitempty"`
	DiscoveryID int       `json:"discovery_id,omitempty"`
	Server      string    `json:"server"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	OutputPath  string    `json:"output_path,omitempty"`
	Error       string    `json:"error,omitempty"`
	Region      string    `json:"region,omitempty"`
}

// Configuration structure
type Config struct {
	Servers             []ServerConfig  `json:"servers"`
	Concurrency         int             `json:"concurrency"`
	Timeout             int             `json:"timeout_seconds"`
	OutputDir           string          `json:"output_directory"`
	CacheTTL            int             `json:"cache_ttl_minutes"`
	BatchSize           int             `json:"batch_size"`
	SkipCertVerify      bool            `json:"skip_cert_verify"`
	PowerShellScript    string          `json:"powershell_script"`
	MetricsPort         int             `json:"metrics_port"`
	TracingEndpoint     string          `json:"tracing_endpoint"`
	DatabaseConfig      DatabaseConfig  `json:"database_config"`
	MaxRetries          int             `json:"max_retries"`
	RetryBackoffSeconds int             `json:"retry_backoff_seconds"`
	ConnectionPoolSize  int             `json:"connection_pool_size"`
	IdleTimeout         int             `json:"idle_timeout_minutes"`
	ResourceThresholds  ResourceConfig  `json:"resource_thresholds"`
	APIServer           APIServerConfig `json:"api_server"`
	LinuxConfig         LinuxConfig     `json:"linux_config"`
	WinRMConfig         WinRMConfig     `json:"winrm_config"`
	ServerPort          int             `json:"server_port"`
}

// Configuration for each server
type ServerConfig struct {
	ID       int    `json:"id"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseHTTPS bool   `json:"use_https"`
	Region   string `json:"region"`
	Tags     []Tag  `json:"tags"`
}

// WinRM configuration
type WinRMConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Linux SSH configuration
type LinuxConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	KeyFile  string `json:"key_file"`
}

// Resource configuration
type ResourceConfig struct {
	CPUThreshold    float64 `json:"cpu_threshold"`
	MemoryThreshold float64 `json:"memory_threshold"`
	NetworkLimit    int64   `json:"network_limit_mbps"`
}

// API Server configuration
type APIServerConfig struct {
	Port            int    `json:"port"`
	AllowedOrigins  string `json:"allowed_origins"`
	ReadTimeout     int    `json:"read_timeout"`
	WriteTimeout    int    `json:"write_timeout"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

// Database configuration
type DatabaseConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ServerDetails represents the detailed information about a server
type ServerDetails struct {
	OSName            string      `json:"os_name"`
	OSVersion         string      `json:"os_version"`
	CPUModel          string      `json:"cpu_model"`
	CPUCount          int         `json:"cpu_count"`
	MemoryTotalGB     float64     `json:"memory_total_gb"`
	DiskTotalGB       float64     `json:"disk_total_gb"`
	DiskFreeGB        float64     `json:"disk_free_gb"`
	LastBootTime      time.Time   `json:"last_boot_time"`
	IPAddresses       []IPAddress `json:"ip_addresses"`
	InstalledSoftware []Software  `json:"installed_software"`
	RunningServices   []Service   `json:"running_services"`
	OpenPorts         []Port      `json:"open_ports"`
	Tags              []Tag       `json:"tags"`
}

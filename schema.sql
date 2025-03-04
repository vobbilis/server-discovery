-- Create schema for server discovery
CREATE SCHEMA IF NOT EXISTS server_discovery;

-- Create servers table
CREATE TABLE IF NOT EXISTS server_discovery.servers (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    region VARCHAR(50),
    ip VARCHAR(50) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    last_checked TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hostname),
    os_type VARCHAR(50) DEFAULT 'windows'
);

-- Create server_metrics table
CREATE TABLE IF NOT EXISTS server_discovery.server_metrics (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
    cpu_usage FLOAT,
    memory_total BIGINT,
    memory_used BIGINT,
    disk_total BIGINT,
    disk_used BIGINT,
    load_average FLOAT,
    process_count INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create discovery_results table
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

-- Create server_tags table
CREATE TABLE IF NOT EXISTS server_discovery.server_tags (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
    tag_name VARCHAR(100) NOT NULL,
    tag_value VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create ip_addresses table
CREATE TABLE IF NOT EXISTS server_discovery.ip_addresses (
    id SERIAL PRIMARY KEY,
    discovery_id INTEGER REFERENCES server_discovery.discovery_results(id) ON DELETE CASCADE,
    ip_address VARCHAR(50) NOT NULL,
    interface_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create installed_software table
CREATE TABLE IF NOT EXISTS server_discovery.installed_software (
    id SERIAL PRIMARY KEY,
    discovery_id INTEGER REFERENCES server_discovery.discovery_results(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(100),
    install_date VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create running_services table
CREATE TABLE IF NOT EXISTS server_discovery.running_services (
    id SERIAL PRIMARY KEY,
    discovery_id INTEGER REFERENCES server_discovery.discovery_results(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    status VARCHAR(50),
    start_mode VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create server_services table
-- This table stores the configured services and their intended listening ports.
-- The port column here represents the port number that a service is configured
-- to use, which may or may not be actually open and listening.
CREATE TABLE IF NOT EXISTS server_discovery.server_services (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    service_status VARCHAR(50) NOT NULL,
    service_description TEXT,
    port INTEGER, -- The configured listening port for this service
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create open_ports table
-- This table stores the actual network ports discovered during server scanning.
-- These are the ports that were found to be actually open and listening,
-- regardless of what services are configured to use them.
CREATE TABLE IF NOT EXISTS server_discovery.open_ports (
    id SERIAL PRIMARY KEY,
    discovery_id INTEGER REFERENCES server_discovery.discovery_results(id) ON DELETE CASCADE,
    local_port INTEGER NOT NULL,      -- The local port number that is open
    local_ip VARCHAR(50),             -- The local IP address the port is bound to
    remote_port INTEGER,              -- The remote port for established connections
    remote_ip VARCHAR(50),            -- The remote IP for established connections
    state VARCHAR(50),                -- The state of the port (e.g., LISTENING, ESTABLISHED)
    description VARCHAR(255),         -- Description of what's using this port
    process_id INTEGER,               -- The ID of the process using this port
    process_name VARCHAR(255),        -- The name of the process using this port
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create server_details table for storing PowerShell discovery results
CREATE TABLE IF NOT EXISTS server_discovery.server_details (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES server_discovery.servers(id),
    discovery_id INTEGER REFERENCES server_discovery.discovery_results(id),
    os_name VARCHAR(255),
    os_version VARCHAR(100),
    cpu_model VARCHAR(255),
    cpu_count INTEGER,
    memory_total_gb NUMERIC(10,2),
    disk_total_gb NUMERIC(10,2),
    disk_free_gb NUMERIC(10,2),
    last_boot_time TIMESTAMP WITH TIME ZONE,
    ip_addresses JSONB,
    installed_software JSONB,
    running_services JSONB,
    open_ports JSONB,  -- Snapshot of open ports at discovery time
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    kernel_version VARCHAR(100),
    package_manager VARCHAR(50),
    init_system VARCHAR(50),
    selinux_status VARCHAR(50),
    firewall_status VARCHAR(50),
    active_users JSONB,
    system_load JSONB,
    network_interfaces JSONB,
    mounted_filesystems JSONB
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_servers_hostname ON server_discovery.servers(hostname);
CREATE INDEX IF NOT EXISTS idx_servers_region ON server_discovery.servers(region);
CREATE INDEX IF NOT EXISTS idx_discovery_results_server_id ON server_discovery.discovery_results(server_id);
CREATE INDEX IF NOT EXISTS idx_discovery_results_success ON server_discovery.discovery_results(success);
CREATE INDEX IF NOT EXISTS idx_server_tags_server_id ON server_discovery.server_tags(server_id);
CREATE INDEX IF NOT EXISTS idx_server_tags_tag_name ON server_discovery.server_tags(tag_name);
CREATE INDEX IF NOT EXISTS idx_server_details_server_id ON server_discovery.server_details(server_id);
CREATE INDEX IF NOT EXISTS idx_ip_addresses_discovery_id ON server_discovery.ip_addresses(discovery_id);
CREATE INDEX IF NOT EXISTS idx_installed_software_discovery_id ON server_discovery.installed_software(discovery_id);
CREATE INDEX IF NOT EXISTS idx_running_services_discovery_id ON server_discovery.running_services(discovery_id);
CREATE INDEX IF NOT EXISTS idx_open_ports_discovery_id ON server_discovery.open_ports(discovery_id);
CREATE INDEX IF NOT EXISTS idx_open_ports_local_port ON server_discovery.open_ports(local_port);
CREATE INDEX IF NOT EXISTS idx_open_ports_process_name ON server_discovery.open_ports(process_name);
CREATE INDEX IF NOT EXISTS idx_open_ports_remote_ip ON server_discovery.open_ports(remote_ip);
CREATE INDEX IF NOT EXISTS idx_servers_os_type ON server_discovery.servers(os_type);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION server_discovery.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for servers table
CREATE TRIGGER update_servers_updated_at
BEFORE UPDATE ON server_discovery.servers
FOR EACH ROW
EXECUTE FUNCTION server_discovery.update_updated_at_column();

-- Create table for SSH keys
CREATE TABLE IF NOT EXISTS server_discovery.ssh_keys (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES server_discovery.servers(id) ON DELETE CASCADE,
    key_type VARCHAR(50) NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(255) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(server_id, fingerprint)
);

-- Create index on ssh_keys
CREATE INDEX IF NOT EXISTS idx_ssh_keys_server_id ON server_discovery.ssh_keys(server_id);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON server_discovery.ssh_keys(fingerprint); 
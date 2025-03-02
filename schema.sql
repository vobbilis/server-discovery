-- Create schema for server discovery
CREATE SCHEMA IF NOT EXISTS server_discovery;

-- Create servers table
CREATE TABLE IF NOT EXISTS server_discovery.servers (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 5985,
    region VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(hostname, port)
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
    key VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
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

-- Create open_ports table
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
    open_ports JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_servers_hostname ON server_discovery.servers(hostname);
CREATE INDEX IF NOT EXISTS idx_servers_region ON server_discovery.servers(region);
CREATE INDEX IF NOT EXISTS idx_discovery_results_server_id ON server_discovery.discovery_results(server_id);
CREATE INDEX IF NOT EXISTS idx_discovery_results_success ON server_discovery.discovery_results(success);
CREATE INDEX IF NOT EXISTS idx_server_tags_server_id ON server_discovery.server_tags(server_id);
CREATE INDEX IF NOT EXISTS idx_server_tags_key ON server_discovery.server_tags(key);
CREATE INDEX IF NOT EXISTS idx_server_details_server_id ON server_discovery.server_details(server_id);
CREATE INDEX IF NOT EXISTS idx_ip_addresses_discovery_id ON server_discovery.ip_addresses(discovery_id);
CREATE INDEX IF NOT EXISTS idx_installed_software_discovery_id ON server_discovery.installed_software(discovery_id);
CREATE INDEX IF NOT EXISTS idx_running_services_discovery_id ON server_discovery.running_services(discovery_id);
CREATE INDEX IF NOT EXISTS idx_open_ports_discovery_id ON server_discovery.open_ports(discovery_id);
CREATE INDEX IF NOT EXISTS idx_open_ports_local_port ON server_discovery.open_ports(local_port);
CREATE INDEX IF NOT EXISTS idx_open_ports_process_name ON server_discovery.open_ports(process_name);
CREATE INDEX IF NOT EXISTS idx_open_ports_remote_ip ON server_discovery.open_ports(remote_ip);

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
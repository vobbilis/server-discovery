-- Create ports table
CREATE TABLE IF NOT EXISTS public.server_ports (
    id SERIAL PRIMARY KEY,
    server_id INTEGER NOT NULL,
    local_port INTEGER NOT NULL,
    local_ip VARCHAR(50),
    remote_port INTEGER,
    remote_ip VARCHAR(50),
    state VARCHAR(50) NOT NULL,
    description TEXT,
    process_id INTEGER,
    process_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id)
);

-- Create index for faster lookups
CREATE INDEX idx_server_ports_server_id ON public.server_ports(server_id);
CREATE INDEX idx_server_ports_local_port ON public.server_ports(local_port);
CREATE INDEX idx_server_ports_state ON public.server_ports(state);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_server_ports_updated_at
    BEFORE UPDATE ON public.server_ports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add some helpful comments
COMMENT ON TABLE public.server_ports IS 'Stores information about open ports and network connections on servers';
COMMENT ON COLUMN public.server_ports.local_port IS 'The local port number on the server';
COMMENT ON COLUMN public.server_ports.local_ip IS 'The local IP address associated with the port';
COMMENT ON COLUMN public.server_ports.remote_port IS 'The remote port number for established connections';
COMMENT ON COLUMN public.server_ports.remote_ip IS 'The remote IP address for established connections';
COMMENT ON COLUMN public.server_ports.state IS 'The state of the port (e.g., LISTENING, ESTABLISHED)';
COMMENT ON COLUMN public.server_ports.description IS 'Description of the service running on this port';
COMMENT ON COLUMN public.server_ports.process_id IS 'ID of the process using this port';
COMMENT ON COLUMN public.server_ports.process_name IS 'Name of the process using this port'; 
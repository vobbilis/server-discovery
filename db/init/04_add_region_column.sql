-- Add region column to servers table
ALTER TABLE server_discovery.servers ADD COLUMN IF NOT EXISTS region VARCHAR(50);

-- Create index for region column
CREATE INDEX IF NOT EXISTS idx_servers_region ON server_discovery.servers(region); 
-- Create server_tags table
CREATE TABLE IF NOT EXISTS server_discovery.server_tags (
    id SERIAL PRIMARY KEY,
    server_id INTEGER NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    tag_value VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES server_discovery.servers(id)
);

-- Create index for faster lookups
CREATE INDEX idx_server_tags_server_id ON server_discovery.server_tags(server_id);
CREATE INDEX idx_server_tags_tag_name ON server_discovery.server_tags(tag_name);

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_server_tags_updated_at
    BEFORE UPDATE ON server_discovery.server_tags
    FOR EACH ROW
    EXECUTE FUNCTION server_discovery.update_updated_at_column();

-- Add some helpful comments
COMMENT ON TABLE server_discovery.server_tags IS 'Stores tags associated with servers';
COMMENT ON COLUMN server_discovery.server_tags.tag_name IS 'The name/key of the tag';
COMMENT ON COLUMN server_discovery.server_tags.tag_value IS 'The value of the tag'; 
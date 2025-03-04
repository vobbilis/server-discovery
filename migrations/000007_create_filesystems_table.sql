-- Create filesystems table
CREATE TABLE IF NOT EXISTS server_discovery.filesystems (
    id SERIAL PRIMARY KEY,
    discovery_id INTEGER NOT NULL REFERENCES server_discovery.discovery_results(id),
    device VARCHAR(255) NOT NULL,
    mount_point VARCHAR(255) NOT NULL,
    fs_type VARCHAR(50) NOT NULL,
    total_bytes BIGINT NOT NULL,
    used_bytes BIGINT NOT NULL,
    free_bytes BIGINT NOT NULL,
    used_percent NUMERIC(5,2) NOT NULL,
    total_inodes BIGINT,
    used_inodes BIGINT,
    free_inodes BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_discovery_mount_point UNIQUE (discovery_id, mount_point)
);

-- Create indexes
CREATE INDEX idx_filesystems_discovery_id ON server_discovery.filesystems(discovery_id);
CREATE INDEX idx_filesystems_mount_point ON server_discovery.filesystems(mount_point);

-- Add sample data
INSERT INTO server_discovery.filesystems (
    discovery_id, device, mount_point, fs_type, 
    total_bytes, used_bytes, free_bytes, used_percent,
    total_inodes, used_inodes, free_inodes
)
SELECT DISTINCT ON (dr.id)
    dr.id,
    CASE 
        WHEN s.os_type = 'linux' THEN '/dev/sda1'
        WHEN s.os_type = 'windows' THEN 'C:'
        ELSE '/dev/disk1s1'
    END as device,
    CASE 
        WHEN s.os_type = 'linux' THEN '/'
        WHEN s.os_type = 'windows' THEN 'C:\'
        ELSE '/'
    END as mount_point,
    CASE 
        WHEN s.os_type = 'linux' THEN 'ext4'
        WHEN s.os_type = 'windows' THEN 'NTFS'
        ELSE 'APFS'
    END as fs_type,
    sd.disk_total_gb * 1024 * 1024 * 1024 as total_bytes,
    (sd.disk_total_gb - sd.disk_free_gb) * 1024 * 1024 * 1024 as used_bytes,
    sd.disk_free_gb * 1024 * 1024 * 1024 as free_bytes,
    ROUND(((sd.disk_total_gb - sd.disk_free_gb) / sd.disk_total_gb * 100)::numeric, 2) as used_percent,
    1048576 as total_inodes,
    524288 as used_inodes,
    524288 as free_inodes
FROM server_discovery.discovery_results dr
JOIN server_discovery.servers s ON dr.server_id = s.id
JOIN server_discovery.server_details sd ON dr.server_id = sd.server_id
ON CONFLICT (discovery_id, mount_point) DO NOTHING; 
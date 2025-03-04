-- Update server_details with realistic hardware information
UPDATE server_discovery.server_details sd
SET 
    cpu_model = CASE 
        WHEN s.os_type = 'linux' THEN 'Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz'
        WHEN s.os_type = 'windows' THEN 'Intel(R) Xeon(R) CPU E-2276G @ 3.80GHz'
        ELSE 'Apple M1 Pro'
    END,
    cpu_count = CASE 
        WHEN s.os_type = 'linux' THEN 16
        WHEN s.os_type = 'windows' THEN 8
        ELSE 10
    END,
    memory_total_gb = CASE 
        WHEN s.os_type = 'linux' THEN 64.00
        WHEN s.os_type = 'windows' THEN 32.00
        ELSE 16.00
    END,
    disk_total_gb = CASE 
        WHEN s.os_type = 'linux' THEN 1000.00
        WHEN s.os_type = 'windows' THEN 500.00
        ELSE 512.00
    END,
    disk_free_gb = CASE 
        WHEN s.os_type = 'linux' THEN 750.00
        WHEN s.os_type = 'windows' THEN 350.00
        ELSE 400.00
    END,
    last_boot_time = NOW() - INTERVAL '7 days' + (random() * INTERVAL '7 days')
FROM server_discovery.servers s
WHERE sd.server_id = s.id;

-- Insert server_metrics with current usage data
INSERT INTO server_discovery.server_metrics (server_id, cpu_usage, memory_total, memory_used, disk_total, disk_used, load_average, process_count, created_at, updated_at)
SELECT 
    s.id,
    -- CPU usage between 20% and 80%
    (20 + random() * 60)::numeric(5,2),
    -- Memory total in bytes (convert from GB)
    (sd.memory_total_gb * 1024 * 1024 * 1024)::bigint,
    -- Memory used between 40% and 90% of total
    ((sd.memory_total_gb * (0.4 + random() * 0.5) * 1024 * 1024 * 1024))::bigint,
    -- Disk total in bytes (convert from GB)
    (sd.disk_total_gb * 1024 * 1024 * 1024)::bigint,
    -- Disk used between 30% and 85% of total
    ((sd.disk_total_gb * (0.3 + random() * 0.55) * 1024 * 1024 * 1024))::bigint,
    -- Load average between 0.5 and 5.0
    (0.5 + random() * 4.5)::numeric(5,2),
    -- Process count between 50 and 500
    (50 + random() * 450)::integer,
    NOW(),
    NOW()
FROM server_discovery.servers s
JOIN server_discovery.server_details sd ON s.id = sd.server_id
WHERE NOT EXISTS (
    SELECT 1 FROM server_discovery.server_metrics sm 
    WHERE sm.server_id = s.id
); 
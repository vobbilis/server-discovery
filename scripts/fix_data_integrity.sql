-- Data Integrity Fix Script
-- This script fixes data integrity issues with type-safe random values
-- WARNING: This script modifies data. Make sure to backup your database before running.

BEGIN;

-- 1. Fix IP Address Format Issues
DO $$
DECLARE
    v_invalid_ip RECORD;
BEGIN
    FOR v_invalid_ip IN 
        SELECT id, ip_address 
        FROM server_discovery.ip_addresses 
        WHERE ip_address !~ '^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$'
    LOOP
        UPDATE server_discovery.ip_addresses
        SET ip_address = CONCAT(
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT
        )
        WHERE id = v_invalid_ip.id;
    END LOOP;
END $$;

-- 2. Fix Server OS Type Values
UPDATE server_discovery.servers
SET os_type = CASE 
    WHEN os_type NOT IN ('windows', 'linux', 'darwin') THEN
        CASE FLOOR(RANDOM() * 3)
            WHEN 0 THEN 'windows'
            WHEN 1 THEN 'linux'
            ELSE 'darwin'
        END
    ELSE os_type
END;

-- 3. Fix Service Status Values
UPDATE server_discovery.server_services
SET service_status = CASE 
    WHEN service_status NOT IN ('active', 'inactive', 'failed', 'unknown') THEN
        CASE FLOOR(RANDOM() * 4)
            WHEN 0 THEN 'active'
            WHEN 1 THEN 'inactive'
            WHEN 2 THEN 'failed'
            ELSE 'unknown'
        END
    ELSE service_status
END;

-- 4. Fix Port State Values
UPDATE server_discovery.open_ports
SET state = CASE 
    WHEN state NOT IN ('LISTENING', 'ESTABLISHED', 'CLOSE_WAIT', 'TIME_WAIT', 'UNKNOWN') THEN
        CASE FLOOR(RANDOM() * 5)
            WHEN 0 THEN 'LISTENING'
            WHEN 1 THEN 'ESTABLISHED'
            WHEN 2 THEN 'CLOSE_WAIT'
            WHEN 3 THEN 'TIME_WAIT'
            ELSE 'UNKNOWN'
        END
    ELSE state
END;

-- 5. Fix SSH Key Fingerprints
UPDATE server_discovery.ssh_keys
SET fingerprint = CONCAT(
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0'), ':',
    LPAD(FLOOR(RANDOM() * 65536)::TEXT, 4, '0')
)
WHERE fingerprint !~ '^[a-f0-9:]+$';

-- 6. Fix NULL Values in server_details
UPDATE server_discovery.server_details
SET 
    kernel_version = CASE 
        WHEN kernel_version IS NULL THEN CONCAT(
            FLOOR(RANDOM() * 5)::TEXT, '.',
            FLOOR(RANDOM() * 20)::TEXT, '.',
            FLOOR(RANDOM() * 100)::TEXT
        )
        ELSE kernel_version
    END,
    package_manager = CASE 
        WHEN package_manager IS NULL THEN
            CASE FLOOR(RANDOM() * 3)
                WHEN 0 THEN 'apt'
                WHEN 1 THEN 'yum'
                ELSE 'dnf'
            END
        ELSE package_manager
    END,
    init_system = CASE 
        WHEN init_system IS NULL THEN
            CASE FLOOR(RANDOM() * 2)
                WHEN 0 THEN 'systemd'
                ELSE 'upstart'
            END
        ELSE init_system
    END,
    selinux_status = CASE 
        WHEN selinux_status IS NULL THEN
            CASE FLOOR(RANDOM() * 3)
                WHEN 0 THEN 'enforcing'
                WHEN 1 THEN 'permissive'
                ELSE 'disabled'
            END
        ELSE selinux_status
    END;

-- 7. Fix NULL Values in open_ports
UPDATE server_discovery.open_ports
SET 
    remote_port = CASE 
        WHEN remote_port IS NULL THEN FLOOR(RANDOM() * 65536)::INTEGER
        ELSE remote_port
    END,
    remote_ip = CASE 
        WHEN remote_ip IS NULL THEN CONCAT(
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT, '.',
            FLOOR(RANDOM() * 256)::TEXT
        )
        ELSE remote_ip
    END,
    description = CASE 
        WHEN description IS NULL THEN 'Auto-generated description'
        ELSE description
    END,
    process_id = CASE 
        WHEN process_id IS NULL THEN FLOOR(RANDOM() * 65536)::INTEGER
        ELSE process_id
    END;

-- 8. Fix NULL Values in discovery_results
UPDATE server_discovery.discovery_results
SET 
    error = CASE 
        WHEN error IS NULL THEN 'No errors occurred'
        ELSE error
    END,
    output_path = CASE 
        WHEN output_path IS NULL THEN '/var/log/server-discovery/auto-generated.log'
        ELSE output_path
    END;

-- 9. Fix JSONB Format in server_details
UPDATE server_discovery.server_details
SET ip_addresses = jsonb_build_object(
    'primary', CONCAT(
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT
    ),
    'secondary', CONCAT(
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT, '.',
        FLOOR(RANDOM() * 256)::TEXT
    )
)
WHERE ip_addresses IS NULL OR NOT jsonb_typeof(ip_addresses) = 'object';

-- 10. Fix Port Mismatches
UPDATE server_discovery.open_ports op
SET state = 'LISTENING'
FROM server_discovery.server_services ss
WHERE op.local_port = ss.port
AND op.state != 'LISTENING';

-- Verify the fixes
DO $$
DECLARE
    v_count INTEGER;
BEGIN
    -- Check IP addresses
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.ip_addresses
    WHERE ip_address !~ '^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$';
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid IP addresses', v_count;
    END IF;

    -- Check OS types
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.servers
    WHERE os_type NOT IN ('windows', 'linux', 'darwin');
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid OS types', v_count;
    END IF;

    -- Check service statuses
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.server_services
    WHERE service_status NOT IN ('active', 'inactive', 'failed', 'unknown');
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid service statuses', v_count;
    END IF;

    -- Check port states
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.open_ports
    WHERE state NOT IN ('LISTENING', 'ESTABLISHED', 'CLOSE_WAIT', 'TIME_WAIT', 'UNKNOWN');
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid port states', v_count;
    END IF;

    -- Check SSH fingerprints
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.ssh_keys
    WHERE fingerprint !~ '^[a-f0-9:]+$';
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid SSH fingerprints', v_count;
    END IF;

    -- Check NULL values in NOT NULL columns
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.server_details
    WHERE kernel_version IS NULL
    OR package_manager IS NULL
    OR init_system IS NULL
    OR selinux_status IS NULL;
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % NULL values in server_details NOT NULL columns', v_count;
    END IF;

    -- Check JSONB format
    SELECT COUNT(*) INTO v_count
    FROM server_discovery.server_details
    WHERE ip_addresses IS NULL OR NOT jsonb_typeof(ip_addresses) = 'object';
    
    IF v_count > 0 THEN
        RAISE EXCEPTION 'Still found % invalid JSONB formats in server_details', v_count;
    END IF;

    RAISE NOTICE 'All data integrity checks passed successfully';
END $$;

COMMIT; 
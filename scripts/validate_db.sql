-- Database Validation Script for server_discovery schema
-- This script validates the database structure and data integrity

DO $$
DECLARE
    v_schema_name text := 'server_discovery';
    v_table_name text;
    v_column_name text;
    v_constraint_name text;
    v_index_name text;
    v_trigger_name text;
    v_count bigint;
    v_error_count integer := 0;
    v_warning_count integer := 0;
    v_info_count integer := 0;
    v_fk_column text;
    v_referenced_table text;
    v_referenced_column text;
BEGIN
    RAISE NOTICE 'Starting database validation for schema: %', v_schema_name;
    RAISE NOTICE '==============================================';

    -- 1. Schema Validation
    RAISE NOTICE '1. Schema Validation';
    RAISE NOTICE '-------------------';

    -- Check if schema exists
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = v_schema_name) THEN
        RAISE EXCEPTION 'Schema % does not exist', v_schema_name;
    END IF;

    -- Check if all required tables exist
    FOR v_table_name IN 
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = v_schema_name
    LOOP
        RAISE NOTICE 'Validating table: %', v_table_name;
        
        -- Check table structure
        FOR v_column_name IN 
            SELECT column_name 
            FROM information_schema.columns 
            WHERE table_schema = v_schema_name 
            AND table_name = v_table_name
        LOOP
            -- Check for NULL values in NOT NULL columns
            EXECUTE format('
                SELECT COUNT(*) 
                FROM %I.%I 
                WHERE %I IS NULL', 
                v_schema_name, v_table_name, v_column_name) INTO v_count;
            
            IF v_count > 0 THEN
                RAISE WARNING 'Found % NULL values in NOT NULL column % of table %', 
                    v_count, v_column_name, v_table_name;
                v_warning_count := v_warning_count + 1;
            END IF;
        END LOOP;

        -- Check foreign key constraints
        FOR v_constraint_name, v_fk_column, v_referenced_table, v_referenced_column IN 
            SELECT 
                tc.constraint_name,
                kcu.column_name,
                ccu.table_name,
                ccu.column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu 
                ON tc.constraint_name = kcu.constraint_name
            JOIN information_schema.constraint_column_usage ccu 
                ON tc.constraint_name = ccu.constraint_name
            WHERE tc.table_schema = v_schema_name
            AND tc.table_name = v_table_name
            AND tc.constraint_type = 'FOREIGN KEY'
        LOOP
            -- Check for orphaned records
            EXECUTE format('
                SELECT COUNT(*) 
                FROM %I.%I t
                LEFT JOIN %I.%I r ON t.%I = r.%I
                WHERE r.%I IS NULL', 
                v_schema_name, v_table_name, v_schema_name, 
                v_referenced_table, v_fk_column, v_referenced_column, v_referenced_column) INTO v_count;
            
            IF v_count > 0 THEN
                RAISE WARNING 'Found % orphaned records in table % referencing %', 
                    v_count, v_table_name, v_referenced_table;
                v_warning_count := v_warning_count + 1;
            END IF;
        END LOOP;

        -- Check indexes
        FOR v_index_name IN 
            SELECT indexname 
            FROM pg_indexes 
            WHERE schemaname = v_schema_name 
            AND tablename = v_table_name
        LOOP
            -- Verify index is being used
            EXECUTE format('
                SELECT COUNT(*) 
                FROM pg_stat_user_indexes 
                WHERE schemaname = %L 
                AND relname = %L 
                AND indexrelname = %L 
                AND idx_scan > 0', 
                v_schema_name, v_table_name, v_index_name) INTO v_count;
            
            IF v_count = 0 THEN
                RAISE WARNING 'Index % on table % has never been used', 
                    v_index_name, v_table_name;
                v_warning_count := v_warning_count + 1;
            END IF;
        END LOOP;

        -- Check triggers
        FOR v_trigger_name IN 
            SELECT trigger_name 
            FROM information_schema.triggers 
            WHERE event_object_schema = v_schema_name 
            AND event_object_table = v_table_name
        LOOP
            -- Verify trigger is enabled
            EXECUTE format('
                SELECT COUNT(*) 
                FROM pg_trigger 
                WHERE tgname = %L 
                AND NOT tgenabled = ''D''', 
                v_trigger_name) INTO v_count;
            
            IF v_count = 0 THEN
                RAISE WARNING 'Trigger % on table % is disabled', 
                    v_trigger_name, v_table_name;
                v_warning_count := v_warning_count + 1;
            END IF;
        END LOOP;
    END LOOP;

    -- 2. Data Integrity Checks
    RAISE NOTICE '2. Data Integrity Checks';
    RAISE NOTICE '----------------------';

    -- Check servers table
    RAISE NOTICE 'Validating servers table...';
    
    -- Check hostname uniqueness
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.servers s1 
        JOIN %I.servers s2 ON s1.hostname = s2.hostname 
        WHERE s1.id != s2.id', 
        v_schema_name, v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % duplicate hostnames in servers table', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check status values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.servers 
        WHERE status NOT IN (''unknown'', ''active'', ''inactive'', ''error'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % servers with invalid status values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check os_type values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.servers 
        WHERE os_type NOT IN (''windows'', ''linux'', ''darwin'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % servers with invalid os_type values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check server_metrics table
    RAISE NOTICE 'Validating server_metrics table...';
    
    -- Check for invalid numeric values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_metrics 
        WHERE cpu_usage < 0 OR cpu_usage > 100 
        OR memory_total < 0 
        OR memory_used < 0 
        OR disk_total < 0 
        OR disk_used < 0
        OR (memory_used > memory_total)
        OR (disk_used > disk_total)', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % invalid numeric values in server_metrics table', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check discovery_results table
    RAISE NOTICE 'Validating discovery_results table...';
    
    -- Check timestamp logic
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.discovery_results 
        WHERE end_time < start_time', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % discovery results with end_time before start_time', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check numeric ranges
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.discovery_results 
        WHERE cpu_count < 0 
        OR memory_total_gb < 0 
        OR disk_total_gb < 0 
        OR disk_free_gb < 0 
        OR disk_free_gb > disk_total_gb', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % discovery results with invalid numeric values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check server_tags table
    RAISE NOTICE 'Validating server_tags table...';
    
    -- Check tag_name format
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_tags 
        WHERE tag_name !~ ''^[a-zA-Z0-9_-]+$''', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % tags with invalid tag_name format', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check ip_addresses table
    RAISE NOTICE 'Validating ip_addresses table...';
    
    -- Check IP address format
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.ip_addresses 
        WHERE ip_address !~ ''^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$''', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % invalid IP address formats', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check installed_software table
    RAISE NOTICE 'Validating installed_software table...';
    
    -- Check version format
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.installed_software 
        WHERE version IS NOT NULL 
        AND version !~ ''^[0-9]+(?:\.[0-9]+)*$''', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % software entries with invalid version format', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check running_services table
    RAISE NOTICE 'Validating running_services table...';
    
    -- Check service status values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.running_services 
        WHERE status NOT IN (''running'', ''stopped'', ''disabled'', ''unknown'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % services with invalid status values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check server_services table
    RAISE NOTICE 'Validating server_services table...';
    
    -- Check service status values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_services 
        WHERE service_status NOT IN (''active'', ''inactive'', ''failed'', ''unknown'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % services with invalid status values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check port ranges
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_services 
        WHERE port IS NOT NULL 
        AND (port < 0 OR port > 65535)', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % services with invalid port numbers', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check open_ports table
    RAISE NOTICE 'Validating open_ports table...';
    
    -- Check port number ranges
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.open_ports 
        WHERE local_port < 0 OR local_port > 65535 
        OR (remote_port IS NOT NULL AND (remote_port < 0 OR remote_port > 65535))', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % invalid port numbers in open_ports table', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check state values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.open_ports 
        WHERE state NOT IN (''LISTENING'', ''ESTABLISHED'', ''CLOSE_WAIT'', ''TIME_WAIT'', ''UNKNOWN'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % ports with invalid state values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check server_details table
    RAISE NOTICE 'Validating server_details table...';
    
    -- Check JSONB field formats
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_details 
        WHERE ip_addresses IS NOT NULL 
        AND NOT jsonb_typeof(ip_addresses) = ''object''', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % invalid ip_addresses JSONB format in server_details table', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check numeric ranges
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_details 
        WHERE cpu_count < 0 
        OR memory_total_gb < 0 
        OR disk_total_gb < 0 
        OR disk_free_gb < 0 
        OR disk_free_gb > disk_total_gb', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % server_details with invalid numeric values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check ssh_keys table
    RAISE NOTICE 'Validating ssh_keys table...';
    
    -- Check key_type values
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.ssh_keys 
        WHERE key_type NOT IN (''ssh-rsa'', ''ssh-ed25519'', ''ecdsa-sha2-nistp256'', ''ecdsa-sha2-nistp384'', ''ecdsa-sha2-nistp521'')', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % SSH keys with invalid key_type values', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check fingerprint format
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.ssh_keys 
        WHERE fingerprint !~ ''^[a-f0-9:]+$''', 
        v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % SSH keys with invalid fingerprint format', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- 3. Cross-Table Validations
    RAISE NOTICE '3. Cross-Table Validations';
    RAISE NOTICE '------------------------';

    -- Check consistency between server_details and discovery_results
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_details sd 
        LEFT JOIN %I.discovery_results dr ON sd.discovery_id = dr.id 
        WHERE dr.id IS NULL', 
        v_schema_name, v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % server_details records without corresponding discovery_results', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check port information consistency
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.open_ports op 
        JOIN %I.server_services ss ON op.local_port = ss.port 
        WHERE op.state != ''LISTENING''', 
        v_schema_name, v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % port mismatches between open_ports and server_services', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- Check metrics consistency
    EXECUTE format('
        SELECT COUNT(*) 
        FROM %I.server_metrics sm 
        JOIN %I.server_details sd ON sm.server_id = sd.server_id 
        WHERE ABS(sm.memory_total - (sd.memory_total_gb * 1024 * 1024 * 1024)) > 1024 * 1024 * 1024', 
        v_schema_name, v_schema_name) INTO v_count;
    
    IF v_count > 0 THEN
        RAISE WARNING 'Found % memory value inconsistencies between server_metrics and server_details', v_count;
        v_warning_count := v_warning_count + 1;
    END IF;

    -- 4. Performance Checks
    RAISE NOTICE '4. Performance Checks';
    RAISE NOTICE '-------------------';

    -- Check for large tables
    FOR v_table_name IN 
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = v_schema_name
    LOOP
        EXECUTE format('
            SELECT COUNT(*) 
            FROM %I.%I', 
            v_schema_name, v_table_name) INTO v_count;
        
        IF v_count > 1000000 THEN
            RAISE WARNING 'Table % has % rows - consider partitioning', v_table_name, v_count;
            v_warning_count := v_warning_count + 1;
        END IF;
    END LOOP;

    -- Summary
    RAISE NOTICE 'Validation Summary';
    RAISE NOTICE '-----------------';
    RAISE NOTICE 'Errors: %', v_error_count;
    RAISE NOTICE 'Warnings: %', v_warning_count;
    RAISE NOTICE 'Info: %', v_info_count;
    
    IF v_error_count > 0 THEN
        RAISE EXCEPTION 'Database validation failed with % errors', v_error_count;
    ELSIF v_warning_count > 0 THEN
        RAISE WARNING 'Database validation completed with % warnings', v_warning_count;
    ELSE
        RAISE NOTICE 'Database validation completed successfully';
    END IF;
END $$; 
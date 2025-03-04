#!/bin/bash

# Start PostgreSQL container using docker-compose
docker-compose -f docker-compose.test.yml up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker exec server_discovery_test_db pg_isready -U postgres > /dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 1
done

# Create schema if needed (schema will be created by the tests, but we can verify connection here)
docker exec server_discovery_test_db psql -U postgres -d server_discovery_test -c "\dt"

# Make the script executable
chmod +x scripts/setup_test_db.sh 
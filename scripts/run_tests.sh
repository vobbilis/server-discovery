#!/bin/bash

# Ensure Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
./scripts/setup_test_db.sh

# Run the tests
echo "Running tests..."
go test -v ./...

# Get the test exit code
TEST_EXIT_CODE=$?

# Clean up: Stop and remove the PostgreSQL container
echo "Cleaning up..."
docker-compose -f docker-compose.test.yml down

# Exit with the test exit code
exit $TEST_EXIT_CODE 
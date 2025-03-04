# Data Generation Tools

This directory contains tools used to generate and populate test data for the server discovery project.

## Files

- `update_ports.go` - Tool to generate and insert port information for all servers in the database
- `fixed_createSampleFiles.go` - Creates sample files for testing
- `mock_data.go` - Contains mock data generation functions
- `mock_api.go` - Mock API implementations for testing
- `mock_winrm.go` - Mock WinRM implementations for testing

## Database Initialization

The generated data has been saved in a PostgreSQL dump file at `../../db/init/01_server_discovery_dump.sql`. This dump is automatically loaded when the database container starts via Docker Compose.

## Usage

These tools are primarily used for:
1. Initial database population
2. Generating test data
3. Creating mock implementations for testing

The data has already been generated and stored in the database dump. You typically don't need to run these tools unless you want to regenerate the data or modify the test dataset. 
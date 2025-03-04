# Server Discovery

A comprehensive tool for discovering and monitoring servers in your network infrastructure. This application consists of a backend API server built with Go and a frontend web interface that provides visualization and management capabilities.

## Features

- Automated server discovery across networks
- Detailed server information collection
- Web-based dashboard for monitoring server status
- Configurable discovery parameters
- Optional database integration for persistent storage
- Kubernetes-ready deployment with Helm charts

## Prerequisites

- Go 1.22 or higher (for backend development)
- Node.js and npm (for frontend development)
- Docker and Docker Compose (for containerized development)
- Minikube (for local Kubernetes deployment)
- kubectl (for Kubernetes management)
- Helm (for Kubernetes deployment)

## Quick Start

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/server-discovery.git
   cd server-discovery
   ```

2. Start the backend:
   ```bash
   # Navigate to the backend directory
   cd backend
   
   # Install dependencies
   go mod download
   
   # Run the backend server
   # Adding this line
   go run server_discovery_controller.go
   ```

3. Start the frontend (in a new terminal):
   ```bash
   # Navigate to the frontend directory
   cd frontend
   
   # Install dependencies
   npm install
   
   # Run the development server
   npm start
   ```

4. Access the application at http://localhost:3000

### Using Docker Compose

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/server-discovery.git
   cd server-discovery
   ```

2. Build and start the containers:
   ```bash
   docker-compose up -d
   ```

3. Access the application at http://localhost:80

### Running Tests

The project uses Docker PostgreSQL for testing to ensure a clean, isolated test environment. Follow these steps to run the tests:

1. Ensure Docker and Docker Compose are installed and running on your system.

2. Make the test script executable:
   ```bash
   chmod +x scripts/run_tests.sh
   ```

3. Run the tests using the provided script:
   ```bash
   ./scripts/run_tests.sh
   ```

This script will:
- Start a PostgreSQL container specifically for testing (on port 5433 to avoid conflicts)
- Create a clean test database
- Run all tests including:
  - Database operations tests
  - Server discovery tests
  - Integration tests
- Clean up by stopping and removing the test container

You can also run specific test files if needed:
```bash
# Start the test database
docker-compose -f docker-compose.test.yml up -d

# Run specific tests
go test -v ./... -run TestDatabaseOperations
# or
go test -v ./... -run TestServerDiscoveryMinimal

# Clean up after testing
docker-compose -f docker-compose.test.yml down
```

Note: The test PostgreSQL instance runs on port 5433 to avoid conflicts with any local PostgreSQL installation.

### Stress Testing

The project includes stress tests to evaluate scalability with large numbers of servers. To run stress tests:

```bash
go test -v -run TestStressDiscovery
```

The stress tests simulate various scenarios:
- Small scale (100 servers)
- Medium scale (1000 servers)
- Large scale (5000 servers)
- Worst case (1000 servers with high failure rate)

Each test evaluates:
- Throughput (servers/second)
- Success/failure rates
- Average discovery time per server
- Performance under different concurrency levels
- Behavior with varying network delays
- System resilience with different failure rates

The stress tests help ensure the system can handle:
- Large numbers of servers (1000s)
- Mixed Windows and Linux environments
- Network delays and timeouts
- High failure rates
- Concurrent operations

Test parameters can be adjusted in `stress_test.go` to simulate specific scenarios or requirements.

## Kubernetes Deployment

### Setting up Minikube

1. Start Minikube:
   ```bash
   minikube start
   ```

2. Enable the Minikube registry:
   ```bash
   minikube addons enable registry
   ```

### Building and Pushing Docker Images

1. Build the backend image:
   ```bash
   docker build -t $(minikube ip):5000/server-discovery-backend:latest -f k8s/Dockerfile.backend .
   ```

2. Build the frontend image:
   ```bash
   docker build -t $(minikube ip):5000/server-discovery-frontend:latest -f k8s/Dockerfile.frontend .
   ```

3. Push images to the Minikube registry:
   ```bash
   docker push $(minikube ip):5000/server-discovery-backend:latest
   docker push $(minikube ip):5000/server-discovery-frontend:latest
   ```

### Deploying with Helm

1. Install the Helm chart:
   ```bash
   helm install server-discovery ./k8s/helm
   ```

2. Set up local DNS (add to /etc/hosts):
   ```
   $(minikube ip) server-discovery.local
   ```

3. Access the application at http://server-discovery.local

## Configuration

The application can be configured through the `config.json` file or environment variables. When deployed to Kubernetes, configuration is managed through the Helm values.

### Key Configuration Parameters

#### API Server
- `port`: The port on which the API server listens (default: 8080)
- `allowedOrigins`: CORS allowed origins (default: "*")
- `readTimeout`: HTTP read timeout in seconds (default: 15)
- `writeTimeout`: HTTP write timeout in seconds (default: 15)

#### Discovery
- `concurrency`: Number of concurrent discovery operations (default: 10)
- `timeout`: Discovery timeout in seconds (default: 300)
- `retryCount`: Number of retry attempts (default: 3)
- `retryDelay`: Delay between retries in seconds (default: 5)

#### Database
- `enabled`: Enable database integration (default: false)
- `host`: Database host (default: "postgres")
- `port`: Database port (default: 5432)
- `database`: Database name (default: "server_discovery")
- `user`: Database user (default: "postgres")
- `password`: Database password (default: "postgres")

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 

# Add all files to staging
git add .

# Verify what will be committed
git status 

## Project Structure

- `api_server.go` - Main API server implementation
- `config.go` - Configuration handling
- `database.go` - Database operations
- `database_queries.go` - SQL queries
- `discovery_types.go` - Type definitions for discovery
- `metrics.go` - Metrics collection
- `server_discovery_controller.go` - Main controller logic
- `ssh_handler.go` - SSH connection handling
- `types.go` - Common type definitions

### Test Data and Development Setup

- `db/init/` - Contains database initialization files
  - `01_server_discovery_dump.sql` - PostgreSQL dump with pre-populated test data

- `tools/data_generation/` - Contains tools used to generate test data
  - See [tools/data_generation/README.md](tools/data_generation/README.md) for details

## Getting Started

1. Start the database:
   ```bash
   docker-compose up -d
   ```
   This will automatically create the database and load the test data.

2. Run the server:
   ```bash
   go run server_discovery_controller.go
   ```

## Development

The project uses PostgreSQL for data storage. The database is automatically populated with test data when started using Docker Compose.

### Database Details
- Host: localhost
- Port: 5433
- User: postgres
- Password: postgres
- Database: server_discovery

### Test Data
The database comes pre-populated with:
- 500 servers (213 Windows, 287 Linux)
- Port information for all servers
- Sample metrics and discovery results 

# Server Discovery Database

This project manages server discovery data in PostgreSQL, using a dedicated schema for isolation and organization.

## Database Structure

- **Database Name**: `server_discovery`
- **Schema Name**: `server_discovery`
- **Port**: 5433 (test database)

### Tables

All tables are created in the `server_discovery` schema:

1. `servers` - Core server information
   - Basic server details (hostname, IP, region)
   - OS information and status
   - Timestamps for creation and updates

2. `server_services` - Configured services on servers
   - Links to server via `server_id`
   - Service details (name, status, port)
   - Timestamps for tracking changes

3. `discovery_results` - Results from server discovery scans
   - System information (OS, CPU, memory, disk)
   - Scan timing and status
   - Links to server via `server_id`

4. `open_ports` - Discovered open ports during scans
   - Port details (local/remote ports and IPs)
   - Process information
   - Links to discovery via `discovery_id`

## Setup Instructions

1. Create the database:
```sql
CREATE DATABASE server_discovery;
```

2. Run the test data generation:
```bash
# From the project root
go test -v ./... -run TestLoadDatabaseWithServers
```

This will:
- Create the schema if it doesn't exist
- Create all required tables
- Generate sample data including:
  - 500 servers across different regions
  - Mix of Windows and Linux systems
  - Various services and open ports
  - Discovery results with system information

## Connection Details

- Host: localhost
- Port: 5433
- Database: server_discovery
- User: postgres
- Password: postgres
- Schema: server_discovery

## Notes

- The schema uses `WITH TIME ZONE` for all timestamps
- Foreign key constraints ensure referential integrity
- Tables include `created_at` and `updated_at` timestamps
- The schema is isolated from the public schema for better organization

# Server Discovery Service

This service provides functionality for discovering and monitoring server configurations and services.

## Port Usage

The service uses the following ports:

- **8090**: Main API server port
  - Handles all HTTP API endpoints including:
    - `/api/stats`: Server statistics
    - `/api/servers`: Server listing
    - `/api/servers/{id}/discoveries`: Server discovery history
    - `/api/server-tags`: Server tags

- **9090**: Metrics port
  - Exposes monitoring metrics

## Configuration

The service configuration is managed through `config.json`. Key configuration options include:

- API server settings (port, timeouts, CORS)
- Database connection details
- Discovery settings (concurrency, timeouts)
- Metrics configuration

## API Endpoints

### GET /api/stats
Returns statistics about discovered servers and their regions.

### GET /api/servers
Lists all servers with their current status and metrics.

### GET /api/servers/{id}/discoveries
Returns the discovery history for a specific server.

### GET /api/server-tags
Returns all unique tags across all servers.

## Database

The service uses PostgreSQL with the following connection details:
- Host: localhost
- Port: 5433
- Database: server_discovery
- Schema: server_discovery

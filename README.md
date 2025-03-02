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
   go run main.go
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
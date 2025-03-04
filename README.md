# Server Discovery Application

A comprehensive server discovery and monitoring application that provides detailed insights into server configurations, open ports, installed software, and more.

## Features

- Automated server discovery across networks
- Detailed server information collection
- Web-based dashboard for monitoring server status
- Configurable discovery parameters
- Database integration for persistent storage
- Kubernetes-ready deployment with Helm charts
- Comprehensive API documentation
- Stress testing capabilities

## Prerequisites

### Docker Setup (Recommended)
- Docker 20.10 or later
- Docker Compose 2.0 or later
- Git

### Manual Setup
- Go 1.22 or later
- Node.js 18.x or later
- PostgreSQL 14.x or later
- Git

## Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/vobbilis/server-discovery.git
cd server-discovery
```

2. Create environment file:
```bash
cp .env.example .env
```

3. Start the application:
```bash
docker compose up -d
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8090
- API Documentation: http://localhost:8090/api/docs

## Manual Setup (Without Docker)

### Database Setup

1. Create a PostgreSQL database:
```bash
createdb server_discovery
```

2. Run the migrations:
```bash
cd migrations
psql -U your_postgres_user -d server_discovery -f add_ports_table.sql
```

### Backend Setup

1. Install Go dependencies:
```bash
go mod download
```

2. Build and run the server:
```bash
cd cmd/server
go build -o server
./server
```

### Frontend Setup

1. Install Node.js dependencies:
```bash
cd frontend
npm install
```

2. Start the development server:
```bash
npm start
```

## Development

### Docker Development

1. Start the development environment:
```bash
docker compose up -d
```

2. View logs:
```bash
docker compose logs -f
```

3. Stop the environment:
```bash
docker compose down
```

### API Documentation

The API documentation is available in OpenAPI/Swagger format at:
- http://localhost:8090/api/docs (when running locally)
- http://localhost:8090/api/swagger.yaml (raw Swagger file)

### Environment Variables

Key environment variables (can be set in `.env` file):

```env
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=server_discovery
DB_PASSWORD=server_discovery
DB_NAME=server_discovery

# API Configuration
API_PORT=8090

# Frontend Configuration
REACT_APP_API_URL=http://localhost:8090
```

## Testing

### Running Tests

1. Backend tests:
```bash
./scripts/run_tests.sh
```

2. Frontend tests:
```bash
cd frontend
npm test
```

### Test Database Setup

```bash
./scripts/setup_test_db.sh
```

### Stress Testing

The project includes stress tests to evaluate scalability:
```bash
go test -v -run TestStressDiscovery
```

Stress tests simulate:
- Small scale (100 servers)
- Medium scale (1000 servers)
- Large scale (5000 servers)
- Worst case (1000 servers with high failure rate)

## Kubernetes Deployment

### Prerequisites
- Minikube
- kubectl
- Helm

### Setting up Minikube

1. Start Minikube:
```bash
minikube start
```

2. Enable the Minikube registry:
```bash
minikube addons enable registry
```

### Deploying with Helm

1. Build and push images:
```bash
# Build images
docker build -t $(minikube ip):5000/server-discovery-backend:latest -f k8s/Dockerfile.backend .
docker build -t $(minikube ip):5000/server-discovery-frontend:latest -f k8s/Dockerfile.frontend .

# Push to Minikube registry
docker push $(minikube ip):5000/server-discovery-backend:latest
docker push $(minikube ip):5000/server-discovery-frontend:latest
```

2. Deploy with Helm:
```bash
helm install server-discovery ./k8s/helm
```

3. Access the application at http://server-discovery.local

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   - If you see "address already in use" errors, check for existing processes:
     ```bash
     lsof -i :8090  # For backend
     lsof -i :3000  # For frontend
     ```

2. **Database Connection Issues**
   - Verify PostgreSQL is running:
     ```bash
     docker compose ps postgres
     ```
   - Check database logs:
     ```bash
     docker compose logs postgres
     ```

3. **Container Issues**
   - Restart containers:
     ```bash
     docker compose restart
     ```
   - Rebuild containers:
     ```bash
     docker compose up -d --build
     ```

### Logs

View logs for specific services:
```bash
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f postgres
```

## Project Structure

```
server-discovery/
├── api/                 # API documentation
├── cmd/                 # Application entrypoints
├── frontend/           # React frontend application
├── k8s/                # Kubernetes deployment files
├── migrations/         # Database migrations
├── pkg/                # Go packages
│   ├── controller/     # Controllers
│   ├── database/      # Database operations
│   ├── discovery/     # Server discovery logic
│   └── models/        # Data models
├── scripts/           # Utility scripts
└── tools/             # Development tools
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

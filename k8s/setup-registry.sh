#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Enabling Minikube registry addon...${NC}"
minikube addons enable registry

echo -e "${YELLOW}Setting up Docker environment to use Minikube's Docker daemon...${NC}"
eval $(minikube docker-env)

# Get Minikube IP
MINIKUBE_IP=$(minikube ip)
REGISTRY_URL="${MINIKUBE_IP}:5000"

echo -e "${YELLOW}Building backend Docker image...${NC}"
docker build -f k8s/Dockerfile.backend -t server-discovery-backend:latest .

echo -e "${YELLOW}Building frontend Docker image...${NC}"
docker build -f k8s/Dockerfile.frontend -t server-discovery-frontend:latest .

echo -e "${YELLOW}Tagging images for local registry...${NC}"
docker tag server-discovery-backend:latest ${REGISTRY_URL}/server-discovery-backend:latest
docker tag server-discovery-frontend:latest ${REGISTRY_URL}/server-discovery-frontend:latest

echo -e "${YELLOW}Pushing images to local registry...${NC}"
docker push ${REGISTRY_URL}/server-discovery-backend:latest
docker push ${REGISTRY_URL}/server-discovery-frontend:latest

echo -e "${YELLOW}Creating values file with registry URL...${NC}"
cat > k8s/helm/values-local.yaml << EOF
image:
  backend:
    repository: ${REGISTRY_URL}/server-discovery-backend
    tag: latest
    pullPolicy: IfNotPresent
  frontend:
    repository: ${REGISTRY_URL}/server-discovery-frontend
    tag: latest
    pullPolicy: IfNotPresent
EOF

echo -e "${YELLOW}Deploying application with Helm...${NC}"
helm upgrade --install server-discovery ./k8s/helm -f k8s/helm/values-local.yaml

echo -e "${YELLOW}Waiting for deployment to be ready...${NC}"
kubectl rollout status deployment/server-discovery --timeout=120s

echo -e "${GREEN}Registry setup complete!${NC}"
echo -e "${GREEN}Images are available at ${REGISTRY_URL}/server-discovery-backend:latest and ${REGISTRY_URL}/server-discovery-frontend:latest${NC}" 
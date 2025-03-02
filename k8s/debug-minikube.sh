#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Checking Minikube status...${NC}"
minikube status

echo -e "${YELLOW}Checking Docker environment...${NC}"
docker info | grep -E 'Name:|Server:'

echo -e "${YELLOW}Checking for server-discovery images...${NC}"
docker images | grep server-discovery

echo -e "${YELLOW}Checking Kubernetes resources...${NC}"
echo -e "${YELLOW}Pods:${NC}"
kubectl get pods

echo -e "${YELLOW}Pod details:${NC}"
POD_NAME=$(kubectl get pods -l app.kubernetes.io/name=server-discovery -o jsonpath="{.items[0].metadata.name}" 2>/dev/null || echo "")
if [ -n "$POD_NAME" ]; then
  kubectl describe pod $POD_NAME
else
  echo -e "${RED}No pods found with label app.kubernetes.io/name=server-discovery${NC}"
fi

echo -e "${YELLOW}Services:${NC}"
kubectl get services

echo -e "${YELLOW}Ingress:${NC}"
kubectl get ingress

echo -e "${GREEN}Debug complete!${NC}" 
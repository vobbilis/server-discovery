#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Minikube...${NC}"
minikube start --driver=docker --memory=4096 --cpus=2

echo -e "${YELLOW}Enabling Ingress addon...${NC}"
minikube addons enable ingress

echo -e "${YELLOW}Setting up registry...${NC}"
./k8s/setup-registry.sh

echo -e "${YELLOW}Setting up local hosts file...${NC}"
MINIKUBE_IP=$(minikube ip)
echo -e "${GREEN}Minikube IP: ${MINIKUBE_IP}${NC}"
echo -e "${YELLOW}Please add the following line to your /etc/hosts file:${NC}"
echo -e "${GREEN}${MINIKUBE_IP} server-discovery.local${NC}"
echo -e "${YELLOW}Command: sudo sh -c \"echo '${MINIKUBE_IP} server-discovery.local' >> /etc/hosts\"${NC}"

echo -e "${GREEN}Setup complete!${NC}"
echo -e "${GREEN}You can access the application at: http://server-discovery.local${NC}"
echo -e "${YELLOW}To open the Kubernetes dashboard, run: minikube dashboard${NC}" 
#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Uninstalling Helm release...${NC}"
helm uninstall server-discovery || true

echo -e "${YELLOW}Stopping Minikube...${NC}"
minikube stop

echo -e "${GREEN}Cleanup complete!${NC}"
echo -e "${YELLOW}To completely delete the Minikube cluster, run: minikube delete${NC}" 
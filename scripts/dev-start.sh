#!/bin/bash

set -e

cd "$(dirname "$0")/.."

echo "starting heroin development environment..."

# ensure docker is running
if ! docker info > /dev/null 2>&1; then
    echo "error: docker is not running"
    exit 1
fi

# build and start services
cd docker
docker-compose down --remove-orphans
docker-compose build heroin-server
docker-compose up -d

echo "waiting for services to start..."
sleep 10

# check service health
echo "checking service health..."
curl -f http://localhost:8082/healthz || echo "warning: heroin server not responding"
curl -f http://localhost:8081/api/v0/version || echo "warning: ipfs not responding"

echo ""
echo "heroin development environment started!"
echo ""
echo "services:"
echo "  heroin server (grpc): localhost:8080"
echo "  heroin server (http): localhost:8082"
echo "  ipfs api:             localhost:5001"
echo "  ipfs gateway:         localhost:8081"
echo "  prometheus:           localhost:9091"
echo "  grafana:              localhost:3000 (admin/admin)"
echo ""
echo "to stop: ./scripts/dev-stop.sh"
echo "to view logs: docker-compose -f docker/docker-compose.yml logs -f"

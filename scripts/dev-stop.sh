#!/bin/bash

cd "$(dirname "$0")/../docker"

echo "stopping heroin development environment..."
docker-compose down

echo "heroin development environment stopped"

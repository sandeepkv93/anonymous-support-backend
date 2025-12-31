#!/bin/bash

set -e

echo "=== Anonymous Support Backend Setup ==="
echo

echo "1. Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting." >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
echo "✓ Prerequisites met"
echo

echo "2. Creating .env file if it doesn't exist..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✓ .env file created"
else
    echo "✓ .env file already exists"
fi
echo

echo "3. Installing Go tools..."
make install-tools
echo

echo "4. Starting infrastructure services..."
docker-compose up -d postgres mongodb redis
echo "✓ Infrastructure services started"
echo

echo "5. Waiting for services to be ready..."
sleep 10
echo "✓ Services ready"
echo

echo "6. Running database migrations..."
make migrate-up
echo

echo "7. Initializing MongoDB..."
make mongo-init
echo

echo "8. Downloading Go dependencies..."
go mod download
echo "✓ Dependencies downloaded"
echo

echo "=== Setup Complete ==="
echo
echo "To start the application:"
echo "  make run"
echo
echo "Or to run everything with Docker:"
echo "  make docker-up"
echo

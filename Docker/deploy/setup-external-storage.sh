#!/bin/bash

# Setup external storage for Stellar Sync
# This script creates the external data directory and rebuilds containers

set -e

echo "====================================="
echo "Setting up external storage for Stellar Sync"
echo "====================================="

# Create data directory (relative to server root)
echo "Creating external data directory..."
mkdir -p ../data/files
chmod 755 ../data/files

echo "Data directory created at: $(pwd)/../data/files"

# Stop existing containers
echo "Stopping existing containers..."
docker-compose down

# Build new images
echo "Building new images..."
docker-compose build

# Start services
echo "Starting services with external storage..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Check service status
echo "Checking service status..."
docker-compose ps

echo "====================================="
echo "Setup complete!"
echo "====================================="
echo "External storage directory: $(pwd)/../data/files"
echo "Files will be automatically cleaned up after 6 hours"
echo ""
echo "To monitor file cleanup, check logs with:"
echo "  docker-compose logs -f file-server | grep CLEANUP"
echo ""
echo "To check storage usage:"
echo "  du -sh ../data/files"
echo "  ls -la ../data/files | wc -l"

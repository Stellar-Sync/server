#!/bin/bash

# 🚀 Stellar Sync - Deploy Anywhere
# This script can be run on any VM with Docker to deploy Stellar Sync

set -e

echo "====================================="
echo "Stellar Sync - Deploy Anywhere"
echo "====================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo "❌ docker-compose not found. Installing..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p ssl logs

# Pull latest images
echo "📥 Pulling latest images..."
docker-compose pull

# Start services
echo "🚀 Starting services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 15

# Check status
echo "📊 Service status:"
docker-compose ps

echo ""
echo "====================================="
echo "✅ Deployment complete!"
echo "====================================="
echo ""
echo "🌐 Services available:"
echo "  - Main server: http://localhost:6000"
echo "  - File server: http://localhost:6200"
echo "  - Nginx: http://localhost:80 -> https://localhost:443"
echo ""
echo "📋 Management commands:"
echo "  docker-compose ps          # Check status"
echo "  docker-compose logs        # View logs"
echo "  docker-compose restart     # Restart services"
echo "  docker-compose down        # Stop services"
echo ""
echo "🔒 To add SSL:"
echo "  1. Point your domain to this server"
echo "  2. Run: ./add-ssl.sh your-domain.com your-email@example.com"
echo ""
echo "💾 Data is stored in Docker volumes:"
echo "  - Files: stellarsync-data"
echo "  - SSL: stellarsync-ssl"
echo "  - Logs: stellarsync-logs"

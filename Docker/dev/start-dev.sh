#!/bin/bash

# StellarSync Development Environment Startup Script

set -e

echo "🚀 Starting StellarSync Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Create logs directory if it doesn't exist
mkdir -p logs

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.dev.yml down

# Build and start the development environment
echo "🔨 Building and starting development containers..."
docker-compose -f docker-compose.dev.yml up --build

echo "✅ Development environment started!"
echo ""
echo "📡 Services available at:"
echo "   Main Server: http://localhost:6000"
echo "   File Server: http://localhost:6200"
echo "   Health Check: http://localhost:6000/health"
echo ""
echo "📝 Logs are available in the ./logs directory"
echo ""
echo "🛠️  Useful commands:"
echo "   View logs: docker-compose -f docker-compose.dev.yml logs -f"
echo "   Stop: docker-compose -f docker-compose.dev.yml down"
echo "   Shell access: docker exec -it stellarsync-main-dev bash"

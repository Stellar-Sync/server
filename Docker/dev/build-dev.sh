#!/bin/bash

# StellarSync Development Environment Build Script

set -e

echo "🔨 Building StellarSync Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Create logs directory if it doesn't exist
mkdir -p logs

# Build the development images
echo "🏗️  Building development images..."
docker-compose -f docker-compose.dev.yml build

echo "✅ Development environment built successfully!"
echo ""
echo "🚀 To start the development environment, run:"
echo "   docker-compose -f docker-compose.dev.yml up"
echo ""
echo "📡 Or use the convenience script:"
echo "   ./start-dev.sh"

#!/bin/bash

# 🔨 Build and Push Stellar Sync Containers
# Run this from your development machine to build and push to Docker Hub

set -e

# Configuration
DOCKER_USERNAME="kasuaberra"
IMAGE_PREFIX="stellarsync"
VERSION="latest"

echo "====================================="
echo "Building and Pushing Stellar Sync Containers"
echo "====================================="

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ This script must be run from the Docker/deploy directory"
    exit 1
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if StellarSync directory exists
if [ ! -d "../../StellarSync" ]; then
    echo "❌ StellarSync directory not found at ../../StellarSync"
    echo "Current directory: $(pwd)"
    echo "Expected structure: server/Docker/deploy/ (current) -> server/StellarSync/"
    exit 1
fi

# Build images
echo "🔨 Building images..."

echo "Building main server..."
docker build -f Dockerfile.main -t ${DOCKER_USERNAME}/${IMAGE_PREFIX}-main:${VERSION} ../../StellarSync

echo "Building file server..."
docker build -f Dockerfile.fileserver -t ${DOCKER_USERNAME}/${IMAGE_PREFIX}-fileserver:${VERSION} ../../StellarSync

# Tag with version
echo "🏷️ Tagging images..."
docker tag ${DOCKER_USERNAME}/${IMAGE_PREFIX}-main:${VERSION} ${DOCKER_USERNAME}/${IMAGE_PREFIX}-main:${VERSION}
docker tag ${DOCKER_USERNAME}/${IMAGE_PREFIX}-fileserver:${VERSION} ${DOCKER_USERNAME}/${IMAGE_PREFIX}-fileserver:${VERSION}

# Push to Docker Hub
echo "📤 Pushing to Docker Hub..."

echo "Pushing main server..."
docker push ${DOCKER_USERNAME}/${IMAGE_PREFIX}-main:${VERSION}

echo "Pushing file server..."
docker push ${DOCKER_USERNAME}/${IMAGE_PREFIX}-fileserver:${VERSION}

echo ""
echo "====================================="
echo "✅ Build and push complete!"
echo "====================================="
echo ""
echo "🐳 Images pushed:"
echo "  - ${DOCKER_USERNAME}/${IMAGE_PREFIX}-main:${VERSION}"
echo "  - ${DOCKER_USERNAME}/${IMAGE_PREFIX}-fileserver:${VERSION}"
echo ""
echo "🚀 To deploy on any VM:"
echo "  1. Copy docker-compose.yml, nginx.conf, proxy_common.conf"
echo "  2. Run: ./deploy-anywhere.sh"
echo ""
echo "📁 Required files for deployment:"
echo "  - docker-compose.yml"
echo "  - nginx.conf"
echo "  - proxy_common.conf"
echo "  - deploy-anywhere.sh"
echo "  - add-ssl.sh (optional, for SSL)"

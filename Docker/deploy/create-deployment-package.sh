#!/bin/bash

# 📦 Create Deployment Package
# Creates a minimal package with only the files needed for VM deployment

set -e

PACKAGE_NAME="stellarsync-deployment"
PACKAGE_DIR="./${PACKAGE_NAME}"

echo "====================================="
echo "Creating Stellar Sync Deployment Package"
echo "====================================="

# Clean up any existing package
if [ -d "$PACKAGE_DIR" ]; then
    echo "🧹 Cleaning up existing package..."
    rm -rf "$PACKAGE_DIR"
fi

# Create package directory
echo "📁 Creating package directory..."
mkdir -p "$PACKAGE_DIR"

# Copy essential files
echo "📋 Copying essential files..."
cp docker-compose.yml "$PACKAGE_DIR/"
cp nginx.conf "$PACKAGE_DIR/"
cp proxy_common.conf "$PACKAGE_DIR/"
cp deploy-anywhere.sh "$PACKAGE_DIR/"
cp deploy-everything.sh "$PACKAGE_DIR/"
cp add-ssl.sh "$PACKAGE_DIR/"

# Make scripts executable
chmod +x "$PACKAGE_DIR"/*.sh

# Create README for deployment
cat > "$PACKAGE_DIR"/README.md << 'EOF'
# Stellar Sync Deployment Package

Minimal deployment package for Stellar Sync server.

## Quick Deploy (Recommended)

### Option 1: Deploy Everything at Once
```bash
# Make scripts executable
chmod +x *.sh

# Deploy everything (Docker, containers, SSL)
./deploy-everything.sh your-domain.com your@email.com
```

### Option 2: Deploy Containers Only
```bash
# Deploy just the containers (if Docker already installed)
./deploy-anywhere.sh

# Add SSL separately
./add-ssl.sh your-domain.com your-email@example.com
```

## Files

- `docker-compose.yml` - Service configuration
- `nginx.conf` - Nginx configuration
- `proxy_common.conf` - Proxy settings
- `deploy-everything.sh` - Complete deployment script
- `deploy-anywhere.sh` - Container-only deployment
- `add-ssl.sh` - SSL setup script

## Requirements

- Root access (sudo)
- Domain pointing to server (for SSL)
- Internet connection (for Docker installation)

## What deploy-everything.sh Does

1. **Installs Docker** (Ubuntu/Debian/CentOS)
2. **Installs docker-compose**
3. **Installs certbot** (Let's Encrypt)
4. **Configures firewall** (UFW/firewalld)
5. **Deploys containers**
6. **Sets up SSL certificate**
7. **Tests everything**

## Management

```bash
docker-compose ps          # Check status
docker-compose logs        # View logs
docker-compose restart     # Restart services
docker-compose down        # Stop services
```
EOF

# Create package
echo "📦 Creating package..."
tar -czf "${PACKAGE_NAME}.tar.gz" "$PACKAGE_DIR"

# Clean up package directory
rm -rf "$PACKAGE_DIR"

echo ""
echo "====================================="
echo "✅ Deployment package created!"
echo "====================================="
echo ""
echo "📦 Package: ${PACKAGE_NAME}.tar.gz"
echo ""
echo "🚀 To deploy on a VM:"
echo "  1. Copy ${PACKAGE_NAME}.tar.gz to your VM"
echo "  2. Extract: tar -xzf ${PACKAGE_NAME}.tar.gz"
echo "  3. Run: ./deploy-everything.sh your-domain.com your@email.com"
echo ""
echo "📁 Package contains only essential deployment files"
echo "💾 No source code needed - containers pull from Docker Hub"
echo "🔧 deploy-everything.sh handles all dependencies automatically"

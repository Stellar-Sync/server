# Stellar Sync Deployment

Docker deployment for Stellar Sync server with Nginx and SSL.

## Quick Start

### Option 1: Deploy Everything at Once (Recommended)

```bash
# Copy deployment package to your VM
tar -xzf stellarsync-deployment.tar.gz
cd stellarsync-deployment

# Deploy everything (Docker, containers, SSL)
./deploy-everything.sh your-domain.com your@email.com
```

### Option 2: Deploy Containers Only

```bash
# Copy deployment package to your VM
tar -xzf stellarsync-deployment.tar.gz
cd stellarsync-deployment

# Deploy just the containers (if Docker already installed)
./deploy-anywhere.sh
```

### Option 3: Local Development

```bash
cd Docker/deploy
./manage.sh setup-storage
./manage.sh start
```

## What deploy-everything.sh Does

The comprehensive script handles everything automatically:

1. **🔧 Installs Dependencies**

   - Docker (Ubuntu/Debian/CentOS)
   - docker-compose
   - certbot (Let's Encrypt)

2. **🛡️ Configures Security**

   - Firewall setup (UFW/firewalld)
   - Port configuration (22, 80, 443)

3. **🐳 Deploys Services**

   - Pulls latest containers
   - Starts all services
   - Tests functionality

4. **🔒 Sets Up SSL**
   - Domain validation
   - Let's Encrypt certificate
   - Automatic renewal setup

## Services

- **Main Server** - WebSocket + HTTP proxy (port 6000)
- **File Server** - File uploads/downloads (port 6200)
- **Nginx** - Reverse proxy with SSL (ports 80, 443)

## Management

```bash
docker-compose ps          # Check status
docker-compose logs        # View logs
docker-compose restart     # Restart services
docker-compose storage     # Storage info
```

## SSL Setup

```bash
# Automatic with deploy-everything.sh
./deploy-everything.sh your-domain.com your@email.com

# Manual setup
./add-ssl.sh your-domain.com your@email.com
```

## Storage

- Files stored in Docker volume `stellarsync-data`
- Automatic cleanup after 6 hours
- SSL certificates in `stellarsync-ssl`
- Logs in `stellarsync-logs`

## Deployment Package

The `stellarsync-deployment.tar.gz` contains only essential files:

```
stellarsync-deployment/
├── docker-compose.yml      # Services
├── nginx.conf             # Nginx config
├── proxy_common.conf      # Proxy settings
├── deploy-everything.sh   # Complete deployment ✅
├── deploy-anywhere.sh     # Container-only deployment
├── add-ssl.sh             # SSL setup script
└── README.md              # This file
```

## Building & Pushing

From your development machine:

```bash
cd Docker/deploy
./build-and-push.sh        # Build and push to Docker Hub
./create-deployment-package.sh  # Create deployment package
```

## Requirements

- **Root access** (sudo)
- **Domain pointing** to server IP
- **Internet connection** (for Docker installation)
- **Supported OS**: Ubuntu, Debian, CentOS, RHEL

For development, see `../StellarSync/README.md`.

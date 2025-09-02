# Stellar Sync Server

Server implementation and deployment for Stellar Sync plugin.

## Structure

```
server/
├── StellarSync/              # Go server source code
├── Docker/                   # Deployment configuration
└── deploy.sh                 # Easy deployment script
```

## Quick Start

### Development

```bash
cd StellarSync
go build ./cmd/server
go run ./cmd/server
```

### Deployment

```bash
./deploy.sh start          # Start services
./deploy.sh status         # Check status
./deploy.sh setup-storage  # Setup storage
```

## What's Where

- **`StellarSync/`** - Go code for WebSocket and file servers
- **`Docker/deploy/`** - Docker Compose, Nginx, management scripts
- **`Docker/k8s/`** - Kubernetes manifests

## Documentation

- **Development**: `StellarSync/README.md`
- **Deployment**: `Docker/deploy/README.md`

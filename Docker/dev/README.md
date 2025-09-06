# StellarSync Development Docker Setup

This directory contains a simplified Docker setup for development that doesn't include nginx or SSL certificates.

## Quick Start

1. **Build and start the development environment:**

   ```bash
   cd server/Docker/dev
   docker-compose -f docker-compose.dev.yml up --build
   ```

2. **Access the services:**
   - Main Server: http://localhost:6000
   - File Server: http://localhost:6200
   - Health Check: http://localhost:6000/health

## Development Features

- **No nginx**: Direct access to services for easier debugging
- **Development tools**: Includes bash, vim, htop, net-tools in containers
- **Volume mounts**: Logs are mounted to `./logs` for easy access
- **Hot reload**: Rebuild containers when you make changes

## Useful Commands

### Start services

```bash
docker-compose -f docker-compose.dev.yml up --build
```

### Start in background

```bash
docker-compose -f docker-compose.dev.yml up -d --build
```

### View logs

```bash
docker-compose -f docker-compose.dev.yml logs -f
```

### Stop services

```bash
docker-compose -f docker-compose.dev.yml down
```

### Rebuild and restart

```bash
docker-compose -f docker-compose.dev.yml down
docker-compose -f docker-compose.dev.yml up --build
```

### Access container shell

```bash
# Main server
docker exec -it stellarsync-main-dev bash

# File server
docker exec -it stellarsync-fileserver-dev bash
```

### Clean up everything

```bash
docker-compose -f docker-compose.dev.yml down -v
docker system prune -f
```

## Configuration

The development setup uses:

- **Main Server**: Port 6000 (WebSocket + HTTP API)
- **File Server**: Port 6200 (File uploads/downloads)
- **Data Volume**: `stellarsync-dev-data` for persistent file storage
- **Logs**: Mounted to `./logs` directory

## Differences from Production

- No nginx reverse proxy
- No SSL/TLS certificates
- Development tools included in containers
- Logs mounted for easy access
- Simplified networking
- No health checks dependencies

## Troubleshooting

### Port conflicts

If ports 6000 or 6200 are already in use:

```bash
# Check what's using the ports
netstat -tulpn | grep :6000
netstat -tulpn | grep :6200

# Kill processes using those ports
sudo kill -9 <PID>
```

### Container won't start

```bash
# Check container logs
docker-compose -f docker-compose.dev.yml logs

# Check if images built correctly
docker images | grep stellarsync
```

### File permissions

```bash
# Fix log directory permissions
sudo chown -R $USER:$USER ./logs
```

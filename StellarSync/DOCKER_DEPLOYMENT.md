# Stellar Sync Server - Docker Cluster Deployment

## Overview

This deployment uses Docker Compose to run the Stellar Sync server as a cluster of two services:

- **Main Server** (port 6000) - Handles WebSocket connections and proxies file operations
- **File Server** (port 6200) - Handles file uploads, downloads, and metadata

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Main Server   │    │  File Server    │
│   (Port 6000)   │◄──►│  (Port 6200)    │
│                 │    │                 │
│ - WebSocket     │    │ - File Upload   │
│ - User Mgmt     │    │ - File Download │
│ - File Proxy    │    │ - Metadata      │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────────┐
│           Docker Network                │
│      (stellarsync-network)              │
└─────────────────────────────────────────┘
```

## Quick Start

### Local Development

1. **Build and start the cluster**:

   ```bash
   docker-compose up --build
   ```

2. **Access the services**:

   - Main Server: http://localhost:6000
   - File Server: http://localhost:6200
   - WebSocket: ws://localhost:6000/ws

3. **Stop the cluster**:
   ```bash
   docker-compose down
   ```

### Production Deployment

#### Option 1: Docker Compose on Server

1. **Copy files to your server**:

   ```bash
   scp -r server/StellarSync user@your-server:/opt/stellarsync/
   ```

2. **SSH to your server and start**:

   ```bash
   cd /opt/stellarsync
   docker-compose up -d --build
   ```

3. **Set up reverse proxy** (nginx example):
   ```nginx
   server {
       listen 80;
       server_name your-domain.com;

       location / {
           proxy_pass http://localhost:6000;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection 'upgrade';
           proxy_set_header Host $host;
           proxy_cache_bypass $http_upgrade;
       }
   }
   ```

#### Option 2: Koyeb with Docker Compose

1. **Create a Koyeb app with Docker Compose**:
   ```bash
   koyeb app init stellarsync-cluster \
     --docker-compose docker-compose.yml \
     --ports 6000:http \
     --routes /:6000
   ```

#### Option 3: Kubernetes Deployment

1. **Create Kubernetes manifests**:
   ```yaml
   # main-server-deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: stellarsync-main
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: stellarsync-main
     template:
       metadata:
         labels:
           app: stellarsync-main
       spec:
         containers:
           - name: main-server
             image: stellarsync-main:latest
             ports:
               - containerPort: 6000
             env:
               - name: FILE_SERVER_URL
                 value: 'http://stellarsync-files:6200'
   ```

## Configuration

### Environment Variables

| Service     | Variable          | Description      | Default                   |
| ----------- | ----------------- | ---------------- | ------------------------- |
| Main Server | `PORT`            | Main server port | `6000`                    |
| Main Server | `FILE_SERVER_URL` | File server URL  | `http://file-server:6200` |
| File Server | `PORT`            | File server port | `6200`                    |

### Volumes

- **file-storage**: Persistent volume for uploaded files
  - Location: `/app/uploads` in file server container
  - Persists across container restarts

### Networks

- **stellarsync-network**: Internal bridge network for service communication
  - Main server can reach file server at `http://file-server:6200`
  - Isolated from external traffic

## Monitoring

### Health Checks

Both services include health check endpoints:

- **Main Server**: http://localhost:6000/health
- **File Server**: http://localhost:6200/health

### Logs

```bash
# View all logs
docker-compose logs

# View specific service logs
docker-compose logs main-server
docker-compose logs file-server

# Follow logs in real-time
docker-compose logs -f
```

### Status Pages

- **Main Server**: http://localhost:6000/
- **File Server**: http://localhost:6200/

## Scaling

### Horizontal Scaling

```bash
# Scale main server (stateless)
docker-compose up -d --scale main-server=3

# Scale file server (requires shared storage)
docker-compose up -d --scale file-server=2
```

### Load Balancing

For production, add a load balancer:

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - '80:80'
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - main-server
```

## Troubleshooting

### Common Issues

1. **File server not reachable**:

   ```bash
   # Check if file server is running
   docker-compose ps

   # Check file server logs
   docker-compose logs file-server

   # Test connectivity from main server
   docker-compose exec main-server curl http://file-server:6200/health
   ```

2. **File uploads failing**:

   ```bash
   # Check file storage volume
   docker-compose exec file-server ls -la /app/uploads

   # Check volume permissions
   docker-compose exec file-server chmod 755 /app/uploads
   ```

3. **WebSocket connections failing**:

   ```bash
   # Check main server logs
   docker-compose logs main-server

   # Verify WebSocket endpoint
   curl -I http://localhost:6000/ws
   ```

### Debug Mode

```bash
# Run with debug logging
docker-compose up --build -e DEBUG=true
```

## Security Considerations

1. **Network Isolation**: Services communicate only through internal Docker network
2. **File Storage**: Files stored in Docker volume, isolated from host
3. **Health Checks**: Regular health monitoring prevents serving unhealthy containers
4. **Restart Policy**: Automatic restart on failure

## Performance Optimization

1. **Resource Limits**:

   ```yaml
   services:
     main-server:
       deploy:
         resources:
           limits:
             memory: 512M
             cpus: '0.5'
   ```

2. **Caching**: Add Redis for session caching
3. **CDN**: Use CDN for file downloads in production
4. **Database**: Replace in-memory storage with persistent database

## Backup and Recovery

### File Storage Backup

```bash
# Backup uploaded files
docker run --rm -v stellarsync_file-storage:/data -v $(pwd):/backup alpine tar czf /backup/files-backup.tar.gz -C /data .

# Restore files
docker run --rm -v stellarsync_file-storage:/data -v $(pwd):/backup alpine tar xzf /backup/files-backup.tar.gz -C /data
```

### Configuration Backup

```bash
# Backup configuration
docker-compose config > docker-compose.backup.yml
```

## Support

For issues with:

- **Docker**: Check Docker documentation
- **Docker Compose**: Check Compose documentation
- **Stellar Sync**: Check the main repository issues

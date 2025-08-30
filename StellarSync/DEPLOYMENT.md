# Stellar Sync Server - Koyeb Deployment Guide

## Overview

This guide explains how to deploy the Stellar Sync server on Koyeb, a cloud platform that supports Go applications.

## Architecture

The server combines both the WebSocket server (for real-time communication) and HTTP file server (for file transfers) into a single application for easier deployment.

### Endpoints

- **WebSocket**: `/ws` - Real-time communication between clients
- **File Upload**: `/upload` - Upload mod files
- **File Download**: `/download/{hash}` - Download mod files
- **File List**: `/files` - List available files
- **Metadata Upload**: `/metadata/upload` - Upload file metadata
- **Metadata Download**: `/metadata/download/{user_id}` - Download file metadata
- **Health Check**: `/` - Server status

## Deployment Options

### Option 1: Koyeb CLI (Recommended)

1. **Install Koyeb CLI**:

   ```bash
   # macOS
   brew install koyeb/tap/cli

   # Windows
   scoop install koyeb

   # Linux
   curl -fsSL https://cli.koyeb.com/install.sh | bash
   ```

2. **Login to Koyeb**:

   ```bash
   koyeb login
   ```

3. **Deploy from Git**:
   ```bash
   koyeb app init stellarsync-server \
     --git github.com/your-username/Stellar-Sync \
     --git-branch main \
     --git-working-dir server/StellarSync \
     --ports 6000:http \
     --routes /:6000 \
     --env PORT=6000 \
     --build-command "go build -o stellarsync-server ./cmd/combined" \
     --run-command "./stellarsync-server"
   ```

### Option 2: Koyeb Dashboard (Easiest)

1. **Go to [Koyeb Dashboard](https://app.koyeb.com/)**
2. **Click "Create App"**
3. **Select "GitHub" as source**
4. **Configure the deployment**:
   - **Repository**: Your Stellar Sync repository
   - **Branch**: `main`
   - **Working Directory**: `server/StellarSync`
   - **Build Command**: `go build -o stellarsync-server ./cmd/combined`
   - **Run Command**: `./stellarsync-server`
   - **Port**: `6000`
   - **Environment Variables**:
     - `PORT`: `6000`

### Option 3: Koyeb CLI with App Manifest

1. **Create app manifest** (use the provided `koyeb.yaml`):
   ```bash
   koyeb app init stellarsync-server --manifest koyeb.yaml
   ```

### Option 4: Docker Deployment

1. **Build and push Docker image**:

   ```bash
   docker build -t stellarsync-server .
   docker tag stellarsync-server your-registry/stellarsync-server:latest
   docker push your-registry/stellarsync-server:latest
   ```

2. **Deploy on Koyeb**:
   ```bash
   koyeb app init stellarsync-server \
     --docker your-registry/stellarsync-server:latest \
     --ports 6000:http \
     --routes /:6000 \
     --env PORT=6000
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT`   | Server port | `6000`  |

### Resource Requirements

- **CPU**: 0.5 cores (minimum)
- **Memory**: 512MB (minimum)
- **Storage**: 1GB (for file uploads)

### Scaling

The server supports horizontal scaling:

- **Minimum instances**: 1
- **Maximum instances**: 3 (or more based on load)

## Client Configuration

After deployment, update your client configuration:

```csharp
// In your client code, update the server URL
private const string SERVER_URL = "https://your-app-name.koyeb.app";
private const string WS_URL = "wss://your-app-name.koyeb.app/ws";
```

## Monitoring

### Health Checks

The server includes a health check endpoint at `/` that returns:

```json
{
  "status": "running",
  "service": "Stellar Sync Server",
  "time": 1234567890,
  "endpoints": {
    "websocket": "/ws",
    "upload": "/upload",
    "download": "/download/",
    "files": "/files",
    "metadata_upload": "/metadata/upload",
    "metadata_download": "/metadata/download/"
  }
}
```

### Logs

Access logs through the Koyeb dashboard:

1. Go to your app
2. Click "Logs" tab
3. View real-time application logs

## Troubleshooting

### Common Issues

1. **Port binding errors**:

   - Ensure `PORT` environment variable is set
   - Check that the application binds to `0.0.0.0:PORT`

2. **File upload failures**:

   - Verify uploads directory exists
   - Check file size limits (32MB default)

3. **WebSocket connection issues**:
   - Ensure client uses `wss://` for secure connections
   - Check CORS settings if needed

### Debug Mode

To enable debug logging, add environment variable:

```bash
--env DEBUG=true
```

## Security Considerations

1. **HTTPS/WSS**: Koyeb automatically provides SSL certificates
2. **File Storage**: Files are stored in memory/disk - consider external storage for production
3. **Rate Limiting**: Consider implementing rate limiting for file uploads
4. **Authentication**: Add authentication if needed for production use

## Cost Optimization

- **Free Tier**: Koyeb offers a free tier with limitations
- **Scaling**: Use auto-scaling based on CPU/memory usage
- **Storage**: Consider external storage for large file volumes

## Support

For issues with:

- **Koyeb Platform**: [Koyeb Support](https://www.koyeb.com/support)
- **Stellar Sync**: Check the main repository issues

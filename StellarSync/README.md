# Stellar Sync Server

This is the server component for Stellar Sync, a FFXIV mod synchronization plugin.

## Architecture

The server is split into two components for better scalability and separation of concerns:

### Main Server (Port 6000)

- **Purpose**: WebSocket connections, user management, message routing
- **Endpoints**:
  - `/ws` - WebSocket endpoint for client connections
  - `/health` - Health check
  - `/` - Status page
  - `/upload`, `/download`, `/list`, `/metadata/*` - Proxied to file server

### File Server (Port 6200)

- **Purpose**: Dedicated file storage and transfer operations
- **Endpoints**:
  - `/upload` - File upload
  - `/download?hash=...` - File download
  - `/list` - List all files
  - `/metadata/upload` - Upload file metadata
  - `/metadata/download?user_id=...` - Download file metadata
  - `/health` - Health check
  - `/` - Status page

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git

### Installation

1. Clone the repository
2. Navigate to the server directory: `cd server/StellarSync`
3. Run the startup script: `start_servers.bat`

### Manual Startup

If you prefer to start servers manually:

1. **Start File Server** (in one terminal):

   ```bash
   go run cmd/fileserver/main.go
   ```

2. **Start Main Server** (in another terminal):
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

### Ports

- Main Server: 6000
- File Server: 6200

### File Storage

- Files are stored in `./uploads/` directory (created automatically)
- File metadata is stored in memory (will be replaced with database in production)

## Development

### Project Structure

```
server/StellarSync/
├── cmd/
│   ├── server/          # Main server application
│   └── fileserver/      # File server application
├── internal/
│   ├── models/          # Shared data models
│   ├── websocket/       # WebSocket handling
│   └── proxy/           # File proxy service
└── start_servers.bat    # Startup script
```

### Adding New Features

1. **WebSocket Messages**: Add handlers in `cmd/server/main.go`
2. **File Operations**: Add endpoints in `cmd/fileserver/main.go`
3. **Data Models**: Add structures in `internal/models/`

## Production Considerations

- Replace in-memory storage with proper database
- Add authentication and authorization
- Implement file cleanup and storage limits
- Add monitoring and logging
- Use environment variables for configuration
- Implement proper error handling and recovery

## API Reference

### WebSocket Messages

- `connect` - Client connection
- `character_data` - Character data synchronization
- `request_users` - Get online users list
- `request_user_data` - Request specific user's data

### HTTP Endpoints

All file-related endpoints are proxied from main server to file server transparently.

## Troubleshooting

### Common Issues

1. **Port already in use**: Check if another instance is running
2. **File server not accessible**: Ensure file server is started before main server
3. **WebSocket connection failed**: Check firewall settings

### Logs

Both servers provide detailed logging for debugging:

- Main server logs: WebSocket connections, message handling
- File server logs: File operations, metadata management

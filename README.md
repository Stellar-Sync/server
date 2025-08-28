# Stellar Sync Server

A clean, scalable Go WebSocket server for the Stellar Sync plugin.

## Architecture

The server follows clean architecture principles with a clear separation of concerns:

```
server/
├── StellarSync/           # Main server code
│   ├── cmd/              # Application entry points
│   │   └── server/       # Main server binary
│   ├── internal/         # Private application code
│   │   ├── models/       # Data models and structures
│   │   ├── websocket/    # WebSocket server implementation
│   │   └── handlers/     # Message handling logic
│   └── pkg/              # Public libraries (future use)
├── build.bat             # Build script
├── start-server.bat      # Development server script
└── README.md             # This file
```

## Features

- **Clean Architecture**: Separation of concerns with proper layering
- **WebSocket Server**: Real-time communication for character sync
- **JSON Message Protocol**: Structured message handling
- **Client Management**: Connection tracking and broadcasting
- **Health Monitoring**: Health check endpoint
- **Scalable Design**: Easy to extend with new features

## Quick Start

### Prerequisites

- Go 1.21 or later

### Installation

1. Clone the repository
2. Navigate to the server directory
3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Server

#### Development Mode

```bash
start-server.bat
```

#### Build and Run

```bash
build.bat
stellarsync-server.exe
```

The server will start on port 6000 by default.

### Endpoints

- **WebSocket**: `ws://localhost:6000/ws`
- **Health Check**: `http://localhost:6000/health`
- **Status Page**: `http://localhost:6000/`

## Message Protocol

The server uses JSON messages with the following structure:

```json
{
  "type": "message_type",
  "client": "client_identifier",
  "data": {},
  "error": "error_message"
}
```

### Supported Message Types

- `connect` - Client connection message
- `character_data` - Character data from client
- `connected` - Server response to connection
- `character_data_received` - Server acknowledgment of character data
- `character_data_broadcast` - Broadcast of character data to other clients
- `error` - Error message

## Development

### Project Structure

- **`cmd/server/`**: Main application entry point
- **`internal/models/`**: Data structures and message definitions
- **`internal/websocket/`**: WebSocket server implementation
- **`internal/handlers/`**: Message processing logic

### Adding New Features

1. **New Message Types**: Add to `internal/models/message.go`
2. **New Handlers**: Add to `cmd/server/main.go` or create new handler files
3. **New Endpoints**: Add to the main function in `cmd/server/main.go`

### Building

```bash
build.bat
```

### Testing

```bash
cd StellarSync
go test ./...
```

## Configuration

The server currently runs on port 6000. To change the port, modify the `port` variable in `cmd/server/main.go`.

## License

See LICENSE file for details.

# Stellar Sync Server

Go server for Stellar Sync plugin - WebSocket communication and file handling.

## Quick Start

### Build

```bash
go build ./cmd/server      # Main server
go build ./cmd/fileserver  # File server
```

### Run

```bash
go run ./cmd/server        # Main server (port 6000)
go run ./cmd/fileserver    # File server (port 6200)
```

## Structure

- `cmd/server/` - WebSocket server + HTTP proxy
- `cmd/fileserver/` - File upload/download with 6-hour TTL
- `internal/` - Business logic and models
- `pkg/` - Public packages

## Configuration

- `STORAGE_PATH` - File storage directory (default: `/app/files`)
- `FILE_SERVER_URL` - File server URL for proxy

## API

- `GET /health` - Health check
- `GET /ws` - WebSocket endpoint
- `POST /upload` - File upload
- `GET /download?hash=<hash>` - File download

## Docker

See `../Docker/deploy/README.md` for deployment.

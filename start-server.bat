@echo off
echo Starting Stellar Sync Server...
echo.
echo Server will be available at:
echo - WebSocket: ws://localhost:6000/ws
echo - Health Check: http://localhost:6000/health
echo - Status Page: http://localhost:6000/
echo.
echo Press Ctrl+C to stop the server
echo.

cd StellarSync
go run ./cmd/server

pause

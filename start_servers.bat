@echo off
echo =====================================
echo Starting Stellar Sync Servers...
echo =====================================

echo Starting File Server (port 6200)...
start "Stellar Sync File Server" cmd /k "cd /d %~dp0StellarSync && go run cmd/fileserver/main.go"

echo Waiting 2 seconds for file server to start...
timeout /t 2 /nobreak > nul

echo Starting Main Server (port 6000)...
start "Stellar Sync Main Server" cmd /k "cd /d %~dp0StellarSync && go run cmd/server/main.go"

echo =====================================
echo Both servers started!
echo =====================================
echo Main Server: http://localhost:6000
echo File Server: http://localhost:6200
echo =====================================
pause

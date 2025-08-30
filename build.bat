@echo off
echo Building Stellar Sync Server...
echo.

cd StellarSync
go build -o ../stellarsync-server.exe ./cmd/server

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Server executable: stellarsync-server.exe
) else (
    echo.
    echo Build failed!
)

pause

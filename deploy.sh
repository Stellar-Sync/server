#!/bin/bash

# 🚀 Stellar Sync Deployment Script
# This script provides easy access to deployment commands from the server root

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -d "Docker/deploy" ]; then
    print_error "This script must be run from the server root directory"
    print_error "Expected structure: server/StellarSync/ and server/Docker/"
    exit 1
fi

# Change to deployment directory and run command
case "$1" in
    "start"|"stop"|"restart"|"logs"|"status"|"update"|"backup"|"ssl-renew"|"build"|"clean"|"storage"|"cleanup-now"|"setup-storage")
        print_status "Running command: $1"
        cd Docker/deploy
        ./manage.sh "$1"
        ;;
    "help"|"--help"|"-h"|"")
        echo "🚀 Stellar Sync Deployment Script"
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo "Commands:"
        echo "  start         - Start all services"
        echo "  stop          - Stop all services"
        echo "  restart       - Restart all services"
        echo "  logs          - Show service logs"
        echo "  status        - Show service status"
        echo "  update        - Update and restart services"
        echo "  backup        - Create backup of configuration"
        echo "  ssl-renew     - Renew SSL certificate"
        echo "  build         - Build services from source"
        echo "  clean         - Clean up containers and volumes"
        echo "  storage       - Show storage information"
        echo "  cleanup-now   - Trigger immediate file cleanup"
        echo "  setup-storage - Set up external storage directory"
        echo ""
        echo "Examples:"
        echo "  $0 start      # Start all services"
        echo "  $0 status     # Check service status"
        echo "  $0 logs       # View logs"
        echo ""
        echo "Note: This script runs commands from the Docker/deploy directory"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for available commands"
        exit 1
        ;;
esac

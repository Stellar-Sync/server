#!/bin/bash

# 🚀 Stellar Sync Management Script

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

case "$1" in
    "start")
        print_status "🚀 Starting Stellar Sync..."
        docker-compose up -d
        print_success "✅ Services started!"
        ;;
    "stop")
        print_status "🛑 Stopping Stellar Sync..."
        docker-compose down
        print_success "✅ Services stopped!"
        ;;
    "restart")
        print_status "🔄 Restarting Stellar Sync..."
        docker-compose restart
        print_success "✅ Services restarted!"
        ;;
    "logs")
        print_status "📋 Showing logs..."
        docker-compose logs -f
        ;;
    "status")
        print_status "📊 Service status:"
        docker-compose ps
        ;;
    "update")
        print_status "📥 Updating Stellar Sync..."
        docker-compose pull
        docker-compose up -d
        print_success "✅ Services updated!"
        ;;
    "backup")
        print_status "💾 Creating backup..."
        tar -czf "stellarsync-backup-$(date +%Y%m%d-%H%M%S).tar.gz" \
            docker-compose.yml nginx.conf ssl/ logs/
        print_success "✅ Backup created!"
        ;;
    "ssl-renew")
        print_status "🔒 Renewing SSL certificate..."
        if [[ $EUID -eq 0 ]]; then
            certbot renew
            cp "/etc/letsencrypt/live/$(grep -o 'server_name [^;]*' nginx.conf | awk '{print $2}')/fullchain.pem" ssl/
            cp "/etc/letsencrypt/live/$(grep -o 'server_name [^;]*' nginx.conf | awk '{print $2}')/privkey.pem" ssl/
            chown root:root ssl/*
        else
            sudo certbot renew
            sudo cp "/etc/letsencrypt/live/$(grep -o 'server_name [^;]*' nginx.conf | awk '{print $2}')/fullchain.pem" ssl/
            sudo cp "/etc/letsencrypt/live/$(grep -o 'server_name [^;]*' nginx.conf | awk '{print $2}')/privkey.pem" ssl/
            sudo chown $(whoami):$(whoami) ssl/*
        fi
        docker-compose restart nginx
        print_success "✅ SSL renewed!"
        ;;
    "build")
        print_status "🔨 Building services..."
        docker-compose build
        print_success "✅ Services built!"
        ;;
    "clean")
        print_status "🧹 Cleaning up..."
        docker-compose down -v
        docker system prune -f
        print_success "✅ Cleanup complete!"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|logs|status|update|backup|ssl-renew|build|clean}"
        echo ""
        echo "Commands:"
        echo "  start      - Start all services"
        echo "  stop       - Stop all services"
        echo "  restart    - Restart all services"
        echo "  logs       - Show service logs"
        echo "  status     - Show service status"
        echo "  update     - Update and restart services"
        echo "  backup     - Create backup of configuration"
        echo "  ssl-renew  - Renew SSL certificate"
        echo "  build      - Build services from source"
        echo "  clean      - Clean up containers and volumes"
        exit 1
        ;;
esac

#!/bin/bash

# 🔒 SSL Setup for Stellar Sync
# This script sets up SSL certificates for the nginx container

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

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root. Please run as a regular user with sudo privileges."
   exit 1
fi

print_status "🔒 Stellar Sync - SSL Setup"
print_status "=========================="

# Get configuration
if [[ -z "$1" || -z "$2" ]]; then
    echo "Usage: $0 <domain> <email>"
    echo ""
    echo "Example: $0 stellar.kasu.network your-email@example.com"
    exit 1
fi

DOMAIN="$1"
EMAIL="$2"

print_status "Configuration:"
echo "  Domain: $DOMAIN"
echo "  Email: $EMAIL"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

# Check if services are running
if ! docker-compose ps | grep -q "stellarsync-nginx"; then
    print_error "Nginx container is not running. Please start services first:"
    echo "  docker-compose up -d"
    exit 1
fi

# Create SSL directory
mkdir -p ssl logs

# Install certbot if not installed
if ! command -v certbot &> /dev/null; then
    print_status "Installing certbot..."
    sudo apt update
    sudo apt install -y certbot
fi

# Stop nginx temporarily
print_status "Stopping nginx container..."
docker-compose stop nginx

# Get SSL certificate
print_status "Getting SSL certificate for $DOMAIN..."
sudo certbot certonly --standalone -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive

# Copy certificates
print_status "Copying certificates..."
sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ssl/
sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ssl/
sudo chown $(whoami):$(whoami) ssl/*

# Start nginx
print_status "Starting nginx container..."
docker-compose up -d nginx

print_success "✅ SSL setup complete!"
echo ""
echo "📋 Certificate details:"
echo "  Domain: $DOMAIN"
echo "  Expires: $(sudo certbot certificates | grep -A 2 "$DOMAIN" | grep "VALID" | awk '{print $2}')"
echo "  Auto-renewal: Every 60 days"
echo ""
echo "🔗 Your server is now available at:"
echo "  - WebSocket: wss://$DOMAIN/ws"
echo "  - File uploads: https://$DOMAIN/upload"
echo "  - Health check: https://$DOMAIN/health"
echo ""
echo "📝 To renew certificates manually:"
echo "  sudo certbot renew"
echo "  sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ssl/"
echo "  sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ssl/"
echo "  sudo chown $(whoami):$(whoami) ssl/*"
echo "  docker-compose restart nginx"

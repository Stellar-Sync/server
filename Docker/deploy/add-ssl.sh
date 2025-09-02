#!/bin/bash

# 🔒 Add SSL Certificate for Stellar Sync
# This script sets up SSL using Let's Encrypt

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <domain> <email>"
    echo "Example: $0 stellar.yourdomain.com your@email.com"
    exit 1
fi

DOMAIN=$1
EMAIL=$2

echo "====================================="
echo "Setting up SSL for Stellar Sync"
echo "====================================="
echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ This script must be run as root (use sudo)"
    exit 1
fi

# Check if domain is reachable
echo "🔍 Checking if domain $DOMAIN points to this server..."
if ! nslookup $DOMAIN > /dev/null 2>&1; then
    echo "❌ Cannot resolve domain $DOMAIN"
    echo "Make sure your domain points to this server's IP address"
    exit 1
fi

# Install certbot if not present
if ! command -v certbot > /dev/null 2>&1; then
    echo "📥 Installing certbot..."
    apt update
    apt install -y certbot
fi

# Stop nginx temporarily
echo "🛑 Stopping nginx temporarily..."
docker-compose stop nginx

# Get SSL certificate
echo "🔒 Getting SSL certificate from Let's Encrypt..."
certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --non-interactive

# Create SSL directory in container volume
echo "📁 Setting up SSL directory..."
docker run --rm -v stellarsync-ssl:/ssl alpine sh -c "mkdir -p /ssl"

# Copy certificates to container volume
echo "📋 Copying certificates..."
docker run --rm -v stellarsync-ssl:/ssl -v /etc/letsencrypt/live/$DOMAIN:/certs alpine sh -c "cp /certs/fullchain.pem /ssl/ && cp /certs/privkey.pem /ssl/"

# Set proper permissions
echo "🔐 Setting permissions..."
docker run --rm -v stellarsync-ssl:/ssl alpine sh -c "chmod 644 /ssl/*.pem"

# Start nginx with SSL
echo "🚀 Starting nginx with SSL..."
docker-compose up -d nginx

# Test SSL
echo "🧪 Testing SSL setup..."
sleep 5
if curl -k -s "https://$DOMAIN/health" > /dev/null; then
    echo "✅ SSL setup successful!"
    echo "Your server is now available at: https://$DOMAIN"
else
    echo "⚠️ SSL setup may have issues. Check logs with: docker-compose logs nginx"
fi

echo ""
echo "====================================="
echo "SSL Setup Complete!"
echo "====================================="
echo ""
echo "🔒 Your server now has SSL enabled"
echo "🌐 Access via: https://$DOMAIN"
echo ""
echo "📋 To renew certificates automatically:"
echo "  echo '0 12 * * * /usr/bin/certbot renew --quiet' | crontab -"
echo ""
echo "📊 Check status: docker-compose ps"
echo "📝 View logs: docker-compose logs nginx"

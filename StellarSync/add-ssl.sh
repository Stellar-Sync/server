#!/bin/bash

# 🔒 Add SSL to Stellar Sync
# This script adds SSL configuration to an existing nginx setup

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
   print_warning "Running as root. This is not recommended for security, but the script will continue."
   print_warning "Consider creating a regular user account for future use."
   echo ""
fi

print_status "🔒 Adding SSL to Stellar Sync"
print_status "============================"

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
    if [[ $EUID -eq 0 ]]; then
        apt update
        apt install -y certbot
    else
        sudo apt update
        sudo apt install -y certbot
    fi
fi

# Stop all containers and clean up
print_status "Stopping all containers..."
docker-compose down

# Remove the nginx container specifically to avoid recreation issues
print_status "Removing nginx container..."
docker rm -f stellarsync-nginx 2>/dev/null || true

# Clean up any dangling containers
print_status "Cleaning up containers..."
docker container prune -f

# Get SSL certificate
print_status "Getting SSL certificate for $DOMAIN..."
if [[ $EUID -eq 0 ]]; then
    certbot certonly --standalone -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive
else
    sudo certbot certonly --standalone -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive
fi

# Copy certificates
print_status "Copying certificates..."
if [[ $EUID -eq 0 ]]; then
    cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ssl/
    cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ssl/
    chown root:root ssl/*
else
    sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ssl/
    sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ssl/
    sudo chown $(whoami):$(whoami) ssl/*
fi

# Create SSL-enabled nginx config
print_status "Creating SSL-enabled nginx configuration..."
cat > nginx-ssl.conf << EOF
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    # Logging
    log_format main '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                    '\$status \$body_bytes_sent "\$http_referer" '
                    '"\$http_user_agent" "\$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    # Performance
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 100M;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Upstream servers
    upstream main-server {
        server main-server:6000;
        keepalive 32;
    }

    upstream file-server {
        server file-server:6200;
        keepalive 32;
    }

    # HTTP server (redirect to HTTPS)
    server {
        listen 80;
        server_name $DOMAIN;
        
        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # Redirect everything else to HTTPS
        location / {
            return 301 https://\$server_name\$request_uri;
        }
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name $DOMAIN;

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_session_timeout 1d;
        ssl_session_cache shared:SSL:50m;
        ssl_session_tickets off;

        # Modern SSL configuration
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header Strict-Transport-Security "max-age=63072000" always;
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # WebSocket support
        location /ws {
            proxy_pass http://main-server;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            proxy_read_timeout 86400;
            proxy_send_timeout 86400;
        }

        # File server endpoints
        location /upload {
            proxy_pass http://file-server;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            proxy_read_timeout 300;
            proxy_send_timeout 300;
        }

        location /download {
            proxy_pass http://file-server;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            proxy_read_timeout 300;
            proxy_send_timeout 300;
        }

        location /list {
            proxy_pass http://file-server;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        location /metadata {
            proxy_pass http://file-server;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        # Main server (fallback)
        location / {
            proxy_pass http://main-server;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
    }
}
EOF

# Update docker-compose.yml to include SSL volume
print_status "Updating docker-compose.yml to include SSL volume..."
sed -i 's|- ./logs:/var/log/nginx|- ./ssl:/etc/nginx/ssl:ro\n      - ./logs:/var/log/nginx|' docker-compose.yml

# Start all services fresh
print_status "Starting all services..."
docker-compose up -d

# Wait a moment for services to start
sleep 5

# Check if services are running
print_status "Checking service status..."
docker-compose ps

print_success "✅ SSL setup complete!"
echo ""
echo "📋 Certificate details:"
echo "  Domain: $DOMAIN"
if [[ $EUID -eq 0 ]]; then
    echo "  Expires: $(certbot certificates | grep -A 2 "$DOMAIN" | grep "VALID" | awk '{print $2}')"
else
    echo "  Expires: $(sudo certbot certificates | grep -A 2 "$DOMAIN" | grep "VALID" | awk '{print $2}')"
fi
echo "  Auto-renewal: Every 60 days"
echo ""
echo "🔗 Your server is now available at:"
echo "  - WebSocket: wss://$DOMAIN/ws"
echo "  - File uploads: https://$DOMAIN/upload"
echo "  - Health check: https://$DOMAIN/health"
echo ""
echo "📝 To renew certificates manually:"
if [[ $EUID -eq 0 ]]; then
    echo "  certbot renew"
    echo "  cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ssl/"
    echo "  cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ssl/"
    echo "  chown root:root ssl/*"
    echo "  docker-compose restart nginx"
else
    echo "  sudo certbot renew"
    echo "  sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ssl/"
    echo "  sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ssl/"
    echo "  sudo chown $(whoami):$(whoami) ssl/*"
    echo "  docker-compose restart nginx"
fi

#!/bin/bash

# 🚀 Stellar Sync - Complete Deployment Script
# This script does everything: installs dependencies, deploys containers, and sets up SSL

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
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Check arguments
if [ $# -lt 2 ]; then
    echo "Usage: $0 <domain> <email> [skip-ssl]"
    echo "Example: $0 stellar.yourdomain.com your@email.com"
    echo "Use 'skip-ssl' as third argument to skip SSL setup"
    exit 1
fi

DOMAIN=$1
EMAIL=$2
SKIP_SSL=${3:-false}

echo "====================================="
echo "Stellar Sync - Complete Deployment"
echo "====================================="
echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo "Skip SSL: $SKIP_SSL"
echo ""

# Function to detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "Cannot detect OS"
        exit 1
    fi
}

# Function to install Docker
install_docker() {
    print_status "Installing Docker..."
    
    if command -v docker > /dev/null 2>&1; then
        print_success "Docker already installed"
        return
    fi
    
    case $OS in
        "Ubuntu"|"Debian GNU/Linux")
            apt update
            apt install -y apt-transport-https ca-certificates curl gnupg lsb-release
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
            apt update
            apt install -y docker-ce docker-ce-cli containerd.io
            ;;
        "CentOS Linux"|"Red Hat Enterprise Linux")
            yum install -y yum-utils
            yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            yum install -y docker-ce docker-ce-cli containerd.io
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    systemctl start docker
    systemctl enable docker
    print_success "Docker installed and started"
}

# Function to install docker-compose
install_docker_compose() {
    print_status "Installing docker-compose..."
    
    if command -v docker-compose > /dev/null 2>&1; then
        print_success "docker-compose already installed"
        return
    fi
    
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    print_success "docker-compose installed"
}

# Function to install certbot
install_certbot() {
    print_status "Installing certbot..."
    
    if command -v certbot > /dev/null 2>&1; then
        print_success "certbot already installed"
        return
    fi
    
    case $OS in
        "Ubuntu"|"Debian GNU/Linux")
            apt update
            apt install -y certbot
            ;;
        "CentOS Linux"|"Red Hat Enterprise Linux")
            yum install -y epel-release
            yum install -y certbot
            ;;
        *)
            print_error "Cannot install certbot on $OS"
            exit 1
            ;;
    esac
    
    print_success "certbot installed"
}

# Function to check domain
check_domain() {
    print_status "Checking if domain $DOMAIN points to this server..."
    
    # Get server's public IP
    SERVER_IP=$(curl -s ifconfig.me)
    print_status "Server public IP: $SERVER_IP"
    
    # Check domain resolution
    DOMAIN_IP=$(nslookup $DOMAIN | grep -A1 "Name:" | tail -1 | awk '{print $2}')
    
    if [ -z "$DOMAIN_IP" ]; then
        print_error "Cannot resolve domain $DOMAIN"
        print_error "Make sure your domain points to this server's IP: $SERVER_IP"
        exit 1
    fi
    
    print_success "Domain $DOMAIN resolves to: $DOMAIN_IP"
    
    if [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
        print_warning "Domain IP ($DOMAIN_IP) doesn't match server IP ($SERVER_IP)"
        print_warning "SSL setup may fail if domain doesn't point here"
    fi
}

# Function to deploy containers
deploy_containers() {
    print_status "Deploying Stellar Sync containers..."
    
    # Create necessary directories
    mkdir -p ssl logs
    
    # Pull latest images
    print_status "Pulling latest images..."
    docker-compose pull
    
    # Start services
    print_status "Starting services..."
    docker-compose up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 20
    
    # Check status
    print_status "Checking service status..."
    docker-compose ps
    
    # Test basic functionality
    print_status "Testing basic functionality..."
    if curl -s http://localhost/health > /dev/null; then
        print_success "Services are running and responding"
    else
        print_warning "Services may not be fully ready yet"
    fi
}

# Function to setup SSL
setup_ssl() {
    if [ "$SKIP_SSL" = "skip-ssl" ]; then
        print_warning "Skipping SSL setup as requested"
        return
    fi
    
    print_status "Setting up SSL certificate..."
    
    # Stop nginx temporarily
    print_status "Stopping nginx temporarily..."
    docker-compose stop nginx
    
    # Get SSL certificate
    print_status "Getting SSL certificate from Let's Encrypt..."
    certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --non-interactive
    
    # Verify certificate files exist
    CERT_PATH="/etc/letsencrypt/live/$DOMAIN"
    ARCHIVE_PATH="/etc/letsencrypt/archive/$DOMAIN"
    
    if [ ! -f "$CERT_PATH/fullchain.pem" ] || [ ! -f "$CERT_PATH/privkey.pem" ]; then
        print_error "SSL certificate files not found at $CERT_PATH"
        print_error "SSL setup failed. Please check certbot logs."
        return 1
    fi
    
    print_status "SSL certificate files found:"
    ls -la "$CERT_PATH/"
    
    # Find the actual certificate files (not symlinks)
    FULLCHAIN_FILE=$(readlink -f "$CERT_PATH/fullchain.pem")
    PRIVKEY_FILE=$(readlink -f "$CERT_PATH/privkey.pem")
    
    print_status "Actual certificate files:"
    print_status "Fullchain: $FULLCHAIN_FILE"
    print_status "Privkey: $PRIVKEY_FILE"
    
    # Get the project name from current directory (this is what docker-compose uses for volume prefix)
    PROJECT_NAME=$(basename "$(pwd)")
    SSL_VOLUME_NAME="${PROJECT_NAME}_stellarsync-ssl"
    
    print_status "Using SSL volume: $SSL_VOLUME_NAME"
    
    # Create SSL directory in container volume (matching nginx's expected path)
    print_status "Setting up SSL directory..."
    docker run --rm -v "$SSL_VOLUME_NAME:/etc/nginx/ssl" alpine sh -c "mkdir -p /etc/nginx/ssl"
    
    # Copy certificates to container volume (using actual files, not symlinks)
    print_status "Copying certificates..."
    docker run --rm -v "$SSL_VOLUME_NAME:/etc/nginx/ssl" -v "$(dirname "$FULLCHAIN_FILE"):/certs:ro" alpine sh -c "cp /certs/$(basename "$FULLCHAIN_FILE") /etc/nginx/ssl/fullchain.pem && cp /certs/$(basename "$PRIVKEY_FILE") /etc/nginx/ssl/privkey.pem && ls -la /etc/nginx/ssl/"
    
    # Verify files were copied
    print_status "Verifying certificate copy..."
    docker run --rm -v "$SSL_VOLUME_NAME:/etc/nginx/ssl" alpine sh -c "ls -la /etc/nginx/ssl/"
    
    # Set proper permissions
    print_status "Setting permissions..."
    docker run --rm -v "$SSL_VOLUME_NAME:/etc/nginx/ssl" alpine sh -c "chmod 644 /etc/nginx/ssl/*.pem"
    
    # Start nginx with SSL
    print_status "Starting nginx with SSL..."
    docker-compose up -d nginx
    
    # Test SSL
    print_status "Testing SSL setup..."
    sleep 10
    if curl -k -s "https://$DOMAIN/health" > /dev/null; then
        print_success "SSL setup successful!"
    else
        print_warning "SSL setup may have issues. Check logs with: docker-compose logs nginx"
    fi
}

# Function to setup firewall
setup_firewall() {
    print_status "Setting up firewall..."
    
    if command -v ufw > /dev/null 2>&1; then
        # Ubuntu/Debian
        ufw allow ssh
        ufw allow 80
        ufw allow 443
        ufw --force enable
        print_success "UFW firewall configured"
    elif command -v firewall-cmd > /dev/null 2>&1; then
        # CentOS/RHEL
        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --permanent --add-service=http
        firewall-cmd --permanent --add-service=https
        firewall-cmd --reload
        print_success "firewalld configured"
    else
        print_warning "No supported firewall found. Please configure manually:"
        print_warning "  - Allow SSH (port 22)"
        print_warning "  - Allow HTTP (port 80)"
        print_warning "  - Allow HTTPS (port 443)"
    fi
}

# Main execution
main() {
    print_status "Detecting OS..."
    detect_os
    print_success "Detected OS: $OS $VER"
    
    print_status "Installing dependencies..."
    install_docker
    install_docker_compose
    install_certbot
    
    print_status "Setting up firewall..."
    setup_firewall
    
    print_status "Checking domain configuration..."
    check_domain
    
    print_status "Deploying containers..."
    deploy_containers
    
    print_status "Setting up SSL..."
    setup_ssl
    
    echo ""
    echo "====================================="
    echo "✅ Deployment Complete!"
    echo "====================================="
    echo ""
    echo "🌐 Your Stellar Sync server is now available at:"
    if [ "$SKIP_SSL" = "skip-ssl" ]; then
        echo "  - HTTP: http://$DOMAIN"
        echo "  - WebSocket: ws://$DOMAIN/ws"
    else
        echo "  - HTTPS: https://$DOMAIN"
        echo "  - WebSocket: wss://$DOMAIN/ws"
    fi
    echo ""
    echo "📋 Management commands:"
    echo "  docker-compose ps          # Check status"
    echo "  docker-compose logs        # View logs"
    echo "  docker-compose restart     # Restart services"
    echo "  docker-compose down        # Stop services"
    echo ""
    echo "🔒 SSL certificate will auto-renew"
    echo "📊 Check status: docker-compose ps"
    echo "📝 View logs: docker-compose logs"
    echo ""
    echo "🚀 Stellar Sync is ready to use!"
}

# Run main function
main "$@"

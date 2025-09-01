# 🚀 Stellar Sync Server

Simple server deployment for Stellar Sync with nginx, SSL, and Docker Compose.

## 📋 Quick Start

### 1. Start Services

```bash
# Start all services (main server, file server, nginx)
docker-compose up -d
```

### 2. Set up SSL (after DNS is configured)

```bash
# Make script executable
chmod +x setup-ssl.sh manage.sh

# Set up SSL certificate
./setup-ssl.sh your-domain.com your-email@example.com
```

### 3. Manage Services

```bash
# View status
./manage.sh status

# View logs
./manage.sh logs

# Restart services
./manage.sh restart
```

## 🔧 What's Included

### Services

- **Main Server** - WebSocket + HTTP proxy (port 6000)
- **File Server** - File uploads/downloads (port 6200)
- **Nginx** - Reverse proxy with SSL (ports 80, 443)

### Features

- ✅ **SSL/TLS** with automatic certificates
- ✅ **WebSocket support** for real-time communication
- ✅ **File uploads** with large file support (100MB)
- ✅ **Gzip compression** for performance
- ✅ **Security headers** for protection
- ✅ **Health checks** and monitoring
- ✅ **Auto-restart** and failover

## 📁 Files

```
stellarsync/
├── docker-compose.yml    # All services configuration
├── nginx.conf           # Nginx configuration
├── setup-ssl.sh         # SSL setup script
├── manage.sh            # Management script
├── ssl/                 # SSL certificates (created by setup)
└── logs/                # Nginx logs (created by setup)
```

## 🚀 Management Commands

```bash
# Start all services
./manage.sh start

# Stop all services
./manage.sh stop

# Restart all services
./manage.sh restart

# View logs
./manage.sh logs

# Check status
./manage.sh status

# Update services
./manage.sh update

# Create backup
./manage.sh backup

# Renew SSL certificate
./manage.sh ssl-renew

# Build from source
./manage.sh build

# Clean up everything
./manage.sh clean
```

## 🔒 SSL Certificate Setup

### Automatic Setup

```bash
# After DNS is configured to point to your server
./setup-ssl.sh your-domain.com your-email@example.com
```

### Manual Renewal

```bash
# Renew SSL certificate
./manage.sh ssl-renew
```

## 🌐 DNS Configuration

### For IPv4

```
Type: A
Name: stellar (or @)
Value: your-server-ip
TTL: 300
```

### For IPv6

```
Type: AAAA
Name: stellar (or @)
Value: your-server-ipv6
TTL: 300
```

## 📊 Monitoring

### Check Service Status

```bash
# View all services
docker-compose ps

# Check resource usage
docker stats

# View logs
docker-compose logs -f
```

### Health Checks

- **HTTP**: `http://your-server-ip/health`
- **HTTPS**: `https://your-domain.com/health`

## 🔧 Configuration

### Nginx Configuration

- **SSL/TLS** - Modern configuration with HSTS
- **WebSocket support** - For real-time communication
- **File uploads** - Large file support (100MB)
- **Gzip compression** - Better performance
- **Security headers** - XSS protection, etc.

### Environment Variables

You can customize the setup by editing `docker-compose.yml`:

```yaml
environment:
  - FILE_SERVER_URL=http://file-server:6001
```

## 🛠️ Troubleshooting

### Common Issues

1. **Services not starting**

   ```bash
   # Check logs
   ./manage.sh logs

   # Check Docker status
   docker ps -a
   ```

2. **SSL certificate issues**

   ```bash
   # Renew certificate
   ./manage.sh ssl-renew

   # Check certificate status
   sudo certbot certificates
   ```

3. **Port conflicts**

   ```bash
   # Check what's using ports
   sudo netstat -tlnp | grep :6000
   sudo netstat -tlnp | grep :6200
   ```

4. **Firewall issues**

   ```bash
   # Check firewall status
   sudo ufw status

   # Allow ports if needed
   sudo ufw allow 80
   sudo ufw allow 443
   ```

### Log Locations

- **Nginx logs**: `./logs/`
- **Docker logs**: `docker-compose logs`
- **System logs**: `sudo journalctl -u docker`

## 🔄 Updates

### Update Services

```bash
# Pull latest images and restart
./manage.sh update
```

### Update Configuration

```bash
# Backup current config
./manage.sh backup

# Edit configuration files
nano docker-compose.yml
nano nginx.conf

# Restart services
./manage.sh restart
```

## 💾 Backup & Restore

### Create Backup

```bash
# Backup configuration and data
./manage.sh backup
```

### Restore from Backup

```bash
# Extract backup
tar -xzf stellarsync-backup-YYYYMMDD-HHMMSS.tar.gz

# Restart services
./manage.sh restart
```

## 🔐 Security

### Firewall Configuration

- **SSH** (port 22) - Secure shell access
- **HTTP** (port 80) - Redirect to HTTPS
- **HTTPS** (port 443) - Secure web traffic
- **Application ports** (6000, 6200) - Internal services only

### SSL Configuration

- **TLS 1.2/1.3** - Modern protocols only
- **HSTS** - Force HTTPS
- **Security headers** - XSS protection, etc.
- **Auto-renewal** - Certificates renew automatically

## 📞 Support

### Getting Help

1. **Check logs**: `./manage.sh logs`
2. **Check status**: `./manage.sh status`
3. **Restart services**: `./manage.sh restart`
4. **Create backup**: `./manage.sh backup`

### Useful Commands

```bash
# View all containers
docker ps -a

# Check nginx configuration
docker exec stellarsync-nginx nginx -t

# Check SSL certificate
sudo certbot certificates

# Monitor real-time logs
docker-compose logs -f --tail=100
```

## 🎉 Success!

Once everything is set up, your Stellar Sync server will be available at:

- **WebSocket**: `wss://your-domain.com/ws`
- **File uploads**: `https://your-domain.com/upload`
- **Health check**: `https://your-domain.com/health`

---

**🚀 Your Stellar Sync server is now ready to serve users!**

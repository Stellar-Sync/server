#!/bin/bash

# Stellar Sync Kubernetes Environment Setup
echo "🚀 Stellar Sync Environment Setup"
echo "=================================="

# Get user input
read -p "Enter your domain (e.g., stellar.kasu.network): " DOMAIN
read -p "Enter your email address: " EMAIL
read -p "Enter your Docker Hub username: " DOCKER_USERNAME

# Update ingress.yaml with domain
sed -i "s/stellar.kasu.network/${DOMAIN}/g" k8s/ingress.yaml

# Update cluster-issuer.yaml with email
sed -i "s/your-email@example.com/${EMAIL}/g" k8s/cluster-issuer.yaml

# Update config.yaml
cat > k8s/config.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: stellarsync-config
  namespace: stellarsync
data:
  DOMAIN: "${DOMAIN}"
  EMAIL: "${EMAIL}"
  DOCKER_IMAGE_MAIN: "${DOCKER_USERNAME}/stellarsync-main:latest"
  DOCKER_IMAGE_FILESERVER: "${DOCKER_USERNAME}/stellarsync-fileserver:latest"
EOF

echo "✅ Environment variables updated in k8s/config.yaml"
echo ""
echo "📋 Configuration:"
echo "  Domain: ${DOMAIN}"
echo "  Email: ${EMAIL}"
echo "  Docker Images: ${DOCKER_USERNAME}/stellarsync-*:latest"
echo ""
echo "🎯 Next steps:"
echo "  1. Point your domain ${DOMAIN} to Oracle Cloud"
echo "  2. Build and push Docker images"
echo "  3. Deploy with: kubectl apply -k k8s/"

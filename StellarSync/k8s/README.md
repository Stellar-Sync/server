# Stellar Sync - Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the Stellar Sync server as a Docker fleet.

## 📁 Files Overview

- `namespace.yaml` - Creates the `stellarsync` namespace
- `storage.yaml` - Persistent volume claim for file storage
- `main-server-deployment.yaml` - WebSocket server deployment
- `file-server-deployment.yaml` - File server deployment
- `services.yaml` - Internal networking services
- `ingress.yaml` - External access and load balancing
- `cluster-issuer.yaml` - SSL certificate configuration
- `hpa.yaml` - Horizontal pod autoscalers
- `kustomization.yaml` - Kustomize configuration for easy deployment

## 🚀 Quick Deployment

### Prerequisites

1. **Set up environment variables**:

   ```bash
   # Run the setup script
   chmod +x k8s/setup-env.sh
   ./k8s/setup-env.sh
   ```

   This will prompt you for:

   - Your domain (e.g., stellar.kasu.network)
   - Your email address
   - Your Docker Hub username

2. **Build and push Docker images**:
   ```bash
   docker build -f Dockerfile.main -t your-dockerhub-username/stellarsync-main:latest .
   docker build -f Dockerfile.fileserver -t your-dockerhub-username/stellarsync-fileserver:latest .
   docker push your-dockerhub-username/stellarsync-main:latest
   docker push your-dockerhub-username/stellarsync-fileserver:latest
   ```

### Deploy with Kustomize (Recommended)

```bash
# Deploy everything at once
kubectl apply -k .

# Verify deployment
kubectl get all -n stellarsync
```

### Deploy Manually

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Deploy storage
kubectl apply -f storage.yaml

# Deploy deployments
kubectl apply -f main-server-deployment.yaml
kubectl apply -f file-server-deployment.yaml

# Deploy services
kubectl apply -f services.yaml

# Deploy ingress (after setting up domain)
kubectl apply -f ingress.yaml

# Deploy autoscalers
kubectl apply -f hpa.yaml
```

## 🔧 Configuration

### Environment Variables

| Service     | Variable          | Description      | Default                 |
| ----------- | ----------------- | ---------------- | ----------------------- |
| Main Server | `PORT`            | Main server port | `6000`                  |
| Main Server | `FILE_SERVER_URL` | File server URL  | `http://file-server:80` |
| File Server | `PORT`            | File server port | `6200`                  |

### Resource Limits

| Service     | CPU Request | CPU Limit | Memory Request | Memory Limit |
| ----------- | ----------- | --------- | -------------- | ------------ |
| Main Server | 250m        | 500m      | 256Mi          | 512Mi        |
| File Server | 250m        | 500m      | 512Mi          | 1Gi          |

### Scaling

| Service     | Min Replicas | Max Replicas | Scale Target |
| ----------- | ------------ | ------------ | ------------ |
| Main Server | 1            | 5            | 70% CPU      |
| File Server | 1            | 3            | 70% CPU      |

## 🔍 Monitoring

### Check Status

```bash
# Check all resources
kubectl get all -n stellarsync

# Check pods
kubectl get pods -n stellarsync

# Check services
kubectl get services -n stellarsync

# Check ingress
kubectl get ingress -n stellarsync
```

### View Logs

```bash
# Main server logs
kubectl logs -f deployment/main-server -n stellarsync

# File server logs
kubectl logs -f deployment/file-server -n stellarsync
```

### Scale Services

```bash
# Scale main server
kubectl scale deployment main-server --replicas=3 -n stellarsync

# Scale file server
kubectl scale deployment file-server --replicas=2 -n stellarsync
```

## 🔒 SSL Configuration

### Install cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

### Deploy ClusterIssuer

```bash
kubectl apply -f cluster-issuer.yaml
```

## 🚨 Troubleshooting

### Common Issues

1. **Pods not starting**:

   ```bash
   kubectl describe pod <pod-name> -n stellarsync
   kubectl logs <pod-name> -n stellarsync
   ```

2. **Services not connecting**:

   ```bash
   kubectl get endpoints -n stellarsync
   kubectl describe service main-server -n stellarsync
   ```

3. **Ingress not working**:
   ```bash
   kubectl get ingress -n stellarsync
   kubectl describe ingress stellarsync-ingress -n stellarsync
   ```

### Debug Commands

```bash
# Check events
kubectl get events -n stellarsync

# Port forward for testing
kubectl port-forward service/main-server 6000:80 -n stellarsync

# Check autoscaler status
kubectl get hpa -n stellarsync
```

## 📊 Platform-Specific Notes

### DigitalOcean Kubernetes

- Storage class: `do-block-storage`
- Ingress class: `nginx`
- Load balancer: Automatic

### Linode Kubernetes Engine

- Storage class: `linode-block-storage`
- Ingress class: `nginx`
- Load balancer: Automatic

### Oracle Cloud Kubernetes

- Storage class: `oci`
- Ingress class: `nginx`
- Load balancer: Automatic

## 🎯 Next Steps

1. **Update image names** in all YAML files
2. **Set your domain** in `ingress.yaml`
3. **Set your email** in `cluster-issuer.yaml`
4. **Deploy with kustomize**: `kubectl apply -k .`
5. **Verify deployment**: `kubectl get all -n stellarsync`

---

**🎯 Ready to deploy your Kubernetes fleet!** 🚀

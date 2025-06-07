# Valkyrie Kubernetes Probes Tutorial

## Introduction

This is a project related to a tutorial on how to enable Kubernetes probes. The Valkyrie application demonstrates startup, liveness, and readiness probes in a controlled environment, allowing you to understand how Kubernetes health checks work in practice.

## Prerequisites

- **k3d** (recommended for local development) or any Kubernetes cluster
- **kubectl** configured to access your cluster
- **Docker** and Docker CLI
- **Go 1.23+** (for local development)

## Link to the tutorial

[mastering-kubernetes-health-checks-probes-for-application-resilience - Part 1](https://devoriales.com/post/136/mastering-kubernetes-health-checks-probes-for-application-resilience-part-1-out-of-3)

[mastering-kubernetes-health-checks-probes-for-application-resilience - Part 2](https://devoriales.com/post/335/mastering-kubernetes-health-checks-deploy-application-with-probe-endpoints-part-2-out-of-3)

[mastering-kubernetes-health-checks-probes-for-application-resilience - Part 3](https://devoriales.com/post/336/mastering-kubernetes-health-checks-probe-configurations-with-valkyrie-part-3-out-of-3)

## Quick Start with k3d (Recommended)

### 1. Set up k3d cluster with registry

```bash
# Create cluster with built-in registry
k3d cluster create devoriales-cluster --registry-create registry.localhost:5000

# Verify cluster and registry
k3d cluster list
k3d registry list
kubectl cluster-info
```

### 2. Build and push the application

```bash
# Initialize Go module (if not already done)
go mod init k8s-probes
go mod tidy

# Build Docker image
docker build -t valkyrie-app:1.0 .

# Tag and push to k3d registry
docker tag valkyrie-app:1.0 registry.localhost:5000/valkyrie-app:1.0
docker push registry.localhost:5000/valkyrie-app:1.0

# Verify image in registry
curl -s http://localhost:5000/v2/_catalog
```

### 3. Deploy the application

```bash
# Apply Kubernetes manifests
kubectl apply -f manifests.yaml

# Watch pod startup and probe behavior
kubectl get pods -n valkyrie -w

# Check probe status
kubectl describe pod -n valkyrie -l app=critical-app
```

### 4. Access the application

```bash
# Port forward to access locally
kubectl port-forward -n valkyrie svc/critical-app-clusterip 8080:80

# Visit http://localhost:8080 in your browser
```

## Local Development

### Initialize Go module

```bash
go mod init k8s-probes
go mod tidy
```

### Run locally for testing

```bash
# Set environment variable
export PRIME_NUMBER_COUNT=100

# Run the application
go run main.go

# Test endpoints
curl http://localhost:8080/liveness-health
curl http://localhost:8080/readiness-health
curl http://localhost:8080/timestamps
```

### Build executable

```bash
go build -o main main.go
./main
```

## Docker Build Options

### Standard build

```bash
docker build -t valkyrie-app:1.0 .
```

### Multi-platform build (for Apple Silicon Macs)

```bash
# Create and use buildx instance
docker buildx create --use --name multiplatform-builder
docker buildx inspect --bootstrap

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-registry/valkyrie-app:1.0 \
  --push .

# Build for Linux amd64 only (most common for k8s)
docker buildx build \
  --platform linux/amd64 \
  -t valkyrie-app:1.0 \
  --load .
```

## Multi-Node k3d Cluster Setup

### Create cluster with multiple nodes

```bash
# Create cluster with 1 server and 2 agents
k3d cluster create devoriales-cluster \
  --servers 1 \
  --agents 1 \
  --registry-create registry.localhost:5000 \
  -p "8080:80@loadbalancer" \
  -p "8443:443@loadbalancer"


# Verify nodes
kubectl get nodes -o wide
```

### Scale deployment across nodes

```bash
# Update replicas in manifests.yaml
spec:
  replicas: 3  # This will distribute pods across nodes

# Apply changes
kubectl apply -f manifests.yaml

# Check pod distribution
kubectl get pods -n valkyrie -o wide
```

## Production Cluster Setup

### For cloud providers (GKE, EKS, AKS)

```bash
# Build and push to public registry
docker tag valkyrie-app:1.0 your-dockerhub-username/valkyrie-app:1.0
docker push your-dockerhub-username/valkyrie-app:1.0

# Update manifests.yaml image reference
image: your-dockerhub-username/valkyrie-app:1.0

# Deploy to production cluster
kubectl apply -f manifests.yaml
```

### For private registries

```bash
# Create registry secret
kubectl create secret docker-registry regcred \
  --docker-server=your-registry.com \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email@example.com

# Add imagePullSecrets to deployment
spec:
  template:
    spec:
      imagePullSecrets:
      - name: regcred
      containers:
      - name: critical-app
        image: your-registry.com/valkyrie-app:1.0
```

## Understanding the Probes

### Startup Probe

- **Purpose**: Prevents premature liveness/readiness checks during slow startup
- **Mechanism**: Checks for `/tmp/startup-file` existence
- **Timing**: Up to 80 seconds (10s initial + 7×10s checks)

### Liveness Probe

- **Purpose**: Detects and restarts unresponsive containers
- **Mechanism**: HTTP GET to `/liveness-health`
- **Timing**: Every 5 seconds after startup complete

### Readiness Probe

- **Purpose**: Controls traffic routing to healthy pods
- **Mechanism**: HTTP GET to `/readiness-health`
- **Timing**: Every 10 seconds after startup complete

## Application Features

### Web Interface

- Real-time probe status indicators
- Failure simulation toggles
- Timestamp tracking for probe executions

### Environment Variables

- `PRIME_NUMBER_COUNT`: Controls startup duration (default: 1000)

### Endpoints

- `/` - Main dashboard with controls
- `/liveness-health` - Liveness probe endpoint
- `/readiness-health` - Readiness probe endpoint
- `/timestamps` - Probe execution timestamps
- `/toggle-liveness-failure` - Simulate liveness failures
- `/toggle-readiness-failure` - Simulate readiness failures

## Troubleshooting

### Common Issues

**ImagePullBackOff Error**

```bash
# Check registry connectivity
curl -s http://localhost:5000/v2/_catalog

# Verify image exists
curl -s http://localhost:5000/v2/valkyrie-app/tags/list

# Test from inside cluster
kubectl run test --image=curlimages/curl --rm -it --restart=Never -- curl -s http://registry.localhost:5000/v2/_catalog
```

**Probe Failures**

```bash
# Check probe status
kubectl describe pod -n valkyrie -l app=critical-app

# View application logs
kubectl logs -n valkyrie -l app=critical-app

# Check probe endpoints manually
kubectl port-forward -n valkyrie svc/critical-app-clusterip 8080:80
curl http://localhost:8080/liveness-health
```

**Startup Taking Too Long**

```bash
# Reduce prime number count
# Update manifests.yaml:
env:
- name: PRIME_NUMBER_COUNT
  value: "100"  # Reduced from 1000
```

### Cleanup

```bash
# Delete application
kubectl delete namespace valkyrie

# Delete k3d cluster
k3d cluster delete devoriales-cluster

# Remove Docker images
docker rmi valkyrie-app:1.0 registry.localhost:5000/valkyrie-app:1.0
```

## Learning Objectives

By completing this tutorial, you will understand:

- How Kubernetes probes work in practice
- The relationship between startup, liveness, and readiness probes
- Proper probe timing and configuration
- How to simulate and troubleshoot probe failures
- Best practices for containerized application health checks

## Next Steps

1. Experiment with different probe configurations
2. Try failure scenarios using the web interface
3. Monitor probe behavior with `kubectl describe pod`
4. Scale the deployment and observe probe behavior across multiple pods
5. Implement similar health checks in your own applications

# Kubernetes Deployment Guide

## Prerequisites

- Kubernetes cluster (1.28+)
- kubectl configured
- Helm 3.x
- Container registry access

## Quick Start

### 1. Create Namespace

```bash
kubectl create namespace anonymous-support
```

### 2. Create Secrets

```bash
kubectl create secret generic app-secrets \
  --from-literal=jwt-secret=your-secret-key-32-chars \
  --from-literal=encryption-key=your-encryption-key-32-chars \
  --from-literal=postgres-password=your-db-password \
  -n anonymous-support
```

### 3. Deploy Dependencies

```bash
# PostgreSQL
helm install postgres bitnami/postgresql \
  --set auth.password=$DB_PASSWORD \
  --namespace anonymous-support

# MongoDB
helm install mongodb bitnami/mongodb \
  --namespace anonymous-support

# Redis
helm install redis bitnami/redis \
  --namespace anonymous-support
```

### 4. Deploy Application

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

### 5. Verify Deployment

```bash
kubectl get pods -n anonymous-support
kubectl logs -f deployment/anonymous-support-api -n anonymous-support
```

## Configuration

### Environment Variables

See `k8s/configmap.yaml` for all configuration options.

### Scaling

```bash
kubectl scale deployment anonymous-support-api \
  --replicas=5 -n anonymous-support
```

### Monitoring

Metrics are exposed at `/metrics` endpoint (Prometheus format).

```bash
kubectl port-forward svc/anonymous-support-api 8080:8080 -n anonymous-support
curl http://localhost:8080/metrics
```

## Troubleshooting

### Check logs

```bash
kubectl logs -f deployment/anonymous-support-api -n anonymous-support
```

### Check events

```bash
kubectl get events -n anonymous-support --sort-by='.lastTimestamp'
```

### Database connection issues

```bash
kubectl exec -it deployment/anonymous-support-api -n anonymous-support -- sh
# Test connections from within pod
```

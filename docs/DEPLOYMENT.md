# Deployment Guide

## Database Migrations as Deploy Step

### Migration Locking Strategy

To prevent concurrent migrations from multiple deployment instances, use migration locking:

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Acquire migration lock
        run: |
          # Use Redis for distributed locking
          redis-cli -h $REDIS_HOST SET migration:lock:$(git rev-parse HEAD) "locked" NX EX 300

      - name: Run PostgreSQL migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate -path migrations/postgres -database "$DATABASE_URL" up

      - name: Run MongoDB migrations
        run: |
          go run cmd/migrate/main.go --up

      - name: Release migration lock
        if: always()
        run: |
          redis-cli -h $REDIS_HOST DEL migration:lock:$(git rev-parse HEAD)

  deploy:
    needs: migrate
    runs-on: ubuntu-latest
    steps:
      - name: Deploy application
        run: kubectl apply -f k8s/
```

### Alternative: Advisory Locks (PostgreSQL)

```sql
-- Acquire advisory lock before migration
SELECT pg_advisory_lock(123456);

-- Run migrations
-- ...

-- Release lock
SELECT pg_advisory_unlock(123456);
```

## Blue/Green Deployment Strategy

### Kubernetes Blue/Green Deployment

#### Step 1: Create Blue and Green Deployments

```yaml
# k8s/deployment-blue.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anonymous-support-api-blue
  labels:
    app: anonymous-support-api
    version: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: anonymous-support-api
      version: blue
  template:
    metadata:
      labels:
        app: anonymous-support-api
        version: blue
    spec:
      containers:
      - name: api
        image: anonymous-support-api:v1.0.0
        # ... rest of container spec
```

```yaml
# k8s/deployment-green.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anonymous-support-api-green
  labels:
    app: anonymous-support-api
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: anonymous-support-api
      version: green
  template:
    metadata:
      labels:
        app: anonymous-support-api
        version: green
    spec:
      containers:
      - name: api
        image: anonymous-support-api:v1.1.0
        # ... rest of container spec
```

#### Step 2: Service Selector Switch

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: anonymous-support-api
spec:
  selector:
    app: anonymous-support-api
    version: blue  # Switch to 'green' to route traffic
  ports:
  - port: 80
    targetPort: 8080
```

#### Step 3: Deployment Script

```bash
#!/bin/bash
# deploy-blue-green.sh

CURRENT_VERSION=$(kubectl get service anonymous-support-api -o jsonpath='{.spec.selector.version}')
NEW_VERSION=${1:-green}

if [ "$CURRENT_VERSION" = "blue" ]; then
  NEW_VERSION="green"
  OLD_VERSION="blue"
else
  NEW_VERSION="blue"
  OLD_VERSION="green"
fi

echo "Current version: $CURRENT_VERSION"
echo "Deploying to: $NEW_VERSION"

# Deploy new version
kubectl apply -f k8s/deployment-$NEW_VERSION.yaml

# Wait for new deployment to be ready
kubectl rollout status deployment/anonymous-support-api-$NEW_VERSION

# Run smoke tests
./scripts/smoke-test.sh http://anonymous-support-api-$NEW_VERSION:8080

if [ $? -eq 0 ]; then
  echo "Smoke tests passed. Switching traffic to $NEW_VERSION"

  # Switch service to new version
  kubectl patch service anonymous-support-api -p '{"spec":{"selector":{"version":"'$NEW_VERSION'"}}}'

  echo "Traffic switched to $NEW_VERSION"
  echo "Old version ($OLD_VERSION) is still running for rollback if needed"
else
  echo "Smoke tests failed. Keeping traffic on $CURRENT_VERSION"
  exit 1
fi
```

### Rollback Procedure

```bash
#!/bin/bash
# rollback.sh

CURRENT_VERSION=$(kubectl get service anonymous-support-api -o jsonpath='{.spec.selector.version}')

if [ "$CURRENT_VERSION" = "blue" ]; then
  ROLLBACK_VERSION="green"
else
  ROLLBACK_VERSION="blue"
fi

echo "Rolling back from $CURRENT_VERSION to $ROLLBACK_VERSION"

kubectl patch service anonymous-support-api -p '{"spec":{"selector":{"version":"'$ROLLBACK_VERSION'"}}}'

echo "Rollback complete"
```

## CDN and Edge Protection

### CloudFlare Configuration

```yaml
# cloudflare-config.yaml
dns:
  - type: A
    name: api.anonymous-support.com
    content: <LOAD_BALANCER_IP>
    proxied: true  # Enable CloudFlare proxy

security:
  ssl:
    mode: full_strict
    min_tls_version: "1.2"

  waf:
    enabled: true
    rules:
      - description: "Rate limit API endpoints"
        expression: '(http.request.uri.path contains "/api/")'
        action: challenge
        rate_limit:
          requests: 100
          period: 60

  ddos_protection:
    enabled: true
    sensitivity: medium

  firewall_rules:
    - description: "Block known bad IPs"
      expression: '(ip.geoip.country in {"XX" "YY"})'
      action: block

    - description: "Require authentication for sensitive endpoints"
      expression: '(http.request.uri.path contains "/admin/")'
      action: js_challenge

caching:
  level: aggressive
  browser_ttl: 14400  # 4 hours
  edge_ttl: 7200      # 2 hours

  page_rules:
    - url: "api.anonymous-support.com/api/*"
      settings:
        cache_level: bypass  # Don't cache API responses

    - url: "api.anonymous-support.com/health"
      settings:
        cache_level: bypass
```

### AWS CloudFront Configuration

```json
{
  "DistributionConfig": {
    "Origins": {
      "Items": [
        {
          "Id": "api-origin",
          "DomainName": "api.anonymous-support.com",
          "CustomOriginConfig": {
            "HTTPPort": 80,
            "HTTPSPort": 443,
            "OriginProtocolPolicy": "https-only",
            "OriginSslProtocols": {
              "Items": ["TLSv1.2"]
            }
          }
        }
      ]
    },
    "DefaultCacheBehavior": {
      "TargetOriginId": "api-origin",
      "ViewerProtocolPolicy": "redirect-to-https",
      "AllowedMethods": {
        "Items": ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
      },
      "CachedMethods": {
        "Items": ["GET", "HEAD", "OPTIONS"]
      },
      "Compress": true,
      "ForwardedValues": {
        "QueryString": true,
        "Headers": {
          "Items": ["Authorization", "Content-Type"]
        }
      }
    },
    "PriceClass": "PriceClass_100",
    "ViewerCertificate": {
      "ACMCertificateArn": "arn:aws:acm:...",
      "SSLSupportMethod": "sni-only",
      "MinimumProtocolVersion": "TLSv1.2_2021"
    },
    "WebACLId": "arn:aws:wafv2:..."
  }
}
```

## Staging Environment Setup

### Namespace-based Staging

```yaml
# k8s/staging/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: anonymous-support-staging
  labels:
    environment: staging
```

```yaml
# k8s/staging/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anonymous-support-api
  namespace: anonymous-support-staging
spec:
  replicas: 2  # Fewer replicas than production
  # ... rest same as production
```

### Environment-specific ConfigMaps

```yaml
# k8s/staging/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: anonymous-support-staging
data:
  SERVER_ENV: "staging"
  LOG_LEVEL: "debug"
  ENABLE_AUTO_MODERATION: "true"
```

### Smoke Tests

```bash
#!/bin/bash
# scripts/smoke-test.sh

API_URL=${1:-http://localhost:8080}

echo "Running smoke tests against $API_URL"

# Test 1: Health check
echo "Test 1: Health check"
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" $API_URL/health)
if [ "$HEALTH_STATUS" != "200" ]; then
  echo "FAILED: Health check returned $HEALTH_STATUS"
  exit 1
fi
echo "PASSED"

# Test 2: Anonymous registration
echo "Test 2: Anonymous registration"
REGISTER_RESPONSE=$(curl -s -X POST $API_URL/api.auth.v1.AuthService/RegisterAnonymous \
  -H "Content-Type: application/json" \
  -d '{"username": "smoke_test_user"}')

ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken')
if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "FAILED: Registration did not return access token"
  exit 1
fi
echo "PASSED"

# Test 3: Create post
echo "Test 3: Create post"
POST_RESPONSE=$(curl -s -X POST $API_URL/api.post.v1.PostService/CreatePost \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type": "CheckIn", "content": "Smoke test post", "categories": ["test"], "urgencyLevel": 1}')

POST_ID=$(echo $POST_RESPONSE | jq -r '.id')
if [ "$POST_ID" = "null" ] || [ -z "$POST_ID" ]; then
  echo "FAILED: Post creation did not return post ID"
  exit 1
fi
echo "PASSED"

# Test 4: Get feed
echo "Test 4: Get feed"
FEED_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  "$API_URL/api.post.v1.PostService/GetFeed?limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if [ "$FEED_STATUS" != "200" ]; then
  echo "FAILED: Feed returned $FEED_STATUS"
  exit 1
fi
echo "PASSED"

echo "All smoke tests passed!"
exit 0
```

### Staging Deployment Workflow

```yaml
# .github/workflows/deploy-staging.yml
name: Deploy to Staging

on:
  push:
    branches: [develop]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build and push image
        run: |
          docker build -t anonymous-support-api:staging-${{ github.sha }} .
          docker push anonymous-support-api:staging-${{ github.sha }}

      - name: Deploy to staging
        run: |
          kubectl set image deployment/anonymous-support-api \
            api=anonymous-support-api:staging-${{ github.sha }} \
            -n anonymous-support-staging

      - name: Wait for rollout
        run: |
          kubectl rollout status deployment/anonymous-support-api \
            -n anonymous-support-staging

      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh https://staging-api.anonymous-support.com

      - name: Notify on failure
        if: failure()
        run: |
          # Send notification (Slack, email, etc.)
          echo "Staging deployment failed!"
```

## Production Deployment Checklist

- [ ] Database migrations completed successfully
- [ ] Smoke tests passed in staging
- [ ] Blue/green deployment ready
- [ ] CDN cache cleared if needed
- [ ] Monitoring alerts configured
- [ ] Rollback procedure tested
- [ ] On-call engineer notified
- [ ] Feature flags configured appropriately

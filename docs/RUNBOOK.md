# Incident Response Runbook

## Critical Alerts

### Database Connection Failure

**Symptoms:** 
- 500 errors on all endpoints
- Logs show "failed to connect to database"

**Resolution:**
1. Check database pod status: `kubectl get pods -n anonymous-support`
2. Check database logs: `kubectl logs postgres-0 -n anonymous-support`
3. Verify secrets: `kubectl get secret app-secrets -n anonymous-support -o yaml`
4. Restart database if needed: `kubectl rollout restart statefulset/postgres`
5. Restart app pods: `kubectl rollout restart deployment/anonymous-support-api`

### High Memory Usage

**Symptoms:**
- OOMKilled pod restarts
- Slow response times

**Resolution:**
1. Check current memory: `kubectl top pods -n anonymous-support`
2. Review memory limits in deployment
3. Check for memory leaks in logs
4. Increase memory limits if legitimate usage:
   ```bash
   kubectl set resources deployment anonymous-support-api \
     --limits=memory=2Gi -n anonymous-support
   ```

### Rate Limit Issues

**Symptoms:**
- Users reporting 429 errors
- Redis connection issues

**Resolution:**
1. Check Redis status: `kubectl get pods -l app=redis`
2. Verify Redis connection from app pod
3. Review rate limit configuration in ConfigMap
4. Temporarily increase limits if needed
5. Check for abuse patterns in logs

### WebSocket Connection Drops

**Symptoms:**
- Real-time features not working
- Clients frequently disconnecting

**Resolution:**
1. Check load balancer timeout settings
2. Verify WebSocket upgrade headers in ingress
3. Check app logs for authentication failures
4. Review network policies
5. Test WebSocket endpoint: `wscat -c wss://api.example.com/ws`

## Monitoring Dashboards

- **Grafana**: https://grafana.example.com/d/app-overview
- **Prometheus**: https://prometheus.example.com
- **Logs**: https://kibana.example.com

## Escalation

1. **L1**: On-call engineer (Slack: #incidents)
2. **L2**: Platform team lead
3. **L3**: CTO

## Emergency Contacts

- On-call rotation: PagerDuty
- Platform team: platform@example.com
- Security team: security@example.com

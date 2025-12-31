# Production Readiness TODO

This list focuses on hardening, observability, security, reliability, and developer workflow. Items are grouped by theme so you can roadmap them in phases.

## Architecture and Dependency Management
- Introduce dependency injection wiring (manual or container). Keep constructors small, pass interfaces, and centralize composition in a dedicated package.
- Convert repositories and services to interfaces for testability; keep concrete implementations in storage-specific packages.
- Add a clean separation between transport layer (HTTP/Connect-RPC/WebSocket) and service layer with request/response DTOs.
- Implement a proper application lifecycle manager (start/stop hooks, readiness gating, graceful shutdown for all clients).
- Add a background job runner abstraction (e.g., for scheduled cleanup, moderation checks, email/notifications).

## API Layer and Transport
- Implement Connect-RPC handlers for all proto services and wire them into the HTTP server.
- Add versioned REST endpoints or a gateway if you plan to expose non-RPC clients.
- Add request validation at the transport boundary with clear error mapping (client-safe messages, internal details in logs).
- Add pagination and filtering conventions consistently across APIs.
- Document API with OpenAPI or Connect-RPC docs and publish a developer portal.
- Add rate limiting at the edge and per-endpoint using configurable policies.

## Authentication and Authorization
- Harden JWT auth: key rotation, token issuer/audience, stronger claims validation, short access token TTL with refresh rotation.
- Store refresh tokens with rotation and revocation lists; detect reuse and lock sessions.
- Add OAuth2 integration (Google/Apple/etc.) with secure redirect handling and PKCE if needed.
- Add RBAC/ABAC for moderator/admin actions (moderation queue, user bans, circle management).
- Add per-resource authorization checks (posts, circles, reports) with consistent ownership rules.
- Add audit logging for auth events (login, logout, refresh, failed attempts).

## Security and Compliance
- Add config validation and required secrets check at startup.
- Use a dedicated secret manager in production (KMS, Vault, or cloud provider secrets).
- Ensure all sensitive data is encrypted at rest and in transit (email, tokens, PII).
- Add request size limits, strict CORS configuration, and secure headers (HSTS, CSP, etc.).
- Add input sanitation for content moderation and prevent injection vectors.
- Add dependency scanning and SCA in CI.
- Add threat modeling and abuse prevention plans (spam, harassment, malicious uploads).

## Data Layer and Persistence
- Add connection pooling and timeouts for all DBs; surface metrics.
- Add migrations for MongoDB beyond indexes (schema validation versioning and migration runner).
- Add transactional consistency where needed (e.g., create post + publish events).
- Implement optimistic concurrency or version fields for critical updates.
- Add soft delete for posts and user accounts.
- Add data retention and archival policies.
- Add backups and restore validation runbooks.

## Real-time and Messaging
- Replace in-process WebSocket hub with a scalable pub/sub approach (Redis Streams, NATS, Kafka, or managed service).
- Add per-user subscriptions and auth checks for WS channels.
- Add backpressure and disconnect handling for slow clients.
- Add message schema versioning for WS payloads.
- Add mobile push notification integration (APNs/FCM) instead of placeholder logger.

## Observability
- Add structured logging with request IDs and correlation IDs.
- Add metrics (Prometheus/OpenTelemetry) for latency, error rates, DB ops, Redis ops, WS connections.
- Add distributed tracing (OTel) and exporter configuration.
- Add health checks (liveness, readiness, dependency checks) and align with orchestrator probes.
- Add error reporting (Sentry or equivalent) for uncaught panics and critical errors.

## Reliability and Performance
- Add retry/backoff for transient DB/Redis errors with circuit breakers.
- Add caching strategies for hot endpoints (feed, user profiles, circle lists).
- Add load tests and performance benchmarks for read/write paths.
- Add timeout and context deadlines consistently across all repository calls.
- Add graceful shutdown for WS clients and background goroutines.

## Content Moderation and Safety
- Replace naive keyword checks with a moderation pipeline (third-party API + rules engine).
- Add escalation and review queue UI/API for moderators.
- Add user blocking and reporting flows to the API layer.
- Add rate-limited emergency/SOS alerts and abuse controls.

## Testing Strategy
- Add unit tests for services with mocked repositories.
- Add integration tests for Postgres/Mongo/Redis using test containers.
- Add API contract tests for Connect-RPC endpoints.
- Add load tests for WebSocket and feed operations.
- Add security tests for auth, rate limiting, and input validation.

## CI/CD and Release
- Add CI workflows (lint, tests, static analysis, SCA, container scan).
- Add code generation checks for proto outputs.
- Add build artifacts versioning and release notes automation.
- Add pre-commit hooks for gofmt, goimports, lint.
- Add environment-specific config files and templating (dev/staging/prod).

## Infrastructure and Deployment
- Add Kubernetes manifests or Helm chart with autoscaling policies.
- Add database migrations as a deploy step with locking.
- Add blue/green or rolling deployment strategy.
- Add CDN and edge protections if serving any public assets.
- Add separate staging environment and smoke tests.

## Developer Experience
- Consider replacing Makefile with Taskfile.yml (or keep Makefile and add Taskfile as an alternative).
- Add local dev scripts for seeding data and resetting environments.
- Add devcontainer or docker-compose override for local tooling.
- Add architecture docs and ADRs to document decisions.

## Documentation
- Add a full runbook for on-call: alerts, remediation steps, known failure modes.
- Add API usage examples and common workflows.
- Add data model diagrams and service interaction diagrams.

## Configuration and Secrets
- Move config to a typed schema with validation errors on startup.
- Add config for timeouts, rate limits, feature flags, and moderation settings.
- Add feature flagging system for rollout control.

## Miscellaneous
- Add linting with golangci-lint and enforce in CI.
- Add static analysis (go vet, govulncheck) to CI.
- Add proper licensing and security policy.

---

## Roadmap Phases (Prioritized)
### Phase 1: Production Foundations (Stability + API Enablement)
- Wire Connect-RPC handlers and transport validation for all proto services.
- Add strict config validation, env defaults, and required secrets checks.
- Add global timeouts, context deadlines, and graceful shutdown coverage.
- Add basic observability: structured logging with request IDs and health checks.
- Add CI pipeline for gofmt, go test, golangci-lint, go vet.
- Add storage timeouts/pooling settings and ensure they are configurable.

### Phase 2: Security + Auth Maturity
- Harden JWT validation (issuer/audience, key rotation, refresh rotation).
- Add OAuth2 integration (Google/Apple) and link/unlink flows.
- Implement RBAC/ABAC for moderation/admin endpoints.
- Add audit logging and security event tracking.
- Add secure headers, request size limits, and rate limits per endpoint.

### Phase 3: Reliability + Observability
- Add metrics (Prometheus/OTel) and tracing with exporter config.
- Add error reporting and panic capture.
- Add caching for hot paths and circuit breakers for Redis/DB.
- Add integration tests with test containers and contract tests for RPC.

### Phase 4: Scale + Ops
- Replace in-process WS hub with scalable pub/sub (Redis Streams/NATS/Kafka).
- Add Kubernetes manifests/Helm chart with autoscaling and probes.
- Add deployment strategy (rolling/blue-green) with migration locks.
- Add runbooks, SLOs, and alerting rules.

---

## Issue-Sized Backlog (Actionable Tasks)
Each item is intended to be small enough for a single PR.

### Core Wiring and API
- [ ] Implement Connect-RPC handlers for AuthService and UserService.
- [ ] Implement Connect-RPC handlers for PostService and SupportService.
- [ ] Implement Connect-RPC handlers for CircleService and ModerationService.
- [ ] Add request/response DTOs and validation at handler boundaries.
- [ ] Add error mapping policy (client-safe errors vs internal logs).

### Dependency Injection and Interfaces
- [ ] Add `internal/app` package to wire dependencies and lifecycle.
- [ ] Convert repositories to interfaces and move concrete types to storage packages.
- [ ] Convert services to interfaces and add mocks for tests.

### Auth and Security
- [ ] Add JWT claims validation (issuer, audience, not-before).
- [ ] Add refresh token rotation and revoke-on-reuse.
- [ ] Add OAuth2 provider integration (Google) with PKCE.
- [ ] Add role model and authorization checks for moderation actions.
- [ ] Add audit log table and service for security events.

### Config and Secrets
- [ ] Add typed config validation with explicit error messages.
- [ ] Add config options for timeouts, rate limits, and moderation flags.
- [ ] Add secret loading from env + optional secret manager integration.

### Observability
- [ ] Add request ID middleware and propagate in logs.
- [ ] Add Prometheus metrics endpoint and base instrumentation.
- [ ] Add OpenTelemetry tracing with configurable exporter.
- [ ] Add panic recovery middleware with error reporting.

### Data and Persistence
- [ ] Add MongoDB migration runner with version tracking.
- [ ] Add soft delete fields and update queries accordingly.
- [ ] Add index review and ensure TTL for expiring posts.
- [ ] Add transactional patterns for multi-step operations.

### Real-time and Notifications
- [ ] Add authz checks for WS channel subscriptions.
- [ ] Add message schema versioning for WS payloads.
- [ ] Add push notification provider integration (FCM/APNs).

### Reliability and Performance
- [ ] Add retry/backoff wrappers for Redis and Mongo operations.
- [ ] Add caching layer for feeds and user profiles with TTLs.
- [ ] Add load tests for feed and WS message fanout.

### Testing
- [ ] Add unit tests for AuthService and PostService with mocks.
- [ ] Add integration tests using test containers for Postgres/Mongo/Redis.
- [ ] Add RPC contract tests for all services.

### CI/CD and DX
- [ ] Add GitHub Actions pipeline for lint/test/scan.
- [ ] Add `Taskfile.yml` and align with existing Make targets.
- [ ] Add devcontainer or docker-compose override for local tooling.

### Documentation and Ops
- [ ] Add API examples for core workflows (register, post, respond).
- [ ] Add runbook with common incidents and recovery steps.
- [ ] Add architecture diagram and ADRs.

# TODO v2: Build-On and Improvements

This list reflects what is present in the repo today (RPC handlers, DTOs, interfaces, tracing, oauth helpers, migrations, tests, k8s manifests, Taskfile) and highlights what to build next or fix.

## 1) Architectural Cleanup and Consistency
- Consolidate the entrypoint: use `internal/app` for wiring or remove it; avoid duplicate wiring paths between `cmd/server/main.go` and `internal/app`.
- Align service constructors with interfaces/DTOs (some services accept raw params while interfaces accept DTOs).
- Standardize middleware usage (current code uses both `NewXxxMiddleware` style and old functions).
- Remove or migrate legacy/unused packages (e.g., old JWT manager name mismatch, duplicate k8s manifests).
- Introduce a single config validation layer with strict required fields and defaults.

## 2) RPC Handler Completeness
- Finish stubbed RPC endpoints (e.g., delete post, update urgency, support stats) and propagate actual service logic.
- Normalize error handling across RPC handlers with consistent error codes.
- Enforce auth and ownership checks in all RPC handlers (posts, circles, moderation).
- Add request validation at the boundary using DTO validation helpers.

## 3) Auth, Sessions, and Security
- Integrate refresh token rotation logic from `internal/repository/redis/token_rotation.go`.
- Wire OAuth2 Google provider + PKCE into auth flows and add account linking.
- Add RBAC/ABAC checks via `internal/pkg/authz` for moderator/admin endpoints.
- Add audit log writing (domain + repo exist) for sensitive actions.
- Add secure header middleware (HSTS, CSP, X-Content-Type-Options, etc.).

## 4) Observability and Tracing
- Wire `internal/pkg/tracing` + `internal/middleware/tracing.go` into server startup.
- Add metrics instrumentation for DB, cache, Redis, and WebSocket events (metrics already defined).
- Add request ID to logging output for correlation.
- Add error reporting integration (Sentry/OTel logs exporter).

## 5) Data Layer and Transactions
- Use `internal/pkg/transaction` for circle join/leave and other multi-step updates.
- Add MongoDB migration runner wiring (`internal/pkg/migrations/mongodb.go`).
- Add soft delete handling in queries (migration exists).
- Add indexes review + verify TTL usage for expiring posts.

## 6) Caching and Performance
- Integrate `internal/pkg/cache` for feeds and profiles.
- Add `internal/pkg/retry` to DB/Redis calls where safe.
- Add per-endpoint rate limit middleware wiring.
- Add streaming or pagination optimizations for feed endpoints.

## 7) Real-time and Notifications
- Wire WebSocket auth middleware (`internal/handler/websocket/auth.go`).
- Add channel-based subscriptions and per-user targeting.
- Integrate push notifications (`internal/pkg/notifications/push.go`).
- Add message schema versioning for WS payloads (`internal/handler/websocket/schema.go`).

## 8) Testing and Quality
- Expand unit tests for service logic (only a few exist).
- Add integration tests for Postgres/Mongo/Redis using test containers.
- Use the contract and load tests (`tests/contract`, `tests/load`) in CI.
- Add lint configuration and enforce via Taskfile + CI.

## 9) CI/CD and Release
- Add CI workflow (lint, test, govulncheck, buf generation).
- Add container build + scan in CI.
- Add release versioning (build metadata + tags).

## 10) Product/Feature Build-On Ideas
- Personalized feeds (by category, circle, and user engagement).
- Moderation review dashboard and queue APIs.
- Circle invites, private circle access control, and admin roles.
- User progress dashboards (streaks, milestones, relapse patterns).
- Content search and discovery endpoints.
- Abuse prevention: spam heuristics, user blocking UI, report workflow.

## 11) Docs and Ops
- Expand `docs/` with API examples and onboarding flows.
- Add runbook for incidents and DB recovery.
- Add deployment docs for k8s secrets/configmaps/ingress.

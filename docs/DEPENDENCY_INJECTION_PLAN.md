# Dependency Injection Implementation Plan

## Current State (DI Assessment)
- Composition root is split between `cmd/server/main.go` (infra clients) and `internal/app/app.go` (repositories/services/handlers).
- Services depend on concrete repository types (examples: `internal/service/post_service.go`, `internal/service/user_service.go`, `internal/service/support_service.go`).
- Handlers depend on concrete service types (examples: `internal/handler/rpc/auth_handler.go`, `internal/handler/rpc/post_handler.go`, `internal/handler/rpc/user_handler.go`).
- App wiring uses concrete type assertions (`internal/app/app.go` uses `a.UserRepo.(*postgres.UserRepository)` and similar).
- Interfaces exist but are not consistently used or aligned:
  - `internal/service/interfaces.go` does not match current service method signatures used by handlers.
  - `internal/repository/interfaces.go` does not include several methods used by services (e.g., realtime publish methods, support stats, analytics tracker helpers).
- Tests call out the DI pain explicitly (see notes in `internal/service/post_service_test.go` and `internal/service/user_service_test.go`).

Conclusion: dependency injection is partial and not “proper” in the sense of interface-driven boundaries and constructor-based composition. Concrete dependencies leak into services and handlers, and the wiring layer relies on type assertions.

## Goals
- Use constructor-based DI across handlers, services, repositories, and infrastructure wrappers.
- Depend on narrow interfaces at boundaries (handlers -> services, services -> repositories/infrastructure).
- Eliminate type assertions in wiring.
- Make unit tests easy with mocks/fakes for services and repositories.
- Keep a single composition root (either explicit manual wiring or a DI tool like Wire).

## Proposed Target Architecture
- **Handlers** depend on service interfaces only.
- **Services** depend on repository interfaces and small infra interfaces (cache, JWT, encryption, tracing, transactions).
- **Repositories** remain concrete types but explicitly implement interfaces declared in `internal/repository` (or split into smaller, domain-specific interfaces).
- **Composition root** is a dedicated wiring module that creates concrete implementations and returns interface-typed dependencies.

## Detailed Plan

### Phase 1: Interface Alignment and Boundary Cleanup
1. **Inventory and map dependencies**
   - Create a table mapping each service to the repo/infra methods it uses.
   - Create a table mapping each handler to the service methods it calls.
   - Use this inventory to define minimal interfaces.

2. **Align repository interfaces to actual usage**
   - Update `internal/repository/interfaces.go` to include methods used by services (e.g., `PublishNewPost`, `PublishNewResponse`, `AddSupporterToPost`, `GetSupporterCount`, `GetResponses`, `GetUserStats`, `GetTracker`, etc.).
   - Prefer small, focused interfaces (e.g., `PostRepository`, `PostRealtimePublisher`, `SupportStatsRepository`) rather than a single large interface.
   - Add compile-time assertions in each repo implementation:
     - Example: `var _ repository.PostRepository = (*mongodb.PostRepository)(nil)`

3. **Align service interfaces to handler usage**
   - Update `internal/service/interfaces.go` (or move to `internal/handler` if preferred) so interfaces match the actual service method signatures.
   - Decide on domain vs DTO boundaries:
     - Option A: services return domain models and handlers map to DTO/proto (current practice).
     - Option B: services return DTOs directly.
   - Make that choice consistent and update interfaces accordingly.

### Phase 2: Refactor Services to Depend on Interfaces
4. **Change service constructors to accept interfaces**
   - Update constructors to accept `repository.*` interfaces instead of concrete repo types.
   - Replace concrete fields like `*mongodb.PostRepository` with interface types.
   - Add small infra interfaces where needed:
     - Cache interface wrapping `internal/pkg/cache.Cache`.
     - Transaction interface wrapping `internal/pkg/transaction.Manager`.
     - JWT/encryption interfaces for easier mocking.

5. **Update service implementations to use interface methods**
   - Replace calls to concrete-only methods with interface methods (and add those methods to interfaces if required).
   - If a concrete method is too implementation-specific, move that logic into the repository layer instead.

### Phase 3: Refactor Handlers and App Wiring
6. **Update handlers to depend on service interfaces**
   - Change handler fields from `*service.FooService` to `service.FooServiceInterface`.
   - Update constructors accordingly.
   - This allows handler tests to use mocked services.

7. **Simplify application wiring**
   - Replace concrete-typed fields in `internal/app/app.go` with interface types.
   - Remove all type assertions.
   - Keep wiring as explicit constructor calls with interface values, or centralize it into a dedicated `internal/app/wire.go` (manual) or `internal/di` package.

### Phase 4: Optional DI Tooling (Recommended)
8. **Introduce Google Wire for compile-time DI**
   - Add `github.com/google/wire` to `go.mod`.
   - Create `internal/di/wire.go` with provider sets:
     - Config, logger, DB clients, repositories, services, handlers, app.
   - Generate `wire_gen.go` and update `cmd/server/main.go` to call `di.InitializeApp(...)`.
   - Benefit: compile-time checking of the dependency graph.

### Phase 5: Testing and Verification
9. **Add mocks/fakes for unit tests**
   - Use `mockgen` or `testify/mock` for service/repository interfaces.
   - Update service tests to use mocked dependencies (e.g., in `internal/service/post_service_test.go`).
   - Add handler tests with service mocks to validate RPC mapping and error handling.

10. **Verification checklist**
   - Build succeeds with no type assertions in wiring.
   - `go test ./...` passes.
   - Services and handlers can be unit-tested without real databases.

## Deliverables
- Updated interfaces in `internal/repository/interfaces.go` and `internal/service/interfaces.go`.
- Service constructors and implementations refactored to use interfaces.
- Handlers refactored to depend on service interfaces.
- Clean wiring layer without type assertions.
- Optional Wire setup for compile-time DI and a single composition root.

## Risks / Notes
- Some method signatures in services/repos drift from the interface definitions today; align these before refactoring to avoid mismatches.
- If domain-vs-DTO boundaries change, update handlers and tests together to avoid inconsistent models.
- Transaction use in `internal/service/circle_service.go` relies on `*sqlx.Tx`; consider a small transaction interface for testing.

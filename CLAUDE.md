## Project Overview

Personal financial app project that stores income or spending on this server, as well as provide aggregate and trends.

## Quick Commands

```bash
make dev          # Run with hot reload (air)
make test         # Run all tests
make lint         # Run linter
make build        # Build binary
docker-compose up # Start dependencies (DB, Redis, etc.)
```

## Project Structure

```
project-root/
├── cmd/
│   └── main.go              # Entrypoint only — wiring happens in config/
├── config/
│   ├── database.go          # DB connection + migrations
│   └── routes.go            # Route registration + dependency injection
├── internal/
│   ├── {domain}/            # One folder per domain (user, payment, order, etc.)
│   │   ├── model.go         # Entity + validation
│   │   ├── repository.go    # Data access interface + implementation
│   │   ├── service.go       # Business logic interface + implementation
│   │   └── handler.go       # HTTP handlers (defines its own Service interface)
│   ├── interfaces/          # Shared interfaces used across domains
│   ├── middleware/          # HTTP middleware
│   └── util/                # Generic helpers
├── migrations/              # SQL migrations (sequential numbering)
├── .gitignore/              # files to ignore, make sure we include binary build files and the like
├── .air.toml                # for air hotreloading when using make dev, make sure to configure to match where our main.go is, its NOT what is default init
└── Makefile
```

### Domain Package Pattern

Each domain is self-contained. Handler defines what it needs from Service. Service defines what it needs from Repository. This follows ISP — consumers own their interfaces.

```
internal/payment/
├── model.go         # Payment, Subscription entities
├── repository.go    # Repository interface + *repository implementation
├── service.go       # Service interface + *service implementation
├── handler.go       # Handler struct + Service interface it consumes
├── processor.go     # PaymentProcessor interface (for Stripe abstraction)
└── stripe.go        # Stripe implementation of PaymentProcessor
```

## Code Style (CRITICAL)

- Use `slog` for structured logging (NOT `log`)
- Error handling: wrap with `fmt.Errorf("context: %w", err)`
- Context propagation: always pass `ctx context.Context` as first param
- Naming: follow Go conventions (no get/set prefixes, no stuttering)

## Interface & Dependency Design

- **Define interfaces at point of use**, not implementation (consumer owns the interface)
- **Follow ISP**: interfaces expose only what the consumer needs
- **Follow DIP**: depend on interfaces, not concrete types
- **Follow IoC**: receiver controls dependencies (inject via constructor, never create internally)

Example — handler defines what it needs from service:

```go
// internal/payment/handler.go
type Service interface {
    CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (string, error)
    ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
    // Only methods handler actually uses — NOT the full service
}
```

Example — service defines what it needs from repository:

```go
// internal/payment/service.go
type Repository interface {
    Create(ctx context.Context, payment *Payment) error
    GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
    // Only methods service actually uses
}
```

Example — cross-domain dependency via narrow interface:

```go
// internal/payment/service.go
type PaymentUserService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
    UpdateStripeCustomer(ctx context.Context, userID uuid.UUID, customerID string) error
    // Only what payment service needs from user service
}
```

## Testing Requirements

- **TDD workflow**: Write failing tests first, then implement
- Test files live next to implementation: `service.go` → `service_test.go`
- Use table-driven tests for multiple cases
- Integration tests use `internal/testutil/suite.go` for setup

```bash
make test                    # All tests
make test-payment           # Single domain
go test ./internal/payment -run TestSpecificFunction -v
```

## Git Workflow

- Commit messages: `type: description` (feat, fix, test, refactor, chore, docs)
- Commit tests separately from implementation when doing TDD
- Never commit code that fails `make lint && make test`

## Architecture Rules

<!-- Add project-specific rules here -->

- All external services (Stripe, SendGrid, etc.) must be behind interfaces
- Database queries only in repository layer
- Business logic only in service layer
- Handlers do: parse request, call service, format response — nothing else
- No domain logic in handlers

## Environment Variables

```bash
# Required
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=myapp

# Optional
PORT=8000
JWT_SECRET=xxx
```

## Common Patterns

### Creating a New Domain

1. Create folder: `internal/{domain}/`
2. Create files: `model.go`, `repository.go`, `service.go`, `handler.go`
3. Define interfaces at consumer level (handler defines Service interface, etc.)
4. Wire up in `config/routes.go`
5. Add migration if needed

### Adding External Service Integration

1. Define interface in the domain that uses it: `internal/payment/processor.go`
2. Create implementation: `internal/payment/stripe.go`
3. Inject via constructor in `config/routes.go`
4. Write integration tests with real service in test mode

## What NOT To Do

- Don't use `log` package — use `slog`
- Don't create dependencies inside functions — inject via constructor
- Don't put business logic in handlers
- Don't define interfaces in the implementing package
- Don't use `panic` for error handling
- Don't skip tests
- Don't modify files outside the scope of the current task

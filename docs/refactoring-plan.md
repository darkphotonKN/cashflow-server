# Cashflow Project Refactoring Plan

## Overview
Refactor the existing project to follow the clean architecture pattern defined in CLAUDE.md, implementing a personal financial tracking system with income and spending records.

## Data Models

### Core Entities

```go
// Transaction represents a single financial record
type Transaction struct {
    ID        uuid.UUID
    Date      time.Time
    Amount    float64
    Type      TransactionType
    CreatedAt time.Time
    UpdatedAt time.Time
}

// TransactionType enum
type TransactionType string
const (
    TransactionTypeSpending TransactionType = "spending"
    TransactionTypeEarning  TransactionType = "earning"
)

// AggregatedData for monthly summaries
type AggregatedData struct {
    Month     string  // format: "2024-01"
    Income    float64
    Spending  float64
    NetTotal  float64
}
```

## Project Structure Changes

### 1. Remove Unnecessary Files
- `/cmd/mcp-server/` - Remove MCP server components
- `/internal/mcp/` - Remove MCP-specific code
- `/internal/figma/` - Remove Figma integration
- `/internal/generator/` - Remove code generation tools
- `/internal/templates/` - Remove template files
- `client.go` - Remove client code

### 2. Create New Structure

```
cashflow/
├── cmd/
│   └── main.go              # Simple entrypoint
├── config/
│   ├── database.go          # PostgreSQL connection + migration runner
│   └── routes.go            # HTTP routes + dependency injection
├── internal/
│   ├── financial/           # Main domain
│   │   ├── model.go         # Transaction, TransactionType entities
│   │   ├── repository.go    # Database operations
│   │   ├── service.go       # Business logic
│   │   └── handler.go       # HTTP endpoints
│   ├── middleware/
│   │   └── cors.go          # CORS middleware
│   └── util/
│       └── response.go      # JSON response helpers
├── migrations/
│   └── 001_create_transactions_table.sql
├── docker-compose.yml       # PostgreSQL + Redis services
├── .env.example
├── .air.toml               # Hot reload config
├── .golangci.yml          # Linter config
└── Makefile               # Build commands
```

## Implementation Details

### Financial Domain Components

#### 1. Handler (internal/financial/handler.go)
**Endpoints:**
- `POST /api/transactions` - Create new transaction
- `GET /api/transactions` - List all transactions with pagination
- `GET /api/transactions/aggregate` - Get monthly aggregates with query filters

**Handler Interface Requirements:**
```go
type Service interface {
    CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*Transaction, error)
    ListTransactions(ctx context.Context, limit, offset int) ([]*Transaction, int64, error)
    GetMonthlyAggregate(ctx context.Context, month string) (*AggregatedData, error)
}
```

#### 2. Service (internal/financial/service.go)
**Business Logic:**
- Validation of transaction data
- Calculation of aggregates
- Date formatting and parsing

**Service Interface Requirements:**
```go
type Repository interface {
    Create(ctx context.Context, transaction *Transaction) error
    List(ctx context.Context, limit, offset int) ([]*Transaction, error)
    Count(ctx context.Context) (int64, error)
    GetByMonth(ctx context.Context, year int, month int) ([]*Transaction, error)
}
```

#### 3. Repository (internal/financial/repository.go)
**Database Operations:**
- CRUD operations for transactions
- Month-based filtering queries
- Aggregation queries

### API Specification

#### Create Transaction
```http
POST /api/transactions
Content-Type: application/json

{
    "date": "2024-01-15",
    "amount": 150.50,
    "type": "spending"
}

Response: 201 Created
{
    "id": "uuid",
    "date": "2024-01-15T00:00:00Z",
    "amount": 150.50,
    "type": "spending",
    "created_at": "2024-01-15T10:30:00Z"
}
```

#### List Transactions
```http
GET /api/transactions?limit=20&offset=0

Response: 200 OK
{
    "transactions": [...],
    "total": 150,
    "limit": 20,
    "offset": 0
}
```

#### Get Monthly Aggregate
```http
GET /api/transactions/aggregate?month=2024-01

Response: 200 OK
{
    "month": "2024-01",
    "income": 5000.00,
    "spending": 3500.00,
    "net_total": 1500.00
}
```

## Database Schema

```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('spending', 'earning')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_date_type ON transactions(date, type);
```

## Environment Configuration

### .env.example
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=cashflow
DB_PASSWORD=cashflow_password
DB_NAME=cashflow_db

# Server
PORT=8080
ENV=development

# Optional
LOG_LEVEL=info
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: cashflow
      POSTGRES_PASSWORD: cashflow_password
      POSTGRES_DB: cashflow_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## Migration Strategy

1. **Phase 1: Clean Up**
   - Remove all MCP/Figma related code
   - Remove unnecessary dependencies from go.mod

2. **Phase 2: Core Implementation**
   - Implement financial domain (model, repository, service, handler)
   - Set up database configuration
   - Create migrations

3. **Phase 3: Configuration**
   - Update main.go entrypoint
   - Configure routes with dependency injection
   - Set up middleware

4. **Phase 4: Testing & Documentation**
   - Add basic integration tests
   - Update README with API documentation

## Dependencies to Keep/Add

### Keep:
- github.com/google/uuid
- github.com/lib/pq (PostgreSQL driver)
- Standard library packages

### Add:
- github.com/golang-migrate/migrate/v4 (for migrations)

### Remove:
- All MCP-related packages
- Figma API packages
- Template/generation packages

## Success Criteria

1. ✅ Clean project structure following CLAUDE.md guidelines
2. ✅ Working CRUD operations for transactions
3. ✅ Monthly aggregation endpoint
4. ✅ PostgreSQL integration with migrations
5. ✅ Hot reload with air
6. ✅ Proper error handling and logging with slog
7. ✅ Docker-compose for local development
8. ✅ All tests pass with `make test`
9. ✅ All linting passes with `make lint`

## Estimated File Count
- ~15 files total (excluding .git, node_modules, binaries)
- Core business logic in 4 files (model, repository, service, handler)
- Configuration in 2 files (database, routes)
- 1 migration file
- Supporting files (Makefile, docker-compose, .env, etc.)
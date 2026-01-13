# Gin Framework Refactoring Plan

## Overview
Refactor the existing HTTP API from Go's standard library `net/http` to use the Gin framework for better performance, middleware support, and cleaner route handling.

## Benefits of Gin
- **Performance**: Gin is one of the fastest HTTP frameworks for Go
- **Middleware**: Built-in middleware ecosystem (CORS, logging, recovery)
- **JSON Binding**: Automatic request/response JSON marshaling
- **Route Groups**: Clean API versioning and organization
- **Parameter Binding**: Automatic path/query parameter extraction
- **Error Handling**: Better error handling and response patterns

## Current vs New Architecture

### Current (net/http)
```go
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
    var req CreateTransactionRequest
    json.NewDecoder(r.Body).Decode(&req)
    // manual error handling
    h.respondWithJSON(w, code, data)
}
```

### New (Gin)
```go
func (h *Handler) CreateTransaction(c *gin.Context) {
    var req CreateTransactionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // automatic JSON response
    c.JSON(201, transaction)
}
```

## Implementation Changes

### 1. Handler Signature Updates

#### Before
```go
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request)
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request)
func (h *Handler) GetMonthlyAggregate(w http.ResponseWriter, r *http.Request)
func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request)
```

#### After
```go
func (h *Handler) CreateTransaction(c *gin.Context)
func (h *Handler) ListTransactions(c *gin.Context)
func (h *Handler) GetMonthlyAggregate(c *gin.Context)
func (h *Handler) DeleteTransaction(c *gin.Context)
```

### 2. Route Registration

#### Before (config/routes.go)
```go
mux := http.NewServeMux()
mux.HandleFunc("POST /api/transactions", handler.CreateTransaction)
mux.HandleFunc("GET /api/transactions", handler.ListTransactions)
mux.HandleFunc("GET /api/transactions/aggregate", handler.GetMonthlyAggregate)
mux.HandleFunc("DELETE /api/transactions/{id}", handler.DeleteTransaction)
return middleware.CORS(mux)
```

#### After
```go
router := gin.New()
router.Use(gin.Logger(), gin.Recovery())
router.Use(corsMiddleware())

api := router.Group("/api")
{
    transactions := api.Group("/transactions")
    {
        transactions.POST("", handler.CreateTransaction)
        transactions.GET("", handler.ListTransactions)
        transactions.GET("/aggregate", handler.GetMonthlyAggregate)
        transactions.DELETE("/:id", handler.DeleteTransaction)
    }
}
return router
```

### 3. Request/Response Handling

#### Query Parameters
```go
// Before
limitStr := r.URL.Query().Get("limit")
limit := 20
if limitStr != "" {
    if l, err := strconv.Atoi(limitStr); err == nil {
        limit = l
    }
}

// After
limit := c.DefaultQuery("limit", "20")
limitInt, _ := strconv.Atoi(limit)
```

#### Path Parameters
```go
// Before
idStr := r.PathValue("id")

// After
idStr := c.Param("id")
```

#### JSON Responses
```go
// Before
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)
json.NewEncoder(w).Encode(payload)

// After
c.JSON(code, payload)
```

### 4. Error Handling

#### Before
```go
func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
    h.respondWithJSON(w, code, map[string]string{"error": message})
}
```

#### After
```go
func (h *Handler) handleError(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{"error": message})
}
// Or use Gin's built-in error handling
c.AbortWithStatusJSON(code, gin.H{"error": message})
```

### 5. Middleware Updates

#### CORS Middleware
```go
// Replace custom CORS middleware with Gin CORS
import "github.com/gin-contrib/cors"

func corsConfig() cors.Config {
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"*"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Content-Type", "Authorization"}
    return config
}
```

### 6. Request Validation

#### Enhanced Validation Tags
```go
type CreateTransactionRequest struct {
    Date        string          `json:"date" binding:"required"`
    Amount      float64         `json:"amount" binding:"required,gt=0"`
    Type        TransactionType `json:"type" binding:"required,oneof=spending earning"`
    Description string          `json:"description"`
    ImageBase64 string          `json:"image_base64,omitempty"`
}
```

## File Changes Required

### 1. Update Dependencies
```bash
go get github.com/gin-gonic/gin@latest
go get github.com/gin-contrib/cors@latest
```

### 2. Files to Modify

#### internal/financial/handler.go
- Change all handler method signatures
- Replace request/response logic with Gin context
- Use Gin's JSON binding and response methods
- Update path parameter extraction
- Replace custom error handling with Gin patterns

#### config/routes.go
- Replace `http.NewServeMux()` with `gin.New()`
- Set up route groups for better organization
- Configure Gin middleware (logger, recovery, CORS)
- Update route registration syntax

#### cmd/main.go
- Update server setup to use Gin router
- Maintain graceful shutdown logic

#### Remove Files
- `internal/middleware/cors.go` (replace with gin-contrib/cors)

### 3. Enhanced Features to Add

#### Structured Logging Middleware
```go
func structuredLogger(logger *slog.Logger) gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        logger.Error("panic recovered",
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.Any("panic", recovered))
        c.AbortWithStatus(500)
    })
}
```

#### Request ID Middleware
```go
func requestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}
```

#### Input Validation
```go
func (h *Handler) CreateTransaction(c *gin.Context) {
    var req CreateTransactionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }
    // Business logic remains the same
}
```

## Migration Steps

1. **Add Gin Dependencies**
   - Add gin-gonic/gin and gin-contrib/cors

2. **Update Handler Methods**
   - Change signatures to use `*gin.Context`
   - Replace manual JSON handling with Gin methods
   - Update parameter extraction

3. **Refactor Route Setup**
   - Replace ServeMux with Gin router
   - Set up route groups
   - Configure middleware

4. **Update Main Server**
   - Integrate Gin router with existing server setup
   - Maintain graceful shutdown

5. **Add Enhanced Middleware**
   - Request logging
   - CORS handling
   - Error recovery

6. **Testing & Validation**
   - Verify all endpoints work correctly
   - Test error handling
   - Validate middleware functionality

## API Endpoints (Post-Migration)

```
GET    /health                    # Health check
POST   /api/transactions          # Create transaction
GET    /api/transactions          # List transactions
GET    /api/transactions/aggregate # Monthly aggregates
DELETE /api/transactions/:id      # Delete transaction
```

## Benefits After Migration

1. **Cleaner Code**: Less boilerplate for JSON handling
2. **Better Error Handling**: Consistent error responses
3. **Enhanced Middleware**: Request logging, CORS, recovery
4. **Input Validation**: Automatic request validation with tags
5. **Performance**: Faster routing and middleware execution
6. **Maintainability**: More organized route structure
7. **Developer Experience**: Better debugging and logging

## Compatibility

- All existing API contracts remain the same
- No breaking changes to external interface
- Enhanced error responses with better structure
- Maintains all current functionality (S3, transactions, aggregates)

## Success Criteria

- ✅ All endpoints respond correctly
- ✅ JSON binding works for all requests
- ✅ Path/query parameters extracted properly
- ✅ Error handling improved
- ✅ CORS functionality maintained
- ✅ Graceful shutdown still works
- ✅ S3 image upload/delete functionality preserved
- ✅ Performance improved or maintained
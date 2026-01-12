# Image Upload & Description Feature Plan

## Overview

Add image upload capability with S3 storage and description field to each transaction entry. This will allow users to attach receipts, invoices, or other relevant images to their financial transactions along with descriptive notes.

## Data Model Updates

### Transaction Model Changes

```go
type Transaction struct {
    ID          uuid.UUID       `json:"id"`
    Date        time.Time       `json:"date"`
    Amount      float64         `json:"amount"`
    Type        TransactionType `json:"type"`
    Description string          `json:"description"`      // NEW
    ImageURL    string          `json:"image_url"`        // NEW: S3 URL
    ImageKey    string          `json:"image_key"`        // NEW: S3 object key
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateTransactionRequest struct {
    Date        string          `json:"date"`
    Amount      float64         `json:"amount"`
    Type        TransactionType `json:"type"`
    Description string          `json:"description"`      // NEW
    ImageBase64 string          `json:"image_base64"`     // NEW: Base64 encoded image
}
```

## Architecture Components

### 1. New S3 Package Structure

```
internal/
├── s3/
│   ├── client.go        # S3 client wrapper
│   ├── service.go       # S3 service interface & implementation
│   └── config.go        # S3 configuration
```

### 2. S3 Service Interface

```go
type S3Service interface {
    UploadImage(ctx context.Context, imageData []byte, contentType string) (url string, key string, error)
    DeleteImage(ctx context.Context, key string) error
    GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
}
```

## Implementation Steps

### Phase 1: Database Updates

#### Migration 002_add_image_and_description.sql

```sql
ALTER TABLE transactions
ADD COLUMN description TEXT,
ADD COLUMN image_url TEXT,
ADD COLUMN image_key TEXT;

CREATE INDEX idx_transactions_image_key ON transactions(image_key);
```

### Phase 2: S3 Integration

#### S3 Configuration

```go
// internal/s3/config.go
type Config struct {
    Region          string
    BucketName      string
    AccessKeyID     string
    SecretAccessKey string
    URLExpiration   time.Duration
}
```

#### S3 Client Implementation

- AWS SDK v2 for Go
- Support for image upload with unique keys
- Generate pre-signed URLs for secure access
- Handle multiple image formats (JPEG, PNG, WebP)

### Phase 3: Update Financial Domain

#### Repository Updates

```go
type Repository interface {
    Create(ctx context.Context, transaction *Transaction) error
    List(ctx context.Context, limit, offset int) ([]*Transaction, error)
    Count(ctx context.Context) (int64, error)
    GetByMonth(ctx context.Context, year int, month int) ([]*Transaction, error)
    Delete(ctx context.Context, id uuid.UUID) (*Transaction, error) // NEW: for cleanup
}
```

#### Service Updates

```go
type Service interface {
    CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*Transaction, error)
    ListTransactions(ctx context.Context, limit, offset int) ([]*Transaction, int64, error)
    GetMonthlyAggregate(ctx context.Context, month string) (*AggregatedData, error)
    DeleteTransaction(ctx context.Context, id uuid.UUID) error // NEW
}

// Service will need S3Service dependency
type service struct {
    repo      Repository
    s3Service S3Service  // NEW
    logger    *slog.Logger
}
```

#### Handler Updates

- Accept multipart/form-data for image uploads
- Parse base64 encoded images from JSON
- Add DELETE endpoint for transaction cleanup

### Phase 4: API Changes

#### Updated Create Transaction Endpoint

```http
POST /api/transactions
Content-Type: application/json

{
    "date": "2024-01-15",
    "amount": 150.50,
    "type": "spending",
    "description": "Grocery shopping at Whole Foods",
    "image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
}

Response: 201 Created
{
    "id": "uuid",
    "date": "2024-01-15T00:00:00Z",
    "amount": 150.50,
    "type": "spending",
    "description": "Grocery shopping at Whole Foods",
    "image_url": "https://bucket.s3.region.amazonaws.com/...",
    "created_at": "2024-01-15T10:30:00Z"
}
```

#### Alternative: Multipart Form Upload

```http
POST /api/transactions/upload
Content-Type: multipart/form-data

Form Fields:
- date: "2024-01-15"
- amount: 150.50
- type: "spending"
- description: "Grocery shopping"
- image: [binary file data]
```

#### New Delete Transaction Endpoint

```http
DELETE /api/transactions/{id}

Response: 204 No Content
```

## Environment Variables

### New S3 Configuration

```env
# Existing
DB_HOST=localhost
DB_PORT=5432
DB_USER=cashflow
DB_PASSWORD=cashflow_password
DB_NAME=cashflow_db
PORT=8080

# New S3 Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
S3_BUCKET_NAME=cashflow-images
S3_URL_EXPIRATION=24h

# Optional
MAX_IMAGE_SIZE=10485760  # 10MB in bytes
ALLOWED_IMAGE_TYPES=image/jpeg,image/png,image/webp
```

## Dependencies to Add

```go
require (
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/aws/aws-sdk-go-v2/config v1.26.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.47.0
    github.com/aws/aws-sdk-go-v2/credentials v1.16.0
)
```

## Image Processing Flow

1. **Upload Process:**

   - Client sends base64 encoded image or multipart form
   - Validate image size (max 10MB)
   - Validate image type (JPEG, PNG, WebP)
   - Decode base64 if needed
   - Generate unique S3 key: `transactions/{year}/{month}/{uuid}_{timestamp}.{ext}`
   - Upload to S3 with metadata
   - Store S3 URL and key in database

2. **Retrieval Process:**

   - Fetch transaction with image URL
   - If URL expired, generate new pre-signed URL
   - Return transaction with valid image URL

3. **Deletion Process:**
   - Delete image from S3 using key
   - Remove transaction from database
   - Handle orphaned images with cleanup job

## Security Considerations

1. **Image Validation:**

   - File size limits (10MB max)
   - Content-type validation
   - Virus scanning (optional, using AWS Lambda)

2. **S3 Security:**

   - Private bucket with no public access
   - Pre-signed URLs with expiration
   - IAM roles with minimal permissions
   - Server-side encryption (AES-256)

3. **Access Control:**
   - URLs expire after 24 hours
   - Regenerate URLs on demand
   - No direct S3 bucket access

## Error Handling

1. **Upload Failures:**

   - Rollback database transaction if S3 upload fails
   - Return appropriate error messages
   - Log failures for monitoring

2. **Missing Images:**
   - Handle gracefully when S3 object not found
   - Provide fallback/placeholder URL
   - Flag for cleanup

## Testing Strategy

1. **Unit Tests:**

   - Mock S3 service for domain tests
   - Test image validation logic
   - Test URL generation

2. **Integration Tests:**

   - LocalStack for S3 testing
   - Test full upload/download flow
   - Test cleanup operations

3. **Manual Testing:**
   - Test with various image formats
   - Test large file handling
   - Test error scenarios

## Migration Rollback Plan

If issues occur:

1. Keep image fields nullable initially
2. Maintain backward compatibility
3. Migration to remove fields if needed:

```sql
ALTER TABLE transactions
DROP COLUMN description,
DROP COLUMN image_url,
DROP COLUMN image_key;
```

## Implementation Order

1. **Phase 1:** Database migration (add fields)
2. **Phase 2:** S3 service implementation
3. **Phase 3:** Update financial domain with S3 integration
4. **Phase 4:** Update API handlers
5. **Phase 5:** Testing & validation
6. **Phase 6:** Documentation update

## Success Criteria

- ✅ Transactions support description field
- ✅ Images upload successfully to S3
- ✅ Pre-signed URLs work correctly
- ✅ Image deletion removes from S3
- ✅ All existing functionality still works
- ✅ Tests pass with new features
- ✅ No security vulnerabilities
- ✅ Performance acceptable (<2s for upload)

## Alternative Considerations

### Option A: Direct S3 Upload from Client

- Client gets pre-signed POST URL
- Upload directly to S3
- Server only stores S3 key
- Pros: Less server load, faster uploads
- Cons: More complex client implementation

### Option B: Local File Storage

- Store images on server disk
- Serve via static file handler
- Pros: Simpler, no AWS dependency
- Cons: Not scalable, backup complexity

### Recommendation:

Go with server-side S3 upload (current plan) for better control and simpler client implementation. Can migrate to direct upload later if needed.

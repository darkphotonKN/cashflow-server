# Postman Collection Setup Guide

## Import the Collection

1. **Open Postman**
2. **Click "Import"** in the top left
3. **Select "Upload Files"**
4. **Choose** `docs/Cashflow-API.postman_collection.json`
5. **Click "Import"**

## Environment Setup

### Option 1: Use Collection Variables (Recommended)
The collection comes with a pre-configured variable:
- `base_url`: `http://localhost:8080`

### Option 2: Create Custom Environment
1. Click the **Environment** dropdown (top right)
2. Select **"Manage Environments"**
3. Click **"Add"**
4. Set environment name: `Cashflow Local`
5. Add variable:
   - Key: `base_url`
   - Value: `http://localhost:8080`
6. **Save**

## API Endpoints Overview

### 1. Health Check
- **GET** `/health`
- **Purpose**: Verify API is running
- **No parameters required**

### 2. Create Transaction (Basic)
- **POST** `/api/transactions`
- **Purpose**: Add new transaction without image
- **Required Fields**:
  ```json
  {
    "date": "2024-01-15",        // YYYY-MM-DD format
    "amount": 150.50,            // Positive number
    "type": "spending"           // "spending" or "earning"
  }
  ```
- **Optional Fields**:
  ```json
  {
    "description": "Grocery shopping at Whole Foods"
  }
  ```

### 3. Create Transaction (With Image)
- **POST** `/api/transactions`
- **Purpose**: Add transaction with receipt/invoice image
- **Additional Field**:
  ```json
  {
    "image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
  }
  ```

### 4. List Transactions
- **GET** `/api/transactions`
- **Query Parameters**:
  - `limit`: Number per page (1-100, default: 20)
  - `offset`: Skip count for pagination (default: 0)
- **Example**: `/api/transactions?limit=10&offset=20`

### 5. Monthly Aggregate
- **GET** `/api/transactions/aggregate`
- **Query Parameters**:
  - `month`: YYYY-MM format (required)
- **Example**: `/api/transactions/aggregate?month=2024-01`

### 6. Delete Transaction
- **DELETE** `/api/transactions/{id}`
- **Path Parameter**: Transaction UUID
- **Note**: Also deletes associated S3 image

## Field Details & Validation

### Transaction Fields

| Field | Type | Required | Validation | Example |
|-------|------|----------|------------|---------|
| `date` | string | Yes | YYYY-MM-DD format | "2024-01-15" |
| `amount` | number | Yes | Must be > 0 | 150.50 |
| `type` | string | Yes | "spending" or "earning" | "spending" |
| `description` | string | No | Any text | "Coffee at Starbucks" |
| `image_base64` | string | No | Base64 encoded image | "data:image/jpeg;base64,..." |

### Image Upload Guidelines

#### Supported Formats
- **JPEG** (.jpg, .jpeg)
- **PNG** (.png)
- **WebP** (.webp)

#### Size Limits
- **Maximum**: 10MB per image
- **Recommended**: Under 2MB for faster uploads

#### Base64 Encoding Options

**Option 1: With Data URL Prefix**
```json
{
  "image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD..."
}
```

**Option 2: Raw Base64**
```json
{
  "image_base64": "/9j/4AAQSkZJRgABAQEAYABgAAD..."
}
```

#### How to Get Base64

**Online Converters:**
- [base64encode.org](https://www.base64encode.org/)
- [base64-image.de](https://www.base64-image.de/)

**Command Line:**
```bash
# Mac/Linux
base64 -i image.jpg

# Windows
certutil -encode image.jpg base64.txt
```

**Browser JavaScript:**
```javascript
// File input element
const fileInput = document.getElementById('fileInput');
const file = fileInput.files[0];
const reader = new FileReader();
reader.onload = function(e) {
    const base64 = e.target.result; // Includes data URL prefix
    console.log(base64);
};
reader.readAsDataURL(file);
```

## Testing Workflow

### 1. Start Local Server
```bash
make dev  # or go run cmd/main.go
```

### 2. Test Health Check
- Run the "Health Check" request
- Should return: `{"status": "ok"}`

### 3. Create Sample Transactions
1. **Basic Spending**: Use "Create Transaction (Basic)"
2. **Income**: Use "Create Earning Transaction"
3. **With Image**: Use "Create Transaction (With Image)"

### 4. List & Verify
- Run "List Transactions" to see all created transactions
- Verify images have valid S3 URLs

### 5. Get Monthly Summary
- Run "Get Monthly Aggregate" with current month
- Verify income, spending, and net totals

### 6. Clean Up
- Use "Delete Transaction" to remove test data
- Verify images are removed from S3

## Error Handling

### Common Error Responses

**400 Bad Request**
```json
{
  "error": "Invalid request body",
  "details": "amount is required"
}
```

**404 Not Found**
```json
{
  "error": "transaction not found"
}
```

**500 Internal Server Error**
```json
{
  "error": "Failed to upload image to S3"
}
```

## Environment Configuration

Ensure your `.env` file contains:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=cashflow
DB_PASSWORD=cashflow_password
DB_NAME=cashflow_db

# Server
PORT=8080

# AWS S3 (Required for image uploads)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
S3_BUCKET_NAME=cashflow-images
```

## Troubleshooting

### Server Won't Start
- Check database connection
- Verify AWS credentials (for S3)
- Ensure port 8080 is available

### Image Upload Fails
- Verify AWS credentials are correct
- Check S3 bucket exists and has write permissions
- Ensure image is under 10MB
- Verify base64 encoding is correct

### Database Errors
- Run migrations: `psql -d cashflow_db -f migrations/001_create_transactions_table.sql`
- Check database connection settings
- Verify PostgreSQL is running

## Collection Features

### Pre-filled Examples
- Each request includes realistic example data
- Response examples show expected formats
- Detailed descriptions for every field

### Documentation
- Field validation rules
- Error handling examples
- Usage guidelines for each endpoint

### Variables
- Easy environment switching
- Configurable base URL
- Ready for production deployment

## Next Steps

1. **Import the collection**
2. **Start your local server**
3. **Test the health endpoint**
4. **Create your first transaction**
5. **Explore the API functionality**

The collection is ready to use with your Cashflow API! 🚀
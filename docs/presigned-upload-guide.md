# Presigned URL Upload Flow Guide

## Overview
The Cashflow API uses AWS S3 presigned URLs to enable direct image uploads from clients, bypassing the API server for better performance and scalability.

## Upload Flow

### Step 1: Request Upload URL
```bash
POST /api/uploads/request
{
    "content_type": "image/jpeg",
    "file_size": 1024000
}
```

Response:
```json
{
    "upload_id": "123e4567-e89b-12d3-a456-426614174000",
    "upload_url": "https://bucket.s3.amazonaws.com/staging/2024/01/...",
    "expires_at": "2024-01-01T12:15:00Z",
    "s3_key": "staging/2024/01/123e4567_1704067200.jpg"
}
```

### Step 2: Upload to S3
Use the presigned URL to upload directly to S3:

```bash
PUT {upload_url}
Content-Type: image/jpeg
Body: [binary image data]
```

### Step 3: Create Transaction
Include the upload_id when creating the transaction:

```bash
POST /api/transactions
{
    "date": "2024-01-15",
    "amount": 150.50,
    "type": "spending",
    "description": "Grocery shopping",
    "upload_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Implementation Examples

### JavaScript/TypeScript
```javascript
// Step 1: Request upload URL
const uploadRequest = await fetch('/api/uploads/request', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        content_type: file.type,
        file_size: file.size
    })
});
const { upload_id, upload_url } = await uploadRequest.json();

// Step 2: Upload to S3
await fetch(upload_url, {
    method: 'PUT',
    headers: { 'Content-Type': file.type },
    body: file
});

// Step 3: Create transaction
await fetch('/api/transactions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        date: '2024-01-15',
        amount: 150.50,
        type: 'spending',
        description: 'Grocery shopping',
        upload_id: upload_id
    })
});
```

### Swift (iOS)
```swift
// Step 1: Request upload URL
let uploadRequest = UploadRequest(contentType: "image/jpeg", fileSize: imageData.count)
let uploadResponse = try await apiClient.requestUpload(uploadRequest)

// Step 2: Upload to S3
var request = URLRequest(url: URL(string: uploadResponse.uploadUrl)!)
request.httpMethod = "PUT"
request.setValue("image/jpeg", forHTTPHeaderField: "Content-Type")
request.httpBody = imageData
let (_, _) = try await URLSession.shared.data(for: request)

// Step 3: Create transaction
let transaction = CreateTransactionRequest(
    date: "2024-01-15",
    amount: 150.50,
    type: .spending,
    description: "Grocery shopping",
    uploadId: uploadResponse.uploadId
)
try await apiClient.createTransaction(transaction)
```

### Go
```go
// Step 1: Request upload URL
uploadReq := UploadRequest{
    ContentType: "image/jpeg",
    FileSize:    len(imageData),
}
uploadResp, _ := client.RequestUpload(ctx, uploadReq)

// Step 2: Upload to S3
req, _ := http.NewRequest("PUT", uploadResp.UploadURL, bytes.NewReader(imageData))
req.Header.Set("Content-Type", "image/jpeg")
http.DefaultClient.Do(req)

// Step 3: Create transaction
transaction := CreateTransactionRequest{
    Date:        "2024-01-15",
    Amount:      150.50,
    Type:        "spending",
    Description: "Grocery shopping",
    UploadID:    uploadResp.UploadID,
}
client.CreateTransaction(ctx, transaction)
```

## Benefits

1. **Performance**: Images upload directly to S3, reducing server load
2. **Scalability**: API server doesn't handle large file transfers
3. **Reliability**: S3 handles retries and resumable uploads
4. **Cost-effective**: Reduces bandwidth costs on API server
5. **Mobile-friendly**: Standard HTTP PUT request works on all platforms

## Error Handling

### Upload Request Errors
- `400 Bad Request`: Invalid content type or file size
- `413 Payload Too Large`: File exceeds 10MB limit

### S3 Upload Errors
- `403 Forbidden`: Presigned URL expired (15 min timeout)
- `400 Bad Request`: Content-Type mismatch

### Transaction Creation Errors
- `404 Not Found`: Upload ID doesn't exist
- `400 Bad Request`: File not uploaded to S3 yet
- `409 Conflict`: Upload already linked to another transaction

## Security Notes

- Presigned URLs expire after 15 minutes
- Each upload_id can only be used once
- Files are moved from staging to production on transaction creation
- Orphaned uploads in staging can be cleaned up after 24 hours
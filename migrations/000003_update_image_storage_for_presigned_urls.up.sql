-- Update image storage for presigned URL flow
-- Remove image_url column since we'll generate URLs dynamically
ALTER TABLE transactions
DROP COLUMN IF EXISTS image_url;

-- Add upload tracking fields
ALTER TABLE transactions
ADD COLUMN upload_id VARCHAR(100),
ADD COLUMN image_upload_status VARCHAR(20) DEFAULT 'none' CHECK (image_upload_status IN ('none', 'pending', 'completed', 'failed')),
ADD COLUMN image_uploaded_at TIMESTAMP WITH TIME ZONE;

-- Create index for upload tracking
CREATE INDEX idx_transactions_upload_id ON transactions(upload_id) WHERE upload_id IS NOT NULL;
CREATE INDEX idx_transactions_upload_status ON transactions(image_upload_status) WHERE image_upload_status != 'none';

-- Create uploads tracking table for orphan cleanup
CREATE TABLE IF NOT EXISTS upload_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    upload_id VARCHAR(100) UNIQUE NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'expired')),
    presigned_url_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL
);

CREATE INDEX idx_upload_requests_status ON upload_requests(status);
CREATE INDEX idx_upload_requests_created_at ON upload_requests(created_at);
CREATE INDEX idx_upload_requests_transaction_id ON upload_requests(transaction_id);
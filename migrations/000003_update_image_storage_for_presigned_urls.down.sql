-- Revert presigned URL changes
DROP TABLE IF EXISTS upload_requests;

-- Remove upload tracking fields
ALTER TABLE transactions
DROP COLUMN IF EXISTS upload_id,
DROP COLUMN IF EXISTS image_upload_status,
DROP COLUMN IF EXISTS image_uploaded_at;

-- Re-add image_url column
ALTER TABLE transactions
ADD COLUMN image_url TEXT;

-- Drop indexes
DROP INDEX IF EXISTS idx_transactions_upload_id;
DROP INDEX IF EXISTS idx_transactions_upload_status;
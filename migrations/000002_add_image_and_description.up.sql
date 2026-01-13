-- Add description and image fields to transactions table
ALTER TABLE transactions
ADD COLUMN description TEXT,
ADD COLUMN image_url TEXT,
ADD COLUMN image_key TEXT;

-- Create index for faster lookups by image key
CREATE INDEX idx_transactions_image_key ON transactions(image_key) WHERE image_key IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN transactions.description IS 'User-provided description of the transaction';
COMMENT ON COLUMN transactions.image_url IS 'Pre-signed S3 URL for the transaction image';
COMMENT ON COLUMN transactions.image_key IS 'S3 object key for the transaction image';
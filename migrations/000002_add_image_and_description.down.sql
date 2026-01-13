-- Remove image and description fields from transactions table
ALTER TABLE transactions
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS image_url,
DROP COLUMN IF EXISTS image_key;

-- Drop index
DROP INDEX IF EXISTS idx_transactions_image_key;
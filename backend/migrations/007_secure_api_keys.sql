-- Remove plain text API key storage for security
-- This migration removes the insecure `key` field and keeps only the hashed version

BEGIN;

-- First, check if we have any existing data and warn
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM api_keys WHERE key IS NOT NULL) THEN
        RAISE NOTICE 'Found existing API keys - they will need to be recreated after this migration';
    END IF;
END $$;

-- Remove the insecure unique index on the plain text key field
DROP INDEX IF EXISTS idx_api_keys_key;

-- Remove the plain text key column (security risk)
ALTER TABLE api_keys DROP COLUMN IF EXISTS key;

-- Ensure the key_hash field has the unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash) WHERE deleted_at IS NULL;

-- Add a comment to document the security improvement
COMMENT ON COLUMN api_keys.key_hash IS 'Bcrypt hash of the API key - never store plain text keys for security';

COMMIT;
-- Remove legal consent tracking columns from users table
ALTER TABLE users
    DROP COLUMN IF EXISTS terms_accepted_at,
    DROP COLUMN IF EXISTS privacy_accepted_at;

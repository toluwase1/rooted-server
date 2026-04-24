-- Add legal consent tracking columns to users table
ALTER TABLE users
    ADD COLUMN terms_accepted_at TIMESTAMPTZ,
    ADD COLUMN privacy_accepted_at TIMESTAMPTZ;

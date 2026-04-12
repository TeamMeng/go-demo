-- Add unique constraint on phone column
-- NULL values are allowed and multiple NULLs are permitted by PostgreSQL
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone) WHERE phone IS NOT NULL;

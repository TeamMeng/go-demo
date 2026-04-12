-- Add phone column to users table
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);

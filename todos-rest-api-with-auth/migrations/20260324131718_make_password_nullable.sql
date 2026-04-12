-- Make password column nullable to support phone-only users (no password needed for SMS login)
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

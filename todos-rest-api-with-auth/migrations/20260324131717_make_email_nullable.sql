-- Make email column nullable to support phone-only users
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

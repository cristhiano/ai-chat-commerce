-- Initialize database for chat-ecommerce application
-- This script runs when the PostgreSQL container starts for the first time

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database if it doesn't exist (this runs in the context of the default database)
-- The database name is set via POSTGRES_DB environment variable

-- Set timezone
SET timezone = 'UTC';

-- Create a user for the application (optional, using postgres user for now)
-- CREATE USER chat_ecommerce_user WITH PASSWORD 'secure_password';
-- GRANT ALL PRIVILEGES ON DATABASE chat_ecommerce TO chat_ecommerce_user;

-- The actual tables will be created by GORM auto-migration
-- This script is mainly for initial setup and extensions

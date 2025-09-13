-- SmartEdify Auth Service - Database Initialization
-- This script creates the initial database structure

-- Create database if it doesn't exist (this is handled by POSTGRES_DB env var)
-- CREATE DATABASE IF NOT EXISTS smartedify_auth;

-- Connect to the database
\c smartedify_auth;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Basic health check table
CREATE TABLE IF NOT EXISTS health_check (
    id SERIAL PRIMARY KEY,
    status VARCHAR(50) NOT NULL DEFAULT 'healthy',
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert initial health check record
INSERT INTO health_check (status) VALUES ('healthy') ON CONFLICT DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE smartedify_auth TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;

-- Log successful initialization
DO $$
BEGIN
    RAISE NOTICE 'SmartEdify Auth Service database initialized successfully';
END $$;
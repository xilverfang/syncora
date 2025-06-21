-- Create the database and user
-- Note: The database syncora_db is already created by POSTGRES_DB environment variable
-- So we just need to set up the user and permissions
SELECT 'CREATE DATABASE syncora_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'syncora_db')\gexec

-- Create user if it doesn't exist (it should already exist from POSTGRES_USER)
-- But let's make sure it has the right permissions
ALTER USER syncora CREATEDB;

-- Connect to the syncora_db database
\c syncora_db;

-- Create the accounts table
CREATE TABLE IF NOT EXISTS accounts (
    address TEXT PRIMARY KEY,
    alias TEXT NOT NULL,
    encrypted_key TEXT NOT NULL,
    salt TEXT NOT NULL DEFAULT '',
    key_version SMALLINT NOT NULL DEFAULT 1,
    CONSTRAINT valid_hex_encrypted_key CHECK (encrypted_key ~ '^[0-9a-fA-F]+$'),
    CONSTRAINT valid_hex_salt CHECK (salt ~ '^[0-9a-fA-F]+$'),
    CONSTRAINT valid_key_version CHECK (key_version >= 1)
);

-- Grant permissions to syncora user
GRANT ALL PRIVILEGES ON DATABASE syncora_db TO syncora;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO syncora;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO syncora;

-- Ensure future tables also have permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO syncora;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO syncora;
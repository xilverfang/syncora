Syncora CLI Architecture
Syncora is a command-line interface (CLI) tool designed to interact with blockchain bridge services, enabling users to manage accounts, check supported bridges and networks, and execute token bridging operations. This document outlines the system’s architecture, focusing on its components, data flow, security mechanisms, and deployment setup. It’s intended for developers, contributors, and users seeking to understand Syncora’s design.
Overview
Syncora is a Go-based CLI that stores user accounts securely in a PostgreSQL database, encrypts private keys using Argon2 and AES, and runs in a Dockerized environment for consistency. The system comprises three main layers:

CLI Layer: Handles user commands (account, info, help) via a Cobra-based interface.
Core Layer: Manages business logic, including cryptography (crypto) and database operations (database).
Infrastructure Layer: Runs the CLI and PostgreSQL in Docker containers, with SSL-secured communication.

The project is located at /Users/mac/Documents/syncora, with a structure optimized for modularity and security.
System Components
1. CLI Layer (cmd/bridge/)
The CLI layer is the user interface, implemented in Go using the Cobra framework (github.com/spf13/cobra). It processes commands and delegates tasks to the core layer.

Main Entry Point (cmd/bridge/main.go):
Initializes the Cobra root command and sets up subcommands.
Loads environment variables from .env (e.g., SYNCORA_DB_URL).


Commands (cmd/bridge/internal/commands/):
account.go: Manages accounts (import, list, remove).
Example: syncora account import --alias test-wallet encrypts and stores private keys.


info.go: Retrieves account or bridge info (check).
Example: syncora info check --account test-wallet decrypts and verifies keys.


help.go: Displays usage info (help).
Example: syncora help lists commands and security tips.




User Input:
Uses golang.org/x/term for secure, hidden input (e.g., private keys, passphrases).
Enforces passphrase strength (minimum 8 characters, recommended 12+ with complexity).



2. Core Layer (internal/core/)
The core layer handles business logic, split into two modules:

Crypto (internal/core/crypto/crypto.go):

Functionality: Encrypts/decrypts private keys using AES-256-GCM, with keys derived via Argon2id.
Security:
Argon2 parameters: time=1, memory=32MB, threads=4.
Generates random salts for key derivation.
Uses HMAC for integrity checks.
Supports key versioning (key_version=1).


Example: EncryptPrivateKey derives a key from a user passphrase, encrypts the private key, and returns the ciphertext, salt, and address.


Database (internal/core/database/database.go):

Functionality: Stores accounts in a PostgreSQL table (accounts).
Schema:CREATE TABLE accounts (
    address TEXT PRIMARY KEY,
    alias TEXT NOT NULL,
    encrypted_key TEXT NOT NULL,
    salt TEXT NOT NULL DEFAULT '',
    key_version SMALLINT NOT NULL DEFAULT 1,
    CONSTRAINT valid_hex_encrypted_key CHECK (encrypted_key ~ '^[0-9a-fA-F]+$'),
    CONSTRAINT valid_hex_salt CHECK (salt ~ '^[0-9a-fA-F]+$'),
    CONSTRAINT valid_key_version CHECK (key_version >= 1)
);


Operations: SaveAccount, ListAccounts, GetAccount, RemoveAccount.
Migration: Automatically adds salt and key_version columns if missing.
Security: Uses SSL (sslmode=verify-ca) and connection pooling (max_open_conns=10).



3. Infrastructure Layer
The infrastructure layer deploys Syncora using Docker, ensuring portability and consistency.

Dockerfile (Dockerfile):

Build Stage: Uses golang:1.24.4-alpine to compile the CLI binary (/app/bin/syncora).
Runtime Stage: Uses alpine:3.18, includes ca-certificates for SSL.
Structure:FROM golang:1.24.4-alpine AS builder
COPY . .
RUN go build -o /app/bin/syncora ./cmd/bridge
FROM alpine:3.18
COPY --from=builder /app/bin/syncora /app/bin/syncora
ENTRYPOINT ["/app/bin/syncora"]
CMD ["help"]




docker-compose.yml (docker-compose.yml):

# Services:
syncora: Runs the CLI, depends on postgres, mounts .env and certs/.
postgres: Runs PostgreSQL 15, with SSL and pgaudit enabled.


# Networking: Uses syncora-net for isolated communication.
Volumes: Persists PostgreSQL data (postgres-data), mounts init-db.sql and certs.
Healthcheck: Ensures postgres is ready (pg_isready and SELECT on accounts).


# Initialization (_init-db.sql_):

Creates syncora role, syncora_db, and accounts table.
Grants permissions to syncora.


# Certificates (certs/):

Includes _client.crt, client.key, ca.crt, server.crt, server.key_.
Configured in pg_hba.conf for client certificate authentication.



# Data Flow
Here’s how data flows through Syncora for a typical operation (syncora account import --alias test-wallet):

# User Input:
CLI prompts for private key and passphrase (hidden via term.ReadPassword).


# Crypto Processing:
crypto.EncryptPrivateKey derives an encryption key using Argon2, generates a salt, encrypts the private key with AES-256-GCM, and computes the Ethereum address.


# Database Storage:
_database.SaveAccount_ stores the address, alias, encrypted key, salt, and key_version in the accounts table via PostgreSQL (SSL-secured).


### Output:
_CLI displays success message with masked address (e.g., 0xA46f88EE...)_.



Diagram (ASCII):
+-------------------+
| User (Terminal)   |
| syncora-cli ...   |
+-------------------+
         |
         v
+-------------------+
| CLI Layer         |
| cmd/bridge/       |
| (Cobra Commands)  |
+-------------------+
         |
         v
+-------------------+
| Core Layer        |
| crypto/ (Argon2)  |
| database/ (SQL)   |
+-------------------+
         |
         v
+-------------------+
| Infrastructure    |
| Docker: syncora   |
| PostgreSQL:       |
| syncora_db (SSL)  |
+-------------------+

# Security Mechanisms
Syncora prioritizes security for private key management and data storage:

# Encryption:
Private keys are encrypted with AES-256-GCM, using Argon2id-derived keys.
Random salts prevent rainbow table attacks.
HMAC ensures data integrity.


# Passphrase:
User-provided, minimum 8 characters (recommended 12+ with complexity).
Zeroed in memory after use (for i := range passphrase { passphrase[i] = 0 }).


# Database:
SSL with client certificates (sslmode=verify-ca, client.crt, client.key).
pgaudit logs write, read, and DDL operations.
Connection pooling limits resource usage.


# Docker:
.env permissions set to 0600.
Read-only mounts for certs and .env.
Isolated network (syncora-net).


# Input:
Hidden inputs prevent shoulder-surfing or shell history leaks.



Project Structure
The project is organized for clarity and maintainability:
/Users/mac/Documents/syncora/
├── cmd/
│   └── bridge/
│       ├── main.go
│       └── internal/
│           └── commands/
│               ├── account.go
│               ├── info.go
│               └── help.go
├── internal/
│   └── core/
│       ├── crypto/
│       │   └── crypto.go
│       └── database/
│           └── database.go
├── certs/
│   ├── client.crt
│   ├── client.key
│   ├── ca.crt
│   ├── server.crt
│   ├── server.key
│   └── pg_hba.conf
├── docs/
│   ├── api/
│   ├── architecture/
│   │   └── architecture.md
│   └── guide/
├── .env
├── Dockerfile
├── docker-compose.yml
├── init-db.sql
├── go.mod
├── go.sum
└── syncora-cli

Extensibility
Syncora is designed to support future features, such as:

API Gateway (services/api-gateway/): Will expose REST endpoints for CLI operations.
Bridge Adapters (services/bridge-adapter/): Will integrate with blockchain bridge services.
Config File: Potential config.yaml for API URLs, logging, or network settings.

# Deployment
Syncora runs in Docker for consistency:

# Setup:
Ensure certs/, .env, init-db.sql, and pg_hba.conf exist.
Run: docker-compose up -d.


# Usage:
Use the syncora-cli wrapper: syncora-cli account import --alias test-wallet.


# Updates:
Rebuild after code changes: docker-compose up --build -d.



Troubleshooting

Schema Errors: Check database.go migrations or init-db.sql.
Connection Issues: Verify SYNCORA_DB_URL and certs.
Logs: Use docker-compose logs syncora or docker-compose logs postgres.


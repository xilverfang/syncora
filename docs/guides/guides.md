Syncora CLI User Guide
Syncora is a command-line tool for interacting with blockchain bridge services, allowing you to manage accounts, check bridge and network support, and perform token bridging operations. This guide walks you through setting up Syncora, using its commands, and troubleshooting common issues. It’s designed for users of all levels, assuming no prior experience with Docker or blockchain tools.
Prerequisites

Operating System: macOS, Linux, or Windows with WSL2.
Docker: Docker Desktop (version 4.20+ recommended) installed and running.
Disk Space: At least 2GB for Docker images and PostgreSQL data.
Terminal: Basic familiarity with command-line interfaces (e.g., Terminal, Bash, or PowerShell).
Git: Optional, for cloning the project repository.

Setup
Syncora runs in Docker containers, ensuring consistency across environments. Follow these steps to set up Syncora at /Users/mac/Documents/syncora.
1. Clone or Create Project Directory
If you have the Syncora repository, clone it:
git clone <repository-url> /Users/mac/Documents/syncora
cd /Users/mac/Documents/syncora

Otherwise, create the directory:
mkdir -p /Users/mac/Documents/syncora
cd /Users/mac/Documents/syncora

2. Prepare SSL Certificates
Syncora uses SSL for secure database communication. Create or copy certificates to certs/:
mkdir -p certs

If certificates exist at /Users/mac/.syncora/certs/, copy them:
cp /Users/mac/.syncora/certs/{client.crt,client.key,ca.crt,server.crt,server.key,pg_hba.conf} certs/

Or generate self-signed certificates (for testing):
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=postgres"

openssl req -x509 -newkey rsa:4096 -keyout certs/client.key -out certs/client.crt -days 365 -nodes -subj "/CN=syncora"

cp certs/server.crt certs/ca.crt

cat > certs/pg_hba.conf << 'EOF'
hostssl syncora_db syncora 0.0.0.0/0 cert
hostssl syncora_db syncora ::0/0 cert
EOF

chmod 600 certs/*

3. Create Environment File
Create .env with the database connection string:
echo "SYNCORA_DB_URL=postgres://syncora:syncora123@postgres:5432/syncora_db?sslmode=verify-ca&sslcert=/app/certs/client.crt&sslkey=/app/certs/client.key&sslrootcert=/app/certs/ca.crt" > .env

chmod 600 .env

4. Create Database Initialization Script
Create init-db.sql to set up the database schema:
cat > init-db.sql << 'EOF'
CREATE ROLE syncora WITH LOGIN PASSWORD 'syncora123';
CREATE DATABASE syncora_db OWNER syncora;
\c syncora_db
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
GRANT SELECT, INSERT, UPDATE, DELETE ON accounts TO syncora;
EOF

5. Verify Project Files
Ensure the following files exist:

Dockerfile: Builds the CLI binary.
docker-compose.yml: Defines syncora and postgres services.
cmd/bridge/main.go: CLI entry point.
internal/core/{crypto, database}/: Core logic.
syncora-cli: Wrapper script.

If missing, refer to the architecture doc (docs/architecture.md) or repository.
6. Create CLI Wrapper Script
Simplify commands with a syncora-cli script:
cat > syncora-cli << 'EOF'
#!/bin/bash
cd /Users/mac/Documents/syncora
docker compose run --rm syncora "$@"
EOF
chmod +x syncora-cli

Add to PATH:
echo 'export PATH="$PATH:/Users/mac/Documents/syncora"' >> ~/.zshrc
source ~/.zshrc

7. Start Syncora
Launch the Docker services:
cd /Users/mac/Documents/syncora
docker compose up --build -d

Verify services are running:
docker compose ps

Expected output:
   Name                 Command               State            Ports          
--------------------------------------------------------------------------------
syncora_postgres_1   docker-entrypoint.sh postgres   Up      5432/tcp        
syncora_syncora_1    /app/bin/syncora help           Up                      

Using Syncora
Syncora provides commands to manage accounts and interact with bridge services. Use the syncora-cli wrapper for all operations.
1. View Help
Display available commands and usage tips:
syncora-cli help

Output (partial):
Syncora is a command-line tool to interact with multiple blockchain bridge services...

Available Commands:
  account     Manage user accounts for signing bridge transactions
  info        Retrieve information about accounts or bridge services
  help        Display help for Syncora CLI or a specific command

2. Import an Account
Add a blockchain account for signing transactions:
syncora-cli account import --alias test-wallet


Prompts:
Private key: Enter a 64-character hex string.
Passphrase: Enter a strong passphrase (12+ characters, mix letters, numbers, symbols, e.g., Y0ur$tr0ngP@ss2025!).
Confirm passphrase: Re-enter to confirm.


Output:Starting import process
Enter private key (input hidden):
Enter passphrase for encryption (input hidden):
Confirm passphrase:
Encrypting private key
Crypto: Starting EncryptPrivateKey
Crypto: Private key encrypted, address: 0xA46f88EE...
Saving account
Database: Starting SaveAccount
Database: Account saved
Account imported: alias=test-wallet, address=0x1....................................



Security Tip: Never store private keys in scripts or shell history. Use the hidden prompt.
3. Check Account Details
Verify an account’s private key:
syncora-cli info check --account test-wallet


Prompt:
Passphrase: Enter the passphrase used during import (e.g., Y0ur$tr0ngP@ss2025!).


Output:Starting account check
Database: Starting GetAccount
Database: Account retrieved
Enter passphrase to decrypt private key (input hidden):
Decrypting private key
Crypto: Starting DecryptPrivateKey
Crypto: Private key decrypted
Private key decrypted for address: 0xA46f88EE...
Account details: alias=test-wallet, address=0x........, key_version=1



4. List Accounts
View all stored accounts:
syncora-cli account list


Output:Database: Starting ListAccounts
Database: Listed accounts, count: 1
Alias       Address                              Key Version
-----       -------                              -----------
test-wallet 0x............................. 1



5. Remove an Account
Delete an account:
syncora-cli account remove --account test-wallet


Output:Database: Starting RemoveAccount
Database: Account deleted
Account removed: test-wallet



Security Best Practices

Passphrases:
Use 12+ characters with letters, numbers, and symbols (e.g., Y0ur$tr0ngP@ss2025!).
Avoid reusing passphrases across services.


Environment File:
Ensure .env has 0600 permissions (chmod 600 .env).
Never commit .env to version control.


Certificates:
Store certs/ securely and avoid sharing private keys (client.key, server.key).


Docker:
Run docker-compose down when not in use to stop containers.
Monitor logs for unauthorized access: docker-compose logs postgres.


Private Keys:
Enter private keys only in hidden prompts to avoid shell history leaks.
Backup keys securely (e.g., encrypted offline storage).



Troubleshooting
1. service "syncora" is not running

Cause: Services aren’t started.
Fix:docker-compose up -d
docker-compose ps



2. pq: column "salt" of relation "accounts" does not exist

Cause: Database schema is outdated.
Fix:
Check schema:docker-compose exec postgres psql -U syncora -d syncora_db -c "\d accounts"


Manually add columns:docker-compose exec postgres psql -U syncora -d syncora_db

ALTER TABLE accounts ADD COLUMN salt TEXT NOT NULL DEFAULT '';
ALTER TABLE accounts ADD CONSTRAINT valid_hex_salt CHECK (salt ~ '^[0-9a-fA-F]+$');
ALTER TABLE accounts ADD COLUMN key_version SMALLINT NOT NULL DEFAULT 1;
ALTER TABLE accounts ADD CONSTRAINT valid_key_version CHECK (key_version >= 1);





3. Connection reset by peer in PostgreSQL Logs

Cause: Transient client disconnection (e.g., CLI exiting).
Fix:
Monitor logs:docker-compose logs postgres


Restart services if persistent:docker-compose down
docker-compose up -d





4. invalid passphrase During Info Check

Cause: Incorrect passphrase for decryption.
Fix:
Re-enter the exact passphrase used during import.
Re-import if forgotten:syncora-cli account import --alias test-wallet





5. Certificate Errors

Cause: Missing or invalid certs/.
Fix:
Verify files:ls -l certs/{client.crt,client.key,ca.crt,server.crt,server.key,pg_hba.conf}


Regenerate if needed (see Setup step 2).



Updating Syncora
To update Syncora (e.g., after pulling new code):
cd /Users/mac/Documents/syncora
git pull
docker-compose up --build -d

Test commands to confirm functionality.
Next Steps

Explore Bridge Services: Use syncora info to check supported bridges (future feature).
Integrate API Gateway: Connect to services/api-gateway/ for REST-based operations (upcoming).
Contribute: See docs/architecture.md for developer details.


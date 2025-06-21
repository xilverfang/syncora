package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// Timeout for database operations
	dbTimeout = 5 * time.Second
)

// Account represents a stored account.
type Account struct {
	Alias        string
	Address      string
	EncryptedKey string
	Salt         string
	KeyVersion   uint8
}

// db is the global database connection.
var db *sql.DB

// init initializes the database connection and ensures the correct schema.
func init() {
	// Check .env permissions
	if err := checkEnvPermissions(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	connStr := os.Getenv("SYNCORA_DB_URL")
	if connStr == "" {
		fmt.Fprintln(os.Stderr, "Error: SYNCORA_DB_URL not set")
		os.Exit(1)
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
		os.Exit(1)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Create accounts table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			address TEXT PRIMARY KEY,
			alias TEXT NOT NULL,
			encrypted_key TEXT NOT NULL,
			CONSTRAINT valid_hex_encrypted_key CHECK (encrypted_key ~ '^[0-9a-fA-F]+$')
		)
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create accounts table: %v\n", err)
		os.Exit(1)
	}

	// Migrate schema to add salt and key_version if missing
	err = migrateSchema(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to migrate schema: %v\n", err)
		os.Exit(1)
	}

	// Enable audit logging
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgaudit`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to enable pgaudit: %v\n", err)
	}

	fmt.Fprintln(os.Stderr, "Database: Initialized successfully")
}

// migrateSchema adds missing columns (salt, key_version) to the accounts table.
func migrateSchema(ctx context.Context) error {
	// Check if salt column exists
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'accounts' AND column_name = 'salt'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check schema: %v", err)
	}

	if count == 0 {
		fmt.Fprintln(os.Stderr, "Database: Adding salt column")
		_, err = db.ExecContext(ctx, `
			ALTER TABLE accounts
			ADD COLUMN salt TEXT NOT NULL DEFAULT '',
			ADD CONSTRAINT valid_hex_salt CHECK (salt ~ '^[0-9a-fA-F]+$')
		`)
		if err != nil {
			return fmt.Errorf("failed to add salt column: %v", err)
		}
	}

	// Check if key_version column exists
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'accounts' AND column_name = 'key_version'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check schema: %v", err)
	}

	if count == 0 {
		fmt.Fprintln(os.Stderr, "Database: Adding key_version column")
		_, err = db.ExecContext(ctx, `
			ALTER TABLE accounts
			ADD COLUMN key_version SMALLINT NOT NULL DEFAULT 1,
			ADD CONSTRAINT valid_key_version CHECK (key_version >= 1)
		`)
		if err != nil {
			return fmt.Errorf("failed to add key_version column: %v", err)
		}
	}

	// Update existing rows with default salt (empty for now, requires re-import)
	_, err = db.ExecContext(ctx, `
		UPDATE accounts
		SET salt = ''
		WHERE salt IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to update existing rows: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Database: Schema migration completed")
	return nil
}

// checkEnvPermissions ensures .env file has secure permissions (0600).
func checkEnvPermissions() error {
	envPath := ".env"
	info, err := os.Stat(envPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat .env: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		return fmt.Errorf(".env permissions too open: %s, expected 0600", info.Mode().Perm())
	}
	return nil
}

// SaveAccount stores an account with its encrypted private key, salt, and key version.
func SaveAccount(alias, address, encryptedKey, salt string, keyVersion uint8) error {
	fmt.Fprintln(os.Stderr, "Database: Starting SaveAccount")
	if !common.IsHexAddress(address) {
		return fmt.Errorf("invalid address: %s", address)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		INSERT INTO accounts (address, alias, encrypted_key, salt, key_version)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (address) DO UPDATE
		SET alias = $2, encrypted_key = $3, salt = $4, key_version = $5
	`, address, alias, encryptedKey, salt, keyVersion)
	if err != nil {
		return fmt.Errorf("failed to save account: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Database: Account saved")
	return nil
}

// ListAccounts retrieves all stored accounts.
func ListAccounts() ([]Account, error) {
	fmt.Fprintln(os.Stderr, "Database: Starting ListAccounts")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, `SELECT address, alias, encrypted_key, salt, key_version FROM accounts`)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %v", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.Address, &acc.Alias, &acc.EncryptedKey, &acc.Salt, &acc.KeyVersion); err != nil {
			return nil, fmt.Errorf("failed to scan account: %v", err)
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Database: Listed accounts, count:", len(accounts))
	return accounts, nil
}

// GetAccount returns an account by alias or address.
func GetAccount(identifier string) (*Account, error) {
	fmt.Fprintln(os.Stderr, "Database: Starting GetAccount")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var acc Account
	err := db.QueryRowContext(ctx, `
		SELECT address, alias, encrypted_key, salt, key_version
		FROM accounts
		WHERE address = $1 OR alias = $1
	`, identifier).Scan(&acc.Address, &acc.Alias, &acc.EncryptedKey, &acc.Salt, &acc.KeyVersion)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found: %s", identifier)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Database: Account retrieved")
	return &acc, nil
}

// RemoveAccount deletes an account by alias or address.
func RemoveAccount(identifier string) error {
	fmt.Fprintln(os.Stderr, "Database: Starting RemoveAccount")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	result, err := db.ExecContext(ctx, `
		DELETE FROM accounts
		WHERE address = $1 OR alias = $1
	`, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete account: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("account not found: %s", identifier)
	}

	fmt.Fprintln(os.Stderr, "Database: Account deleted")
	return nil
}
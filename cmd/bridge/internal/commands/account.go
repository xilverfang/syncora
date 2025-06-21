package commands

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/xilverfang/syncora/internal/core/crypto"
	"github.com/xilverfang/syncora/internal/core/database"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func AccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage user accounts for signing bridge transactions",
		Long:  `Commands to import, list, and remove accounts, storing private keys securely for signing bridge transactions.`,
	}

	cmd.AddCommand(accountImportCmd())
	cmd.AddCommand(accountListCmd())
	cmd.AddCommand(accountRemoveCmd())

	return cmd
}

func accountImportCmd() *cobra.Command {
	var alias string
	cmd := &cobra.Command{
		Use:   "import [--alias <name>]",
		Short: "Import a private key to create or update an account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Starting import process")
			fmt.Fprint(os.Stdout, "Enter private key (input hidden): ")
			privateKeyBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to read private key: %v", err)
			}
			privateKey := strings.TrimSpace(string(privateKeyBytes))

			fmt.Fprint(os.Stdout, "Enter passphrase for encryption (input hidden): ")
			passphrase, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to read passphrase: %v", err)
			}
			if len(passphrase) < 8 {
				return fmt.Errorf("passphrase too short, minimum 8 characters")
			}

			fmt.Fprint(os.Stdout, "Confirm passphrase: ")
			passphraseConfirm, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to read passphrase confirmation: %v", err)
			}
			if string(passphrase) != string(passphraseConfirm) {
				return fmt.Errorf("passphrases do not match")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			fmt.Fprintln(os.Stderr, "Encrypting private key")
			encryptedKey, address, salt, err := crypto.EncryptPrivateKey(ctx, privateKey, passphrase, crypto.KeyVersion1)
			if err != nil {
				return fmt.Errorf("failed to encrypt private key: %v", err)
			}
			fmt.Fprintln(os.Stderr, "Private key encrypted, address:", address[:10]+"...")

			if alias == "" {
				alias = address
			}

			fmt.Fprintln(os.Stderr, "Saving account")
			if err := database.SaveAccount(alias, address, encryptedKey, hex.EncodeToString(salt), uint8(crypto.KeyVersion1)); err != nil {
				return fmt.Errorf("failed to save account: %v", err)
			}
			fmt.Fprintln(os.Stderr, "Account saved")

			// Zero sensitive data
			for i := range privateKeyBytes {
				privateKeyBytes[i] = 0
			}
			for i := range passphrase {
				passphrase[i] = 0
			}
			for i := range passphraseConfirm {
				passphraseConfirm[i] = 0
			}

			fmt.Fprintf(os.Stdout, "Account imported: alias=%s, address=%s\n", alias, address)
			return nil
		},
	}

	cmd.Flags().StringVarP(&alias, "alias", "a", "", "Optional alias for the account")
	return cmd
}

func accountListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all imported accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			accounts, err := database.ListAccounts()
			if err != nil {
				return fmt.Errorf("failed to list accounts: %v", err)
			}

			if len(accounts) == 0 {
				fmt.Println("No accounts found.")
				return nil
			}

			fmt.Println("Alias\tAddress\tKey Version")
			fmt.Println("-----\t-------\t-----------")
			for _, acc := range accounts {
				fmt.Printf("%s\t%s\t%d\n", acc.Alias, acc.Address, acc.KeyVersion)
			}
			return nil
		},
	}
	return cmd
}

func accountRemoveCmd() *cobra.Command {
	var account string
	cmd := &cobra.Command{
		Use:   "remove --account <alias-or-address>",
		Short: "Remove an account by alias or address",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stdout, "Are you sure you want to remove account %s? (y/N): ", account)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return fmt.Errorf("account removal cancelled")
			}

			if err := database.RemoveAccount(account); err != nil {
				return fmt.Errorf("failed to remove account: %v", err)
			}
			fmt.Fprintf(os.Stdout, "Account removed: %s\n", account)
			return nil
		},
	}

	cmd.Flags().StringVarP(&account, "account", "a", "", "Alias or address of the account to remove (required)")
	cmd.MarkFlagRequired("account")
	return cmd
}
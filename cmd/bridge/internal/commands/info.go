package commands

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/xilverfang/syncora/internal/core/crypto"
	"github.com/xilverfang/syncora/internal/core/database"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func InfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Retrieve information about accounts or bridge services",
		Long:  `Commands to check account details or bridge service status, using securely stored accounts.`,
	}

	cmd.AddCommand(infoCheckCmd())
	return cmd
}

func infoCheckCmd() *cobra.Command {
	var account string
	cmd := &cobra.Command{
		Use:   "check --account <alias-or-address>",
		Short: "Check account details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Starting account check")
			acc, err := database.GetAccount(account)
			if err != nil {
				return fmt.Errorf("failed to get account: %v", err)
			}

			fmt.Fprint(os.Stdout, "Enter passphrase to decrypt private key (input hidden): ")
			passphrase, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to read passphrase: %v", err)
			}

			salt, err := hex.DecodeString(acc.Salt)
			if err != nil {
				return fmt.Errorf("failed to decode salt: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			fmt.Fprintln(os.Stderr, "Decrypting private key")
			privateKeyHex, err := crypto.DecryptPrivateKey(ctx, acc.EncryptedKey, passphrase, salt, crypto.KeyVersion(acc.KeyVersion))
			if err != nil {
				return fmt.Errorf("failed to decrypt private key: %v", err)
			}

			// Convert to byte slice for zeroing
			privateKeyBytes := []byte(privateKeyHex)
			fmt.Fprintln(os.Stderr, "Private key decrypted for address:", acc.Address[:10]+"...")
			fmt.Fprintf(os.Stdout, "Account details: alias=%s, address=%s, key_version=%d\n", acc.Alias, acc.Address, acc.KeyVersion)

			// Zero sensitive data
			for i := range privateKeyBytes {
				privateKeyBytes[i] = 0
			}
			for i := range passphrase {
				passphrase[i] = 0
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&account, "account", "a", "", "Alias or address of the account to check (required)")
	cmd.MarkFlagRequired("account")
	return cmd
}
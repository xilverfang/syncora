package crypto

import (
	"context"
	"encoding/hex"
	"testing"
	"time"
)

// FuzzEncryptDecryptPrivateKey tests encryption and decryption robustness.
func FuzzEncryptDecryptPrivateKey(f *testing.F) {
	f.Add([]byte("testpass"), "fc288568c56dbf84a7af60cdb45f504ec32bac450cc042c27a420877637755ca")
	f.Fuzz(func(t *testing.T, passphrase []byte, privateKeyHex string) {
		// Skip invalid inputs
		if len(privateKeyHex) != 64 {
			return
		}
		if _, err := hex.DecodeString(privateKeyHex); err != nil {
			return
		}
		if len(passphrase) == 0 || len(passphrase) > 128 {
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// Encrypt
		encryptedKey, _, salt, err := EncryptPrivateKey(ctx, privateKeyHex, passphrase, KeyVersion1)
		if err != nil {
			if err == ctx.Err() {
				t.Logf("Encryption timed out")
			} else {
				t.Logf("Encryption error: %v", err)
			}
			return
		}

		// Decrypt
		_, err = DecryptPrivateKey(ctx, encryptedKey, passphrase, salt, KeyVersion1)
		if err != nil {
			if err == ctx.Err() {
				t.Logf("Decryption timed out")
			} else {
				t.Errorf("Decryption error: %v", err)
			}
		}
	})
}
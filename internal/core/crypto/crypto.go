package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 32 * 1024 // Reduced for fuzzing
	argon2Threads = 4
	argon2KeyLen  = 32
)

// KeyVersion represents the encryption key version.
type KeyVersion uint8

const (
	KeyVersion1 KeyVersion = 1
)

// deriveKey uses Argon2 to derive an encryption key from a passphrase, salt, and version.
func deriveKey(passphrase, salt []byte, version KeyVersion) (key, macKey []byte) {
	saltWithVersion := append(salt, byte(version))
	derived := argon2.IDKey(passphrase, saltWithVersion, argon2Time, argon2Memory, argon2Threads, argon2KeyLen*2)
	return derived[:argon2KeyLen], derived[argon2KeyLen:]
}

// EncryptPrivateKey encrypts a private key and returns the encrypted key, address, and salt.
func EncryptPrivateKey(ctx context.Context, privateKeyHex string, passphrase []byte, version KeyVersion) (string, string, []byte, error) {
	select {
	case <-ctx.Done():
		return "", "", nil, ctx.Err()
	default:
	}
	fmt.Fprintln(os.Stderr, "Crypto: Starting EncryptPrivateKey")
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	if len(privateKeyHex) != 64 {
		return "", "", nil, fmt.Errorf("invalid private key length: %d", len(privateKeyHex))
	}

	_, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", nil, err
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	key, macKey := deriveKey(passphrase, salt, version)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", "", nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	plaintext, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to decode private key: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	data := append(iv, ciphertext...)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(data)
	hmacSum := mac.Sum(nil)

	encryptedKey := hex.EncodeToString(append(data, hmacSum...))
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	fmt.Fprintln(os.Stderr, "Crypto: Private key encrypted, address:", address[:10]+"...")
	return encryptedKey, address, salt, nil
}

// DecryptPrivateKey decrypts an encrypted private key using the provided passphrase, salt, and version.
func DecryptPrivateKey(ctx context.Context, encryptedKey string, passphrase, salt []byte, version KeyVersion) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	fmt.Fprintln(os.Stderr, "Crypto: Starting DecryptPrivateKey")
	data, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted key: %v", err)
	}
	if len(data) < aes.BlockSize+sha256.Size {
		return "", fmt.Errorf("encrypted key too short")
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:len(data)-sha256.Size]
	receivedHmac := data[len(data)-sha256.Size:]

	key, macKey := deriveKey(passphrase, salt, version)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(data[:len(data)-sha256.Size])
	expectedHmac := mac.Sum(nil)
	if !hmac.Equal(receivedHmac, expectedHmac) {
		return "", fmt.Errorf("HMAC verification failed")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	privateKeyHex := hex.EncodeToString(plaintext)
	fmt.Fprintln(os.Stderr, "Crypto: Private key decrypted")
	return privateKeyHex, nil
}
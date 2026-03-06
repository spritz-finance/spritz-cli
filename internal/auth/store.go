package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/denisbrodbeck/machineid"
	"github.com/spritz-finance/spritz-cli/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "spritz"
	accountName = "api_key"
)

var ErrNotAuthenticated = errors.New("not authenticated. Run 'spritz login' to authenticate")

type StorageMethod int

const (
	StorageEnv     StorageMethod = iota
	StorageKeychain
	StorageFile
)

func GetAPIKey() (string, error) {
	if key := os.Getenv("SPRITZ_API_KEY"); key != "" {
		return key, nil
	}

	key, err := keyring.Get(serviceName, accountName)
	if err == nil && key != "" {
		return key, nil
	}

	data, err := os.ReadFile(credentialFilePath())
	if err == nil {
		key, err := decryptKey(data)
		if err == nil {
			fmt.Fprintln(os.Stderr,
				"Note: using machine-encrypted credentials file.",
				"For stronger security, set SPRITZ_API_KEY via a secrets manager.")
			return key, nil
		}
	}

	return "", ErrNotAuthenticated
}

func StoreAPIKey(apiKey string) (StorageMethod, error) {
	err := keyring.Set(serviceName, accountName, apiKey)
	if err == nil {
		return StorageKeychain, nil
	}

	encrypted, err := encryptKey(apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed to encrypt key: %w", err)
	}

	path := credentialFilePath()
	os.MkdirAll(filepath.Dir(path), 0700)
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return 0, fmt.Errorf("failed to write credentials file: %w", err)
	}
	return StorageFile, nil
}

func DeleteAPIKey() {
	keyring.Delete(serviceName, accountName)
	os.Remove(credentialFilePath())
}

func credentialFilePath() string {
	return filepath.Join(config.Dir(), "credentials")
}

func deriveKey() ([]byte, error) {
	id, err := machineid.ProtectedID("spritz")
	if err != nil {
		return nil, fmt.Errorf("failed to derive machine ID: %w", err)
	}
	h := sha256.Sum256([]byte(id + "spritz-credential-v1"))
	return h[:], nil
}

func encryptKey(apiKey string) ([]byte, error) {
	key, err := deriveKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(apiKey), nil), nil
}

func decryptKey(data []byte) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("encrypted data too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

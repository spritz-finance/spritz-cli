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

var ErrNotAuthenticated = errors.New("not authenticated. Run 'spritz auth login' or set SPRITZ_API_KEY")

type StorageMethod int

const (
	StorageEnv StorageMethod = iota
	StorageKeychain
	StorageFile
)

func (m StorageMethod) String() string {
	switch m {
	case StorageEnv:
		return "environment variable"
	case StorageKeychain:
		return "system keychain"
	case StorageFile:
		return "encrypted file"
	default:
		return "unknown"
	}
}

func GetAPIKey() (string, error) {
	key, method, err := GetAPIKeyWithSource()
	if err != nil {
		return "", err
	}
	if method == StorageFile {
		fmt.Fprintln(os.Stderr,
			"Note: using machine-encrypted credentials file.",
			"For stronger security, set SPRITZ_API_KEY via a secrets manager.")
	}
	return key, nil
}

func GetAPIKeyWithSource() (string, StorageMethod, error) {
	if key := os.Getenv("SPRITZ_API_KEY"); key != "" {
		return key, StorageEnv, nil
	}

	key, err := keyring.Get(serviceName, accountName)
	if err == nil && key != "" {
		return key, StorageKeychain, nil
	}

	data, err := os.ReadFile(credentialFilePath())
	if err == nil {
		key, err := decryptKey(data)
		if err == nil {
			return key, StorageFile, nil
		}
	}

	return "", StorageMethod(-1), ErrNotAuthenticated
}

// StoreAPIKey stores the key in the system keychain. If the keychain is
// unavailable and allowFile is true, falls back to an encrypted file.
// If allowFile is false and the keychain fails, returns an error.
func StoreAPIKey(apiKey string, allowFile bool) (StorageMethod, error) {
	err := keyring.Set(serviceName, accountName, apiKey)
	if err == nil {
		return StorageKeychain, nil
	}

	if !allowFile {
		return 0, fmt.Errorf("system keychain unavailable: %w\n\nRe-run with --allow-file-storage to use an encrypted file, "+
			"or set SPRITZ_API_KEY directly", err)
	}

	encrypted, err := encryptKey(apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed to encrypt key: %w", err)
	}

	path := credentialFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return 0, fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return 0, fmt.Errorf("failed to write credentials file: %w", err)
	}

	fmt.Fprintln(os.Stderr,
		"Warning: system keychain unavailable. Key stored in encrypted file.",
		"\nFor stronger security, set SPRITZ_API_KEY via a secrets manager.")
	return StorageFile, nil
}

func DeleteAPIKey() error {
	var errs []error
	if err := keyring.Delete(serviceName, accountName); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("keychain: %w", err))
	}
	if err := os.Remove(credentialFilePath()); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("credentials file: %w", err))
	}
	deleteKeyMetadata()
	return errors.Join(errs...)
}

// HasStoredCredentials checks for credentials in the keychain or encrypted file,
// ignoring the SPRITZ_API_KEY environment variable.
func HasStoredCredentials() bool {
	if key, err := keyring.Get(serviceName, accountName); err == nil && key != "" {
		return true
	}
	if data, err := os.ReadFile(credentialFilePath()); err == nil {
		if _, err := decryptKey(data); err == nil {
			return true
		}
	}
	return false
}

// StoreKeyMetadata persists the key ID so logout can offer server-side revocation.
func StoreKeyMetadata(keyID string) {
	if keyID == "" {
		return
	}
	path := keyMetadataFilePath()
	os.MkdirAll(filepath.Dir(path), 0700)
	os.WriteFile(path, []byte(keyID), 0600)
}

// LoadKeyMetadata returns the stored key ID, or empty string if none.
func LoadKeyMetadata() string {
	data, err := os.ReadFile(keyMetadataFilePath())
	if err != nil {
		return ""
	}
	return string(data)
}

func deleteKeyMetadata() {
	os.Remove(keyMetadataFilePath())
}

func credentialFilePath() string {
	return filepath.Join(config.Dir(), "credentials")
}

func keyMetadataFilePath() string {
	return filepath.Join(config.Dir(), "key_metadata")
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

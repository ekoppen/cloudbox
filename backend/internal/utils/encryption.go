package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Encryption parameters
	saltSize   = 32
	nonceSize  = 12
	keySize    = 32
	iterations = 10000
)

// EncryptPrivateKey encrypts a private key using AES-256-GCM with PBKDF2 key derivation
func EncryptPrivateKey(privateKey string, masterPassword string) (string, error) {
	if privateKey == "" {
		return "", errors.New("private key cannot be empty")
	}
	if masterPassword == "" {
		return "", errors.New("master password cannot be empty")
	}

	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(masterPassword), salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate random nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the private key
	ciphertext := gcm.Seal(nil, nonce, []byte(privateKey), nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, 0, saltSize+nonceSize+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptPrivateKey decrypts a private key using the master password
func DecryptPrivateKey(encryptedKey string, masterPassword string) (string, error) {
	if encryptedKey == "" {
		return "", errors.New("encrypted key cannot be empty")
	}
	if masterPassword == "" {
		return "", errors.New("master password cannot be empty")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}

	// Check minimum length
	if len(data) < saltSize+nonceSize {
		return "", errors.New("invalid encrypted data format")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(masterPassword), salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt the private key
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateMasterKey generates a secure master key for encryption
func GenerateMasterKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
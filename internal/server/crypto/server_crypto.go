package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// ServerCrypto provides server-side cryptographic operations
type ServerCrypto struct {
	masterKey []byte
	hmacKey   []byte
}

// NewServerCrypto creates a new server crypto instance
func NewServerCrypto(masterKeyPath, hmacKeyPath string) (*ServerCrypto, error) {
	// In a real implementation, these would be read from files
	// For now, we use placeholder keys
	masterKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
	hmacKey := []byte("fedcba9876543210fedcba9876543210")   // 32 bytes for HMAC

	if len(masterKey) != 32 {
		return nil, errors.New("master key must be 32 bytes")
	}
	if len(hmacKey) != 32 {
		return nil, errors.New("HMAC key must be 32 bytes")
	}

	return &ServerCrypto{
		masterKey: masterKey,
		hmacKey:   hmacKey,
	}, nil
}

// Encrypt encrypts data using AES-GCM
func (c *ServerCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// In a real implementation, we would use a proper nonce
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// Decrypt decrypts data using AES-GCM
func (c *ServerCrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Sign generates an HMAC signature
func (c *ServerCrypto) Sign(data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, c.hmacKey)
	h.Write(data)
	return h.Sum(nil), nil
}

// Verify verifies an HMAC signature
func (c *ServerCrypto) Verify(data, signature []byte) bool {
	expected, err := c.Sign(data)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, signature)
}
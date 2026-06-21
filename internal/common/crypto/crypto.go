package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// Provider defines the interface for cryptographic operations
type Provider interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	Sign(data []byte) ([]byte, error)
	Verify(data, signature []byte) bool
}

// AESGCMProvider implements encryption using AES-GCM
type AESGCMProvider struct {
	key []byte
}

// NewAESGCMProvider creates a new AES-GCM provider
func NewAESGCMProvider(key []byte) (*AESGCMProvider, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	return &AESGCMProvider{key: key}, nil
}

// Encrypt encrypts plaintext using AES-GCM
func (p *AESGCMProvider) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (p *AESGCMProvider) Decrypt(ciphertext []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(p.key)
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
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// HMACProvider implements HMAC signing
type HMACProvider struct {
	secret []byte
}

// NewHMACProvider creates a new HMAC provider
func NewHMACProvider(secret []byte) *HMACProvider {
	return &HMACProvider{secret: secret}
}

// Sign generates an HMAC signature for the data
func (p *HMACProvider) Sign(data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, p.secret)
	h.Write(data)
	return h.Sum(nil), nil
}

// Verify verifies an HMAC signature
func (p *HMACProvider) Verify(data, signature []byte) bool {
	expected, err := p.Sign(data)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, signature)
}

// GenerateRandomID generates a random identifier
func GenerateRandomID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes)[:length], nil
}
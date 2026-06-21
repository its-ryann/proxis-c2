package memory

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"unsafe"
)

// MemoryExecutor handles in-memory code execution
type MemoryExecutor struct {
	payloadRegion []byte
	regionSize    uintptr
	mu            sync.Mutex
}

// NewMemoryExecutor creates a new memory executor
func NewMemoryExecutor() *MemoryExecutor {
	return &MemoryExecutor{}
}

// AllocateMemory allocates executable memory in the heap
func (m *MemoryExecutor) AllocateMemory(size int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate random bytes for the payload region
	m.payloadRegion = make([]byte, size)
	if _, err := rand.Read(m.payloadRegion); err != nil {
		return err
	}
	m.regionSize = uintptr(size)

	return nil
}

// WritePayload writes a payload to the allocated memory region
func (m *MemoryExecutor) WritePayload(offset int, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.payloadRegion == nil {
		return errors.New("memory not allocated")
	}

	if offset+len(data) > len(m.payloadRegion) {
		return errors.New("payload exceeds allocated memory")
	}

	copy(m.payloadRegion[offset:], data)
	return nil
}

// ReadPayload reads data from the memory region
func (m *MemoryExecutor) ReadPayload(offset, size int) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.payloadRegion == nil {
		return nil, errors.New("memory not allocated")
	}

	if offset+size > len(m.payloadRegion) {
		return nil, errors.New("read exceeds allocated memory")
	}

	return m.payloadRegion[offset : offset+size], nil
}

// GetRegionSize returns the size of the allocated region
func (m *MemoryExecutor) GetRegionSize() uintptr {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.regionSize
}

// GetRegionAddress returns the address of the memory region
func (m *MemoryExecutor) GetRegionAddress() uintptr {
	m.mu.Lock()
	defer m.mu.Unlock()
	return uintptr(unsafe.Pointer(&m.payloadRegion[0]))
}

// Cleanup releases the allocated memory
func (m *MemoryExecutor) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Zero out the memory before releasing
	for i := range m.payloadRegion {
		m.payloadRegion[i] = 0
	}
	m.payloadRegion = nil
	m.regionSize = 0

	return nil
}

// EncryptPayload encrypts a payload for storage
func (m *MemoryExecutor) EncryptPayload(plaintext []byte, key []byte) ([]byte, error) {
	// Placeholder for encryption - will use the common crypto package
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

// DecryptPayload decrypts a payload for execution
func (m *MemoryExecutor) DecryptPayload(ciphertext []byte, key []byte) ([]byte, error) {
	// Placeholder for decryption - will use the common crypto package
	return base64.StdEncoding.DecodeString(string(ciphertext))
}
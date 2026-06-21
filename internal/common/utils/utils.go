package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// PlatformInfo contains information about the current platform
type PlatformInfo struct {
	OS           string
	Architecture string
	Hostname     string
}

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() (*PlatformInfo, error) {
	hostname, err := getHostname()
	if err != nil {
		return nil, err
	}

	return &PlatformInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
	}, nil
}

// getHostname returns the system hostname
func getHostname() (string, error) {
	// Platform-specific hostname retrieval
	switch runtime.GOOS {
	case "windows":
		return getWindowsHostname()
	case "linux":
		return getLinuxHostname()
	case "darwin":
		return getDarwinHostname()
	default:
		return "unknown", nil
	}
}

// getWindowsHostname returns the Windows hostname
func getWindowsHostname() (string, error) {
	// This will be implemented in platform-specific files
	return "windows-host", nil
}

// getLinuxHostname returns the Linux hostname
func getLinuxHostname() (string, error) {
	// This will be implemented in platform-specific files
	return "linux-host", nil
}

// getDarwinHostname returns the macOS hostname
func getDarwinHostname() (string, error) {
	// This will be implemented in platform-specific files
	return "darwin-host", nil
}

// CalculateJitter calculates a jittered duration
func CalculateJitter(base time.Duration, min, max float64) time.Duration {
	jitter := min + randFloat64()*(max-min)
	return time.Duration(float64(base) * (1 + jitter))
}

// randFloat64 generates a random float64 between 0 and 1
func randFloat64() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	return float64(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 |
		uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 |
		uint64(b[6])<<8 | uint64(b[7])) / float64(1<<63)
}

// ToJSON converts an interface to JSON bytes
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON converts JSON bytes to an interface
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%dh%dm%ds", h, m, s)
}
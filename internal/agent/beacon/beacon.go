package beacon

import (
	"context"
	"math/rand"
	"time"

	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
)

// Beacon handles agent beaconing logic
type Beacon struct {
	baseInterval time.Duration
	jitterMin    float64
	jitterMax    float64
	maxRetries   int
	retryBackoff time.Duration
}

// NewBeacon creates a new beacon instance
func NewBeacon(baseInterval time.Duration, jitterMin, jitterMax float64, maxRetries int, retryBackoff time.Duration) *Beacon {
	return &Beacon{
		baseInterval: baseInterval,
		jitterMin:    jitterMin,
		jitterMax:    jitterMax,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
	}
}

// CalculateNextInterval calculates the next beacon interval with jitter
func (b *Beacon) CalculateNextInterval() time.Duration {
	jitter := b.jitterMin + rand.Float64()*(b.jitterMax-b.jitterMin)
	return time.Duration(float64(b.baseInterval) * (1 + jitter))
}

// Run starts the beacon loop
func (b *Beacon) Run(ctx context.Context, agentID string, sendFunc func(*protocol.Beacon) error) {
	ticker := time.NewTicker(b.CalculateNextInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beacon := protocol.NewBeacon(agentID, nil)
			if err := sendFunc(beacon); err != nil {
				// Exponential backoff on failure
				b.exponentialBackoff()
			}
			// Reschedule with new jitter
			ticker.Reset(b.CalculateNextInterval())
		}
	}
}

// exponentialBackoff implements exponential backoff for retries
func (b *Beacon) exponentialBackoff() {
	// Placeholder for exponential backoff logic
	// Will be implemented in Phase 2
}

// SendBeacon sends a beacon to the server
func (b *Beacon) SendBeacon(ctx context.Context, client BeaconClient, agentID string) (*protocol.BeaconResponse, error) {
	beacon := protocol.NewBeacon(agentID, nil)
	return client.SendBeacon(ctx, beacon)
}

// BeaconClient is the interface for sending beacons
type BeaconClient interface {
	SendBeacon(ctx context.Context, beacon *protocol.Beacon) (*protocol.BeaconResponse, error)
}
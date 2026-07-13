package mixnode

import (
	"crypto/sha256"
	"sync"
	"time"
)

const (
	replayCacheTTL           = 30 * time.Second
	replayCacheEvictInterval = 10 * time.Second
)

type ReplayCache struct {
	mu   sync.RWMutex
	seen map[[32]byte]time.Time
}

// NewReplayCache creates a replay cache and starts a background goroutine that cleans up old entries every 10 seconds.
func NewReplayCache() *ReplayCache {
	rc := &ReplayCache{
		seen: make(map[[32]byte]time.Time),
	}
	go rc.evictLoop()
	return rc
}

// IsReplay checks if a packet has been seen before. We hash the packet with SHA-256 and look up the hash in the map. If the hash exists, it's a replay
// and we return true. Otherwise we store the hash and return false.
func (rc *ReplayCache) IsReplay(packet []byte) bool {
	hash := sha256.Sum256(packet)

	rc.mu.RLock()
	_, exists := rc.seen[hash]
	rc.mu.RUnlock()

	if exists {
		return true
	}

	rc.mu.Lock()
	rc.seen[hash] = time.Now()
	rc.mu.Unlock()

	return false
}

// evictLoop runs in the background and deletes entries older than 30 seconds. This runs every 10 seconds so the map doesn't grow forever.
func (rc *ReplayCache) evictLoop() {
	for {
		time.Sleep(replayCacheEvictInterval)
		cutoff := time.Now().Add(-replayCacheTTL)

		rc.mu.Lock()
		for hash, seenAt := range rc.seen {
			if seenAt.Before(cutoff) {
				delete(rc.seen, hash)
			}
		}
		rc.mu.Unlock()
	}
}

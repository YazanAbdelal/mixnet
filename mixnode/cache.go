package mixnode

import (
	"crypto/sha256"
	"sync"
	"time"
)

const (
	replayCacheTTL     = 30 * time.Second
	replayCacheEvictInterval = 10 * time.Second
)

type ReplayCache struct {
	mu   sync.RWMutex
	seen map[[32]byte]time.Time
}

func NewReplayCache() *ReplayCache {
	rc := &ReplayCache{
		seen: make(map[[32]byte]time.Time),
	}
	go rc.evictLoop()
	return rc
}

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

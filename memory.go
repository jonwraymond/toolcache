package toolcache

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	policy  Policy
}

func NewMemoryCache(policy Policy) *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*cacheEntry),
		policy:  policy,
	}
}

func (c *MemoryCache) Get(_ context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}

	return entry.value, true
}

func (c *MemoryCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}

	c.mu.Lock()
	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()

	return nil
}

func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	return nil
}

var _ Cache = (*MemoryCache)(nil)

package zenmanage

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// InMemoryCache is a process-local cache backend.
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

// NewInMemoryCache creates an in-memory cache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{entries: map[string]cacheEntry{}}
}

// Get retrieves a cache entry if present and unexpired.
func (c *InMemoryCache) Get(key string) (string, bool, error) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return "", false, nil
	}
	return entry.value, true, nil
}

// Set stores a cache entry.
func (c *InMemoryCache) Set(key, value string, ttl time.Duration) error {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	c.mu.Lock()
	c.entries[key] = cacheEntry{value: value, expiresAt: expiresAt}
	c.mu.Unlock()
	return nil
}

// Delete removes a cache entry.
func (c *InMemoryCache) Delete(key string) error {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	return nil
}

// Clear removes all cache entries.
func (c *InMemoryCache) Clear() error {
	c.mu.Lock()
	c.entries = map[string]cacheEntry{}
	c.mu.Unlock()
	return nil
}

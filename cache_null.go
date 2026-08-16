package zenmanage

import "time"

// NullCache is a no-op cache backend.
type NullCache struct{}

// NewNullCache creates a null cache.
func NewNullCache() *NullCache {
	return &NullCache{}
}

// Get always misses.
func (c *NullCache) Get(string) (string, bool, error) {
	return "", false, nil
}

// Set is a no-op.
func (c *NullCache) Set(string, string, time.Duration) error {
	return nil
}

// Delete is a no-op.
func (c *NullCache) Delete(string) error {
	return nil
}

// Clear is a no-op.
func (c *NullCache) Clear() error {
	return nil
}

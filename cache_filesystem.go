package zenmanage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fsPayload struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

// FileSystemCache stores cache entries on disk.
type FileSystemCache struct {
	dir string
	mu  sync.Mutex
}

// NewFileSystemCache creates a filesystem cache backend.
func NewFileSystemCache(dir string) *FileSystemCache {
	return &FileSystemCache{dir: dir}
}

func (c *FileSystemCache) filePath(key string) string {
	return filepath.Join(c.dir, key+".json")
}

// Get reads a cache entry from disk.
func (c *FileSystemCache) Get(key string) (string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b, err := os.ReadFile(c.filePath(key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	var payload fsPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return "", false, err
	}
	if !payload.ExpiresAt.IsZero() && time.Now().After(payload.ExpiresAt) {
		_ = os.Remove(c.filePath(key))
		return "", false, nil
	}
	return payload.Value, true, nil
}

// Set writes a cache entry to disk.
func (c *FileSystemCache) Set(key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	p := fsPayload{Value: value}
	if ttl > 0 {
		p.ExpiresAt = time.Now().Add(ttl)
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath(key), b, 0o644)
}

// Delete removes a cache entry.
func (c *FileSystemCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := os.Remove(c.filePath(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Clear removes all cache files in the cache directory.
func (c *FileSystemCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

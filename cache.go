package zenmanage

import "time"

// Cache is the cache backend contract.
type Cache interface {
	Get(key string) (value string, found bool, err error)
	Set(key, value string, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}

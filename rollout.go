package zenmanage

import (
	"hash/crc32"
)

// CRC32B returns CRC32 checksum compatible with PHP hash('crc32b', ...).
func CRC32B(input string) uint32 {
	return crc32.ChecksumIEEE([]byte(input))
}

// IsInBucket computes deterministic rollout inclusion for percentage rollouts.
func IsInBucket(salt, contextIdentifier string, percentage int) bool {
	if contextIdentifier == "" {
		return false
	}
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}
	hash := CRC32B(salt + ":" + contextIdentifier)
	bucket := int(hash % 100)
	return bucket < percentage
}

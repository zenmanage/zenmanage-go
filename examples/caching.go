package examples

import (
	"time"

	"github.com/zenmanage/zenmanage-go"
)

// Caching demonstrates memory and filesystem cache configuration.
func Caching(token, cacheDir string) (zenmanage.Config, zenmanage.Config, error) {
	memoryCfg, err := zenmanage.NewConfigBuilder().
		WithEnvironmentToken(token).
		WithCacheBackend("memory").
		WithCacheTTL(time.Hour).
		Build()
	if err != nil {
		return zenmanage.Config{}, zenmanage.Config{}, err
	}

	fileCfg, err := zenmanage.NewConfigBuilder().
		WithEnvironmentToken(token).
		WithCacheBackend("filesystem").
		WithCacheDirectory(cacheDir).
		WithCacheTTL(6 * time.Hour).
		Build()
	if err != nil {
		return zenmanage.Config{}, zenmanage.Config{}, err
	}

	return memoryCfg, fileCfg, nil
}

package zenmanage

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Logger is the SDK logger contract.
type Logger interface {
	Debug(message string, meta map[string]any)
	Info(message string, meta map[string]any)
	Warn(message string, meta map[string]any)
	Error(message string, meta map[string]any)
}

// NullLogger is a silent logger.
type NullLogger struct{}

func (NullLogger) Debug(string, map[string]any) {}
func (NullLogger) Info(string, map[string]any)  {}
func (NullLogger) Warn(string, map[string]any)  {}
func (NullLogger) Error(string, map[string]any) {}

// Config stores SDK configuration.
type Config struct {
	EnvironmentToken     string
	CacheTTL             time.Duration
	CacheBackend         string
	CacheDirectory       string
	EnableUsageReporting bool
	APIEndpoint          string
	Logger               Logger
	CustomCache          Cache
	HTTPClient           *http.Client
	ClientAgent          string
	SDKVersion           string
}

// ConfigBuilder builds SDK Config instances.
type ConfigBuilder struct {
	cfg Config
}

// NewConfigBuilder creates a builder with defaults.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{cfg: Config{
		CacheTTL:             time.Hour,
		CacheBackend:         "memory",
		EnableUsageReporting: true,
		APIEndpoint:          "https://api.zenmanage.com",
		Logger:               NullLogger{},
		HTTPClient:           &http.Client{Timeout: 10 * time.Second},
		ClientAgent:          "zenmanage-go",
		SDKVersion:           Version,
	}}
}

// ConfigFromEnvironment creates a builder populated from env vars.
func ConfigFromEnvironment() *ConfigBuilder {
	b := NewConfigBuilder()
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_ENVIRONMENT_TOKEN")); v != "" {
		b.WithEnvironmentToken(v)
	}
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_CACHE_TTL")); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			b.WithCacheTTL(time.Duration(i) * time.Second)
		}
	}
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_CACHE_BACKEND")); v != "" {
		b.WithCacheBackend(v)
	}
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_CACHE_DIR")); v != "" {
		b.WithCacheDirectory(v)
	}
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_ENABLE_USAGE_REPORTING")); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "on":
			b.WithUsageReporting(true)
		case "0", "false", "no", "off":
			b.WithUsageReporting(false)
		}
	}
	if v := strings.TrimSpace(os.Getenv("ZENMANAGE_API_ENDPOINT")); v != "" {
		b.WithAPIEndpoint(v)
	}
	return b
}

// WithEnvironmentToken sets the environment token.
func (b *ConfigBuilder) WithEnvironmentToken(token string) *ConfigBuilder {
	b.cfg.EnvironmentToken = strings.TrimSpace(token)
	return b
}

// WithCacheTTL sets cache TTL.
func (b *ConfigBuilder) WithCacheTTL(ttl time.Duration) *ConfigBuilder {
	b.cfg.CacheTTL = ttl
	return b
}

// WithCacheBackend sets cache backend.
func (b *ConfigBuilder) WithCacheBackend(backend string) *ConfigBuilder {
	b.cfg.CacheBackend = strings.TrimSpace(strings.ToLower(backend))
	return b
}

// WithCacheDirectory sets filesystem cache directory.
func (b *ConfigBuilder) WithCacheDirectory(dir string) *ConfigBuilder {
	b.cfg.CacheDirectory = strings.TrimSpace(dir)
	return b
}

// WithUsageReporting toggles usage reporting.
func (b *ConfigBuilder) WithUsageReporting(enabled bool) *ConfigBuilder {
	b.cfg.EnableUsageReporting = enabled
	return b
}

// WithAPIEndpoint sets custom API endpoint.
func (b *ConfigBuilder) WithAPIEndpoint(endpoint string) *ConfigBuilder {
	b.cfg.APIEndpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	return b
}

// WithLogger sets logger.
func (b *ConfigBuilder) WithLogger(logger Logger) *ConfigBuilder {
	if logger != nil {
		b.cfg.Logger = logger
	}
	return b
}

// WithCache sets custom cache implementation.
func (b *ConfigBuilder) WithCache(cache Cache) *ConfigBuilder {
	b.cfg.CustomCache = cache
	return b
}

// WithHTTPClient sets custom http client.
func (b *ConfigBuilder) WithHTTPClient(client *http.Client) *ConfigBuilder {
	if client != nil {
		b.cfg.HTTPClient = client
	}
	return b
}

// WithClientAgent sets client agent name.
func (b *ConfigBuilder) WithClientAgent(agent string) *ConfigBuilder {
	if trimmed := strings.TrimSpace(agent); trimmed != "" {
		b.cfg.ClientAgent = trimmed
	}
	return b
}

// WithSDKVersion sets sdk version string.
func (b *ConfigBuilder) WithSDKVersion(version string) *ConfigBuilder {
	if trimmed := strings.TrimSpace(version); trimmed != "" {
		b.cfg.SDKVersion = trimmed
	}
	return b
}

// Build validates and returns config.
func (b *ConfigBuilder) Build() (Config, error) {
	cfg := b.cfg
	if cfg.EnvironmentToken == "" {
		return Config{}, &ConfigurationError{Message: "environment token is required"}
	}
	if !strings.HasPrefix(cfg.EnvironmentToken, "srv_") {
		return Config{}, &ConfigurationError{Message: "Go SDK only supports server tokens prefixed with srv_"}
	}
	if cfg.CacheTTL < 0 {
		return Config{}, &ConfigurationError{Message: "cache TTL cannot be negative"}
	}
	if cfg.CacheBackend == "" {
		cfg.CacheBackend = "memory"
	}
	if cfg.Logger == nil {
		cfg.Logger = NullLogger{}
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.CustomCache == nil {
		switch cfg.CacheBackend {
		case "memory":
			cfg.CustomCache = NewInMemoryCache()
		case "filesystem":
			if cfg.CacheDirectory == "" {
				return Config{}, &ConfigurationError{Message: "cache directory is required for filesystem cache"}
			}
			cfg.CustomCache = NewFileSystemCache(cfg.CacheDirectory)
		case "null":
			cfg.CustomCache = NewNullCache()
		default:
			return Config{}, &ConfigurationError{Message: "invalid cache backend: " + cfg.CacheBackend}
		}
	}
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = "https://api.zenmanage.com"
	}
	if cfg.ClientAgent == "" {
		cfg.ClientAgent = "zenmanage-go"
	}
	if cfg.SDKVersion == "" {
		cfg.SDKVersion = Version
	}
	return cfg, nil
}

package zenmanage

import (
	"os"
	"testing"
	"time"
)

func TestConfigBuilderValidation(t *testing.T) {
	_, err := NewConfigBuilder().Build()
	if err == nil {
		t.Fatalf("expected missing token error")
	}

	_, err = NewConfigBuilder().WithEnvironmentToken("cli_abc").Build()
	if err == nil {
		t.Fatalf("expected server-token-only validation")
	}

	cfg, err := NewConfigBuilder().WithEnvironmentToken("srv_abc").Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if cfg.CacheTTL != time.Hour || cfg.CustomCache == nil {
		t.Fatalf("expected defaults")
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	t.Setenv("ZENMANAGE_ENVIRONMENT_TOKEN", "srv_env")
	t.Setenv("ZENMANAGE_CACHE_TTL", "120")
	t.Setenv("ZENMANAGE_CACHE_BACKEND", "null")
	t.Setenv("ZENMANAGE_ENABLE_USAGE_REPORTING", "false")
	t.Setenv("ZENMANAGE_API_ENDPOINT", "https://example.com")

	cfg, err := ConfigFromEnvironment().Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if cfg.CacheTTL != 120*time.Second || cfg.EnableUsageReporting != false || cfg.APIEndpoint != "https://example.com" {
		t.Fatalf("expected env-applied config")
	}
}

func TestConfigFromEnvironmentNoPanicWithoutVars(t *testing.T) {
	_ = os.Unsetenv("ZENMANAGE_ENVIRONMENT_TOKEN")
	_ = ConfigFromEnvironment()
}

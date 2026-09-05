package zenmanage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFlagManagerSingleAndDefaultsPriority(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[]}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	zm := New(cfg)
	defaults := DefaultsFromMap(map[string]any{"missing": "from-collection"})
	manager := zm.Flags().WithDefaults(defaults)

	flag, err := manager.Single(context.Background(), "missing", "inline")
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if flag.AsString() != "inline" {
		t.Fatalf("expected inline default to win")
	}

	flag, err = manager.Single(context.Background(), "missing")
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if flag.AsString() != "from-collection" {
		t.Fatalf("expected defaults collection value")
	}
}

func TestFlagManagerReportsUsageWithDefaultValue(t *testing.T) {
	usageHeaders := make(chan http.Header, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[]}`))
		case "/v1/flags/missing/usage":
			usageHeaders <- r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags()

	flag, err := manager.Single(context.Background(), "missing", true)
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if !flag.AsBool() {
		t.Fatalf("expected inline default to win")
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"missing":true}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerReportsUsageWithDefaultsCollectionValue(t *testing.T) {
	usageHeaders := make(chan http.Header, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[]}`))
		case "/v1/flags/missing/usage":
			usageHeaders <- r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	defaults := DefaultsFromMap(map[string]any{"missing": "from-collection"})
	manager := New(cfg).Flags().WithDefaults(defaults)

	flag, err := manager.Single(context.Background(), "missing")
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if flag.AsString() != "from-collection" {
		t.Fatalf("expected defaults collection value")
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"missing":"from-collection"}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerReportsUsageWithoutDefaultValueWhenFlagFound(t *testing.T) {
	usageHeaders := make(chan http.Header, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`))
		case "/v1/flags/real-flag/usage":
			usageHeaders <- r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags()

	flag, err := manager.Single(context.Background(), "real-flag")
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if !flag.AsBool() {
		t.Fatalf("expected real flag value")
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != "" {
			t.Fatalf("expected no default-value header, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerReportsUsageWithInlineDefaultWhenFlagFound(t *testing.T) {
	usageHeaders := make(chan http.Header, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`))
		case "/v1/flags/real-flag/usage":
			usageHeaders <- r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags()

	flag, err := manager.Single(context.Background(), "real-flag", false)
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if !flag.AsBool() {
		t.Fatalf("expected real flag value, not the inline default")
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"real-flag":false}` {
			t.Fatalf("expected inline default to be reported even though the flag was found, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerReportsUsageWithDefaultsCollectionValueWhenFlagFound(t *testing.T) {
	usageHeaders := make(chan http.Header, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`))
		case "/v1/flags/real-flag/usage":
			usageHeaders <- r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	defaults := DefaultsFromMap(map[string]any{"real-flag": false})
	manager := New(cfg).Flags().WithDefaults(defaults)

	flag, err := manager.Single(context.Background(), "real-flag")
	if err != nil {
		t.Fatalf("single failed: %v", err)
	}
	if !flag.AsBool() {
		t.Fatalf("expected real flag value, not the defaults collection value")
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"real-flag":false}` {
			t.Fatalf("expected defaults collection value to be reported even though the flag was found, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerAllDoesNotReportUsage(t *testing.T) {
	usageHits := make(chan string, 2)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case r.URL.Path == "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[{"version":"1","type":"boolean","key":"flag-a","name":"flag-a","target":{"value":{"value":{"boolean":true}}}},{"version":"1","type":"boolean","key":"flag-b","name":"flag-b","target":{"value":{"value":{"boolean":false}}}}]}`))
		case strings.HasPrefix(r.URL.Path, "/v1/flags/") && strings.HasSuffix(r.URL.Path, "/usage"):
			usageHits <- r.URL.Path
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags()

	flags, err := manager.All(context.Background())
	if err != nil {
		t.Fatalf("all failed: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}

	select {
	case path := <-usageHits:
		t.Fatalf("expected no usage report from All(), got request to %q", path)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestFlagManagerEvaluatesRulesAndRollout(t *testing.T) {
	cfg, _ := NewConfigBuilder().WithEnvironmentToken("srv_token").Build()
	manager := New(cfg).Flags()

	trueVal := true
	falseVal := false
	on := "on"
	fallback := "off"

	manager.rules = &RulesResponse{Version: "1", Flags: []FlagData{
		{
			Version: "1",
			Type:    FlagTypeBoolean,
			Key:     "beta",
			Name:    "Beta",
			Target: Target{Value: ValueEnvelope{Value: struct {
				Boolean *bool    `json:"boolean,omitempty"`
				String  *string  `json:"string,omitempty"`
				Number  *float64 `json:"number,omitempty"`
			}{Boolean: &falseVal}}},
			Rules: []Rule{{Clauses: []RuleCondition{{Attribute: "country", Operator: "equal", Value: "US"}}, Value: ValueEnvelope{Value: struct {
				Boolean *bool    `json:"boolean,omitempty"`
				String  *string  `json:"string,omitempty"`
				Number  *float64 `json:"number,omitempty"`
			}{Boolean: &trueVal}}}},
		},
		{
			Version: "1",
			Type:    FlagTypeString,
			Key:     "rollout",
			Name:    "Rollout",
			Target: Target{Value: ValueEnvelope{Value: struct {
				Boolean *bool    `json:"boolean,omitempty"`
				String  *string  `json:"string,omitempty"`
				Number  *float64 `json:"number,omitempty"`
			}{String: &fallback}}},
			Rollout: &RolloutData{
				Percentage: 100,
				Salt:       "salt",
				Status:     "active",
				Target: Target{Value: ValueEnvelope{Value: struct {
					Boolean *bool    `json:"boolean,omitempty"`
					String  *string  `json:"string,omitempty"`
					Number  *float64 `json:"number,omitempty"`
				}{String: &on}}},
			},
		},
	}}

	ctx := NewContext("user", "u-1", "", []Attribute{NewAttribute("country", []string{"US"})})
	fm := manager.WithContext(ctx)

	flag, err := fm.Single(context.Background(), "beta")
	if err != nil || !flag.IsEnabled() {
		t.Fatalf("expected rule evaluation match")
	}

	rolloutFlag, err := fm.Single(context.Background(), "rollout")
	if err != nil || rolloutFlag.AsString() != "on" {
		t.Fatalf("expected rollout target")
	}
}

func TestFlagManagerRefreshRulesClearsCache(t *testing.T) {
	cfg, _ := NewConfigBuilder().WithEnvironmentToken("srv_token").Build()
	manager := New(cfg).Flags()
	_ = manager.cache.Set(rulesCacheKey, `{"version":"1","flags":[]}`, time.Minute)
	manager.rules = &RulesResponse{Version: "1", Flags: []FlagData{}}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"2","flags":[]}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg2, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()
	manager.apiClient = NewAPIClient(cfg2)

	if err := manager.RefreshRules(context.Background()); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if manager.rules == nil || manager.rules.Version != "2" {
		t.Fatalf("expected refreshed rules")
	}
}

func TestFlagManagerManualReportUsage(t *testing.T) {
	received := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flags/manual-flag/usage" {
			received <- r.Header.Clone()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags().WithContext(SingleContext("user", "user-1", ""))

	if err := manager.ReportUsage(context.Background(), "manual-flag", false); err != nil {
		t.Fatalf("ReportUsage failed: %v", err)
	}

	select {
	case h := <-received:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"manual-flag":false}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
		if h.Get("X-ZEN-CONTEXT") == "" {
			t.Fatalf("expected context header to be set from manager's current context")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestFlagManagerManualReportUsagePropagatesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	manager := New(cfg).Flags()

	if err := manager.ReportUsage(context.Background(), "manual-flag", nil); err == nil {
		t.Fatal("expected error when usage reporting fails")
	}
}

package zenmanage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// startMockRulesServer starts an httptest.Server that serves rulesJSON at
// /rules.json (fronted by the usual /v1/flag-json CDN indirection). If
// onUsage is non-nil, it's invoked with the request for any
// /v1/flags/*/usage hit before responding 200.
func startMockRulesServer(t *testing.T, rulesJSON string, onUsage func(r *http.Request)) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": server.URL, "path": "/rules.json"},
			})
		case r.URL.Path == "/rules.json":
			_, _ = io.WriteString(w, rulesJSON)
		case strings.HasPrefix(r.URL.Path, "/v1/flags/") && strings.HasSuffix(r.URL.Path, "/usage"):
			if onUsage != nil {
				onUsage(r)
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestFlagManagerSingleAndDefaultsPriority(t *testing.T) {
	server := startMockRulesServer(t, `{"version":"1","flags":[]}`, nil)

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
	server := startMockRulesServer(t, `{"version":"1","flags":[]}`, func(r *http.Request) {
		usageHeaders <- r.Header.Clone()
	})

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
	server := startMockRulesServer(t, `{"version":"1","flags":[]}`, func(r *http.Request) {
		usageHeaders <- r.Header.Clone()
	})

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
	server := startMockRulesServer(t,
		`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`,
		func(r *http.Request) { usageHeaders <- r.Header.Clone() },
	)

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
	server := startMockRulesServer(t,
		`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`,
		func(r *http.Request) { usageHeaders <- r.Header.Clone() },
	)

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
	server := startMockRulesServer(t,
		`{"version":"1","flags":[{"version":"1","type":"boolean","key":"real-flag","name":"real-flag","target":{"value":{"value":{"boolean":true}}}}]}`,
		func(r *http.Request) { usageHeaders <- r.Header.Clone() },
	)

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
	server := startMockRulesServer(t,
		`{"version":"1","flags":[{"version":"1","type":"boolean","key":"flag-a","name":"flag-a","target":{"value":{"value":{"boolean":true}}}},{"version":"1","type":"boolean","key":"flag-b","name":"flag-b","target":{"value":{"value":{"boolean":false}}}}]}`,
		func(r *http.Request) { usageHits <- r.URL.Path },
	)

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

	server := startMockRulesServer(t, `{"version":"2","flags":[]}`, nil)

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

// capturingLogger records every Warn() call so tests can assert an unknown
// flag type is logged exactly once, not silently swallowed and not spammed.
type capturingLogger struct {
	NullLogger
	mu    sync.Mutex
	warns []string
}

func (l *capturingLogger) Warn(message string, _ map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warns = append(l.warns, message)
}

func (l *capturingLogger) warnCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.warns)
}

// mixedTypeRulesJSON mirrors a rules payload once the API starts serving a
// json-typed flag alongside the existing boolean/string/number types.
const mixedTypeRulesJSON = `{"version":"1","flags":[` +
	`{"version":"1","type":"boolean","key":"bool-flag","name":"bool-flag","target":{"value":{"value":{"boolean":true}}}},` +
	`{"version":"1","type":"string","key":"string-flag","name":"string-flag","target":{"value":{"value":{"string":"hello"}}}},` +
	`{"version":"1","type":"number","key":"number-flag","name":"number-flag","target":{"value":{"value":{"number":42}}}},` +
	`{"version":"1","type":"json","key":"json-flag","name":"json-flag","target":{"value":{"value":{"json":{"nested":{"a":1,"b":[1,2,3]}}}}}}` +
	`]}`

// TestFlagManagerToleratesUnknownFlagType confirms ZEN-1667: a rules payload
// containing a flag of a type this SDK release doesn't recognize (e.g. the
// upcoming "json" type) must not panic, must not error the whole payload,
// and must not silently resolve to a garbage/wrong value for that flag —
// looking it up must degrade to the caller's own default, while every other
// flag in the same payload evaluates normally.
func TestFlagManagerToleratesUnknownFlagType(t *testing.T) {
	server := startMockRulesServer(t, mixedTypeRulesJSON, nil)

	logger := &capturingLogger{}
	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		WithLogger(logger).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	manager := New(cfg).Flags()

	// The other flags in the same payload must still evaluate correctly.
	flags, err := manager.All(context.Background())
	if err != nil {
		t.Fatalf("All() should not error on a payload containing an unknown flag type: %v", err)
	}
	byKey := map[string]Flag{}
	for _, f := range flags {
		byKey[f.Key()] = f
	}
	if f, ok := byKey["bool-flag"]; !ok || !f.AsBool() {
		t.Fatalf("expected bool-flag to evaluate to true, got %+v (ok=%v)", f, ok)
	}
	if f, ok := byKey["string-flag"]; !ok || f.AsString() != "hello" {
		t.Fatalf("expected string-flag to evaluate to %q, got %+v (ok=%v)", "hello", f, ok)
	}
	if f, ok := byKey["number-flag"]; !ok || f.AsNumber() != 42 {
		t.Fatalf("expected number-flag to evaluate to 42, got %+v (ok=%v)", f, ok)
	}

	// Looking up the unknown-typed flag directly must resolve to the
	// caller's own default, not panic, not error, and not a garbage value
	// (e.g. an empty string coerced from a mis-parsed value wrapper).
	flag, err := manager.Single(context.Background(), "json-flag", "caller-default")
	if err != nil {
		t.Fatalf("Single() should not error on an unknown flag type, got: %v", err)
	}
	if flag.AsString() != "caller-default" {
		t.Fatalf("expected unknown-typed flag to resolve to the caller's default %q, got %q", "caller-default", flag.AsString())
	}

	numFlag, err := manager.Single(context.Background(), "json-flag", 99.5)
	if err != nil {
		t.Fatalf("Single() should not error on an unknown flag type, got: %v", err)
	}
	if numFlag.AsNumber() != 99.5 {
		t.Fatalf("expected unknown-typed flag to resolve to the caller's numeric default %v, got %v", 99.5, numFlag.AsNumber())
	}

	// A single unknown flag type in the payload should produce exactly one
	// warning log, not one per lookup/evaluation.
	if got := logger.warnCount(); got != 1 {
		t.Fatalf("expected exactly one warning logged for the unknown flag type, got %d", got)
	}
}

// TestFlagManagerUnknownFlagTypeUsesDefaultsCollection confirms Single()
// falls back to a DefaultsCollection entry (not just an inline default) when
// the flag it finds has a type this SDK release doesn't recognize.
func TestFlagManagerUnknownFlagTypeUsesDefaultsCollection(t *testing.T) {
	server := startMockRulesServer(t, mixedTypeRulesJSON, nil)

	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	defaults := DefaultsFromMap(map[string]any{"json-flag": "from-collection"})
	manager := New(cfg).Flags().WithDefaults(defaults)

	flag, err := manager.Single(context.Background(), "json-flag")
	if err != nil {
		t.Fatalf("Single() should not error on an unknown flag type, got: %v", err)
	}
	if flag.AsString() != "from-collection" {
		t.Fatalf("expected unknown-typed flag to fall back to the DefaultsCollection value %q, got %q", "from-collection", flag.AsString())
	}
}

// TestFlagManagerUnknownFlagTypeWarningDedupesAcrossClones confirms the
// warned-type dedup state is shared across managers cloned via
// WithContext/WithDefaults, since the Zenmanage convenience methods
// (IsEnabled/GetString/GetNumber) clone a fresh FlagManager on every call —
// without sharing that state, every request would re-log the same warning.
func TestFlagManagerUnknownFlagTypeWarningDedupesAcrossClones(t *testing.T) {
	server := startMockRulesServer(t, mixedTypeRulesJSON, nil)

	logger := &capturingLogger{}
	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		WithLogger(logger).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	root := New(cfg).Flags()

	// Simulate repeated per-request usage: each lookup clones a fresh
	// FlagManager via WithContext, the way IsEnabled/GetString/GetNumber do.
	for i := 0; i < 3; i++ {
		clone := root.WithContext(SingleContext("user", "u-1", ""))
		if _, err := clone.Single(context.Background(), "json-flag", "default"); err != nil {
			t.Fatalf("Single() should not error on an unknown flag type, got: %v", err)
		}
	}

	if got := logger.warnCount(); got != 1 {
		t.Fatalf("expected exactly one warning logged across clones sharing the same root manager, got %d", got)
	}
}

func TestFlagManagerManualReportUsagePropagatesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

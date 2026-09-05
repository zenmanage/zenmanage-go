package zenmanage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Errors ---

func TestErrorMessages(t *testing.T) {
	if (&ConfigurationError{Message: "bad config"}).Error() != "bad config" {
		t.Fatalf("ConfigurationError message wrong")
	}
	if (&EvaluationError{Message: "eval fail"}).Error() != "eval fail" {
		t.Fatalf("EvaluationError message wrong")
	}
	e := &FetchRulesError{Message: "net error", StatusCode: 503}
	if e.Error() == "" {
		t.Fatalf("FetchRulesError message empty")
	}
	e2 := &FetchRulesError{Message: "no status"}
	if e2.Error() != "no status" {
		t.Fatalf("FetchRulesError no-status message wrong")
	}
	if (&InvalidRulesError{Message: "bad json"}).Error() != "bad json" {
		t.Fatalf("InvalidRulesError message wrong")
	}
}

func TestErrorInterfaceSatisfiedByAllErrorTypes(t *testing.T) {
	var errs []Error
	errs = append(errs, &ConfigurationError{Message: "x"})
	errs = append(errs, &EvaluationError{Message: "x"})
	errs = append(errs, &FetchRulesError{Message: "x"})
	errs = append(errs, &InvalidRulesError{Message: "x"})

	for _, e := range errs {
		if e.Error() == "" {
			t.Fatalf("expected non-empty message for %T", e)
		}
	}

	var target Error
	var err error = &ConfigurationError{Message: "bad config"}
	if !errors.As(err, &target) {
		t.Fatalf("expected errors.As to match the shared Error interface")
	}
}

// --- NullCache ---

func TestNullCacheAlwaysMisses(t *testing.T) {
	c := NewNullCache()
	if err := c.Set("k", "v", time.Minute); err != nil {
		t.Fatalf("set error: %v", err)
	}
	v, ok, err := c.Get("k")
	if err != nil || ok || v != "" {
		t.Fatalf("expected null cache miss")
	}
	if err := c.Delete("k"); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("clear error: %v", err)
	}
}

// --- InMemoryCache.Clear ---

func TestInMemoryCacheClears(t *testing.T) {
	c := NewInMemoryCache()
	_ = c.Set("a", "1", time.Minute)
	_ = c.Set("b", "2", time.Minute)
	if err := c.Clear(); err != nil {
		t.Fatalf("clear error: %v", err)
	}
	_, ok, _ := c.Get("a")
	if ok {
		t.Fatalf("expected cleared cache")
	}
}

// --- Attribute helpers ---

func TestAttributeHelpers(t *testing.T) {
	a := NewAttribute("plan", []string{"pro"})
	if a.Key() != "plan" {
		t.Fatalf("expected key plan")
	}
	vals := a.Values()
	if len(vals) != 1 || vals[0] != "pro" {
		t.Fatalf("expected values")
	}
	a.AddValue("enterprise")
	if len(a.Values()) != 2 {
		t.Fatalf("expected added value")
	}
}

// --- Context helpers ---

func TestContextHelpers(t *testing.T) {
	ctx := ContextFromData(ContextData{Type: "org", Identifier: "org-1"})
	if ctx.Type() != "org" || ctx.Identifier() != "org-1" {
		t.Fatalf("expected context from data")
	}
	d := ctx.Data()
	if d.Type != "org" {
		t.Fatalf("expected data copy")
	}
	attrs := ctx.Attributes()
	if attrs == nil {
		t.Fatalf("expected non-nil attrs slice")
	}
}

// --- DefaultsCollection edge cases ---

func TestDefaultsCollectionExtraEdges(t *testing.T) {
	c := NewDefaultsCollection()
	c.Clear()
	if c.Size() != 0 {
		t.Fatalf("expected empty after clear")
	}

	c.Set("x", "y")
	keys := c.Keys()
	if len(keys) != 1 || keys[0] != "x" {
		t.Fatalf("expected keys slice")
	}

	// nil-safe operations
	var nilCol *DefaultsCollection
	_, ok := nilCol.Get("k")
	if ok {
		t.Fatalf("nil collection get should miss")
	}
	if nilCol.Has("k") {
		t.Fatalf("nil collection has should return false")
	}
	nilCol.Delete("k")
	nilCol.Clear()
	if nilCol.Size() != 0 {
		t.Fatalf("nil collection size should be 0")
	}
	if nilCol.Keys() != nil {
		t.Fatalf("nil collection keys should be nil")
	}
	if len(nilCol.All()) != 0 {
		t.Fatalf("nil collection all should be empty map")
	}
}

// --- Config builder chains ---

func TestConfigBuilderChainedOptions(t *testing.T) {
	cache := NewInMemoryCache()
	logger := NullLogger{}
	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithCacheTTL(30 * time.Second).
		WithCacheBackend("memory").
		WithCacheDirectory("/tmp").
		WithUsageReporting(false).
		WithAPIEndpoint("https://example.com/").
		WithLogger(logger).
		WithCache(cache).
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}).
		WithClientAgent("myapp").
		WithSDKVersion("1.2.3").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if cfg.ClientAgent != "myapp" || cfg.SDKVersion != "1.2.3" {
		t.Fatalf("expected chain values")
	}
	if cfg.APIEndpoint != "https://example.com" {
		t.Fatalf("expected trailing slash stripped")
	}
	if cfg.EnableUsageReporting {
		t.Fatalf("expected usage reporting off")
	}
}

func TestConfigBuilderFilesystemMissingDir(t *testing.T) {
	_, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithCacheBackend("filesystem").
		Build()
	if err == nil {
		t.Fatalf("expected error for missing cache dir")
	}
}

func TestConfigBuilderInvalidBackend(t *testing.T) {
	_, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithCacheBackend("redis").
		Build()
	if err == nil {
		t.Fatalf("expected error for invalid backend")
	}
}

// --- Flag accessor methods ---

func TestFlagAllAccessors(t *testing.T) {
	b := false
	f := newFlag(FlagData{Version: "1", Type: FlagTypeBoolean, Key: "k", Name: "N"}, Target{}, nil)
	if f.Version() != "1" || f.Key() != "k" || f.Name() != "N" {
		t.Fatalf("expected flag metadata")
	}
	_ = f.Target()
	_ = f.Rules()
	_ = f.Rollout()

	// disabled bool
	f2 := Flag{typ: FlagTypeBoolean, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{Boolean: &b}}}}
	if f2.IsEnabled() {
		t.Fatalf("expected false flag disabled")
	}
	if f2.AsBool() {
		t.Fatalf("expected false as bool")
	}
	if f2.AsNumber() != 0 {
		t.Fatalf("expected 0 number for false")
	}
}

func TestFlagNilValues(t *testing.T) {
	// flags with nil value pointers should return zero values
	fb := Flag{typ: FlagTypeBoolean}
	if fb.AsBool() || fb.AsNumber() != 0 || fb.AsString() != "false" {
		t.Fatalf("expected false defaults")
	}
	fn := Flag{typ: FlagTypeNumber}
	if fn.AsNumber() != 0 {
		t.Fatalf("expected 0 for nil number")
	}
	fs := Flag{typ: FlagTypeString}
	if fs.AsString() != "" {
		t.Fatalf("expected empty string for nil string")
	}
}

func TestDefaultFlagFloat32AndInt64(t *testing.T) {
	if f := newDefaultFlag("a", float32(1.5)); f.Type() != FlagTypeNumber {
		t.Fatalf("expected number type for float32")
	}
	if f := newDefaultFlag("a", int64(10)); f.Type() != FlagTypeNumber {
		t.Fatalf("expected number type for int64")
	}
	if f := newDefaultFlag("a", float64(3.14)); f.Type() != FlagTypeNumber {
		t.Fatalf("expected number type for float64")
	}
	// unknown type should default to string
	if f := newDefaultFlag("a", []int{1, 2}); f.Type() != FlagTypeString {
		t.Fatalf("expected string for unknown type")
	}
}

func TestFlagAsNumberFromStringError(t *testing.T) {
	s := "not-a-number"
	f := Flag{typ: FlagTypeString, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{String: &s}}}}
	if f.AsNumber() != 0 {
		t.Fatalf("expected 0 for non-numeric string")
	}
}

// --- Rule engine operator coverage ---

func TestRuleEngineAllOperators(t *testing.T) {
	engine := NewRuleEngine()
	attr := NewAttribute("score", []string{"50"})
	ctx := NewContext("user", "u-1", "", []Attribute{attr})

	cases := []struct {
		op    string
		value any
		want  bool
	}{
		{"notequal", "10", true},
		{"in", []any{"50", "60"}, true},
		{"notin", []any{"10", "20"}, true},
		{"contains", "5", true},
		{"notcontains", "99", true},
		{"startswith", "5", true},
		{"notstartswith", "9", true},
		{"endswith", "0", true},
		{"notendswith", "9", true},
		{"gt", "40", true},
		{"gte", "50", true},
		{"lt", "60", true},
		{"lte", "50", true},
		{"isnull", nil, false},
		{"notnull", nil, true},
	}

	for _, tc := range cases {
		ok, err := engine.matchesCondition(RuleCondition{Attribute: "score", Operator: tc.op, Value: tc.value}, ctx)
		if err != nil {
			t.Fatalf("operator %s error: %v", tc.op, err)
		}
		if ok != tc.want {
			t.Fatalf("operator %s: expected %v got %v", tc.op, tc.want, ok)
		}
	}
}

func TestRuleEngineMissingAttributeFails(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", nil)
	ok, err := engine.matchesCondition(RuleCondition{Attribute: "missing", Operator: "equal", Value: "x"}, ctx)
	if err != nil || ok {
		t.Fatalf("missing attribute should return false, got ok=%v err=%v", ok, err)
	}
}

func TestRuleEngineContextTargetMissingIdentifier(t *testing.T) {
	engine := NewRuleEngine()
	ctx := SingleContext("user", "", "")
	ok, err := engine.matchesCondition(RuleCondition{Attribute: "context", Operator: "in", Value: []any{map[string]any{"identifier": "u-1"}}}, ctx)
	if err != nil || ok {
		t.Fatalf("expected false for empty identifier")
	}
}

func TestRuleEngineNoClauses(t *testing.T) {
	engine := NewRuleEngine()
	ctx := SingleContext("user", "u-1", "")
	s := "val"
	rule := Rule{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{String: &s}}}
	v, err := engine.Evaluate([]Rule{rule}, ctx)
	if err != nil || v == nil || v.Value.String == nil || *v.Value.String != "val" {
		t.Fatalf("expected unconditional rule match")
	}
}

func TestRuleEngineNoMatch(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", []Attribute{NewAttribute("country", []string{"CA"})})
	rule := Rule{Clauses: []RuleCondition{{Attribute: "country", Operator: "equal", Value: "US"}}}
	v, err := engine.Evaluate([]Rule{rule}, ctx)
	if err != nil || v != nil {
		t.Fatalf("expected no match")
	}
}

// --- FlagManager.All ---

func TestFlagManagerAll(t *testing.T) {
	cfg, _ := NewConfigBuilder().WithEnvironmentToken("srv_token").Build()
	manager := New(cfg).Flags()

	b := true
	manager.rules = &RulesResponse{Version: "1", Flags: []FlagData{
		{Version: "1", Type: FlagTypeBoolean, Key: "a", Name: "A", Target: Target{Value: ValueEnvelope{Value: struct {
			Boolean *bool    `json:"boolean,omitempty"`
			String  *string  `json:"string,omitempty"`
			Number  *float64 `json:"number,omitempty"`
		}{Boolean: &b}}}},
	}}

	flags, err := manager.All(context.Background())
	if err != nil || len(flags) != 1 || flags[0].Key() != "a" {
		t.Fatalf("expected all flags, got %v %v", len(flags), err)
	}
}

// --- FetchRules invalid response ---

func TestAPIClientInvalidRulesResponse(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_, _ = w.Write([]byte(`{"data":{"cdn":"` + server.URL + `","path":"/rules.json"}}`))
		case "/rules.json":
			_, _ = w.Write([]byte(`not json`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	_, err := client.FetchRules(context.Background())
	if err == nil {
		t.Fatalf("expected invalid rules error")
	}
}

func TestAPIClientMissingCDNPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	_, err := client.FetchRules(context.Background())
	if err == nil {
		t.Fatalf("expected error for missing cdn/path")
	}
}

func TestAPIClientNonJSON404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{bad`))
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	_, err := client.FetchRules(context.Background())
	if err == nil {
		t.Fatalf("expected error for bad json")
	}
}

func TestReportUsageDisabled(t *testing.T) {
	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithUsageReporting(false).
		Build()
	client := NewAPIClient(cfg)
	err := client.ReportUsage(context.Background(), "flag-key", nil, nil)
	if err != nil {
		t.Fatalf("expected no error when usage reporting disabled")
	}
}

// --- FlagManager cache hit ---

func TestFlagManagerLoadRulesCacheHit(t *testing.T) {
	cfg, _ := NewConfigBuilder().WithEnvironmentToken("srv_token").Build()
	manager := New(cfg).Flags()

	_ = manager.cache.Set(rulesCacheKey, `{"version":"1","flags":[]}`, time.Minute)

	flags, err := manager.All(context.Background())
	if err != nil || len(flags) != 0 {
		t.Fatalf("expected cached rules load, got err=%v", err)
	}
}

// --- FlagManager Single with missing flag and no defaults ---

func TestFlagManagerSingleMissingNoDefault(t *testing.T) {
	cfg, _ := NewConfigBuilder().WithEnvironmentToken("srv_token").Build()
	manager := New(cfg).Flags()
	manager.rules = &RulesResponse{Version: "1", Flags: []FlagData{}}

	_, err := manager.Single(context.Background(), "ghost")
	if err == nil {
		t.Fatalf("expected evaluation error for missing flag")
	}
}

// --- asStringSlice ---

func TestAsStringSliceVariants(t *testing.T) {
	cases := []struct {
		in  any
		out []string
	}{
		{[]string{"a", "b"}, []string{"a", "b"}},
		{[]any{"x", "y"}, []string{"x", "y"}},
		{"single", []string{"single"}},
		{nil, nil},
	}
	for _, tc := range cases {
		got := asStringSlice(tc.in)
		if tc.out == nil {
			if got != nil {
				t.Fatalf("expected nil for %v", tc.in)
			}
			continue
		}
		if len(got) != len(tc.out) {
			t.Fatalf("expected %v got %v", tc.out, got)
		}
	}
}

func TestRuleEngineAbsentAttributeBehavior(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", nil)

	cases := []struct {
		op   string
		want bool
	}{
		{"isnull", true},
		{"notnull", false},
		{"notequal", true},
		{"notin", true},
		{"notcontains", true},
		{"notstartswith", true},
		{"notendswith", true},
		{"equal", false},
		{"in", false},
		{"contains", false},
		{"startswith", false},
		{"endswith", false},
	}

	for _, tc := range cases {
		ok, err := engine.matchesCondition(RuleCondition{Attribute: "missing", Operator: tc.op, Value: "x"}, ctx)
		if err != nil {
			t.Fatalf("absent attr operator %s error: %v", tc.op, err)
		}
		if ok != tc.want {
			t.Fatalf("absent attr operator %s: expected %v got %v", tc.op, tc.want, ok)
		}
	}
}

func TestRuleEngineIsNullNotnullWithEmptyValue(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", []Attribute{NewAttribute("tag", []string{""})})

	ok, err := engine.matchesCondition(RuleCondition{Attribute: "tag", Operator: "isnull", Value: nil}, ctx)
	if err != nil || !ok {
		t.Fatalf("isnull with empty value: expected true, got ok=%v err=%v", ok, err)
	}

	ok, err = engine.matchesCondition(RuleCondition{Attribute: "tag", Operator: "notnull", Value: nil}, ctx)
	if err != nil || ok {
		t.Fatalf("notnull with empty value: expected false, got ok=%v err=%v", ok, err)
	}
}

func TestAsStringVariants(t *testing.T) {
	cases := []struct {
		in  any
		out string
	}{
		{nil, ""},
		{true, "true"},
		{false, "false"},
		{float64(3.14), "3.14"},
		{float32(1.5), "1.5"},
		{int(5), "5"},
		{int64(100), "100"},
	}
	for _, tc := range cases {
		got := asString(tc.in)
		if got != tc.out {
			t.Fatalf("asString(%v): expected %q got %q", tc.in, tc.out, got)
		}
	}
}

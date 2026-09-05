package zenmanage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const convenienceRulesJSON = `{"version":"1","flags":[` +
	`{"version":"1","type":"boolean","key":"enabled-flag","name":"Enabled","target":{"value":{"value":{"boolean":true}}},"rules":[]},` +
	`{"version":"1","type":"boolean","key":"disabled-flag","name":"Disabled","target":{"value":{"value":{"boolean":false}}},"rules":[]},` +
	`{"version":"1","type":"string","key":"string-flag","name":"Str","target":{"value":{"value":{"string":"variant-b"}}},"rules":[]},` +
	`{"version":"1","type":"number","key":"number-flag","name":"Num","target":{"value":{"value":{"number":42.5}}},"rules":[]}` +
	`]}`

func newConvenienceClient(t *testing.T) *Zenmanage {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": srv.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			_, _ = w.Write([]byte(convenienceRulesJSON))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)

	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(srv.URL).
		WithHTTPClient(srv.Client()).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return New(cfg)
}

func TestIsEnabledConvenience(t *testing.T) {
	client := newConvenienceClient(t)

	ok, err := client.IsEnabled(context.Background(), "enabled-flag", "user-1")
	if err != nil || !ok {
		t.Fatalf("expected enabled: err=%v ok=%v", err, ok)
	}

	ok, err = client.IsEnabled(context.Background(), "disabled-flag", "user-1")
	if err != nil || ok {
		t.Fatalf("expected disabled: err=%v ok=%v", err, ok)
	}
}

func TestIsEnabledNoUserID(t *testing.T) {
	client := newConvenienceClient(t)
	ok, err := client.IsEnabled(context.Background(), "enabled-flag", "")
	if err != nil || !ok {
		t.Fatalf("expected enabled with empty userID: err=%v ok=%v", err, ok)
	}
}

func TestIsEnabledMissingFlag(t *testing.T) {
	client := newConvenienceClient(t)
	// missing flag uses inline default false, no error
	ok, err := client.IsEnabled(context.Background(), "does-not-exist", "u-1")
	if err != nil || ok {
		t.Fatalf("expected false default for missing flag: err=%v ok=%v", err, ok)
	}
}

func TestGetStringConvenience(t *testing.T) {
	client := newConvenienceClient(t)

	v, err := client.GetString(context.Background(), "string-flag", "user-1", "fallback")
	if err != nil || v != "variant-b" {
		t.Fatalf("expected variant-b: err=%v v=%q", err, v)
	}

	v, err = client.GetString(context.Background(), "missing-flag", "user-1", "fallback")
	if err != nil || v != "fallback" {
		t.Fatalf("expected fallback: err=%v v=%q", err, v)
	}
}

func TestGetNumberConvenience(t *testing.T) {
	client := newConvenienceClient(t)

	n, err := client.GetNumber(context.Background(), "number-flag", "user-1", 0)
	if err != nil || n != 42.5 {
		t.Fatalf("expected 42.5: err=%v n=%v", err, n)
	}

	n, err = client.GetNumber(context.Background(), "missing-flag", "user-1", 99.9)
	if err != nil || n != 99.9 {
		t.Fatalf("expected default 99.9: err=%v n=%v", err, n)
	}
}

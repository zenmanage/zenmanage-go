package zenmanage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAPIClientFetchRulesAndReportUsage(t *testing.T) {
	var usageCount atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			_, _ = w.Write([]byte(`{"data":{"cdn":"` + serverURL(t, r) + `","path":"/rules.json"}}`))
		case "/rules.json":
			_, _ = w.Write([]byte(`{"version":"1","flags":[]}`))
		case "/v1/flags/test-flag/usage":
			usageCount.Add(1)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	client := NewAPIClient(cfg)
	if _, err := client.FetchRules(context.Background()); err != nil {
		t.Fatalf("fetch rules failed: %v", err)
	}
	ctx := SingleContext("user", "u-1", "")
	if err := client.ReportUsage(context.Background(), "test-flag", &ctx, nil); err != nil {
		t.Fatalf("report usage failed: %v", err)
	}
	if usageCount.Load() != 1 {
		t.Fatalf("expected one usage report")
	}
}

func TestAPIClientReportUsageSendsDefaultValueHeader(t *testing.T) {
	usageHeaders := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flags/new-ui/usage" {
			usageHeaders <- r.Header.Clone()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, err := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	client := NewAPIClient(cfg)
	if err := client.ReportUsage(context.Background(), "new-ui", nil, true); err != nil {
		t.Fatalf("report usage failed: %v", err)
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"new-ui":true}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestAPIClientReportUsageSendsNonBooleanDefaultValueHeader(t *testing.T) {
	usageHeaders := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flags/num-flag/usage" {
			usageHeaders <- r.Header.Clone()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	if err := client.ReportUsage(context.Background(), "num-flag", nil, 42); err != nil {
		t.Fatalf("report usage failed: %v", err)
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"num-flag":42}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestAPIClientReportUsageOmitsDefaultValueHeaderWhenNotProvided(t *testing.T) {
	usageHeaders := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flags/new-ui/usage" {
			usageHeaders <- r.Header.Clone()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	if err := client.ReportUsage(context.Background(), "new-ui", nil, nil); err != nil {
		t.Fatalf("report usage failed: %v", err)
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

func TestAPIClientReportUsageSendsBothContextAndDefaultValueHeaders(t *testing.T) {
	usageHeaders := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flags/num-flag/usage" {
			usageHeaders <- r.Header.Clone()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	ctx := SingleContext("user", "u-1", "")
	if err := client.ReportUsage(context.Background(), "num-flag", &ctx, 42); err != nil {
		t.Fatalf("report usage failed: %v", err)
	}

	select {
	case h := <-usageHeaders:
		if got := h.Get("X-ZEN-DEFAULT-VALUE"); got != `{"num-flag":42}` {
			t.Fatalf("expected default-value header, got %q", got)
		}
		if h.Get("X-ZEN-CONTEXT") == "" {
			t.Fatalf("expected context header to be present")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for usage report")
	}
}

func TestAPIClientRetriesServerErrors(t *testing.T) {
	var attempts atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/flag-json" {
			if attempts.Add(1) < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"data":{"cdn":"` + server.URL + `","path":"/rules.json"}}`))
			return
		}
		if r.URL.Path == "/rules.json" {
			_, _ = w.Write([]byte(`{"version":"1","flags":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg, _ := NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(server.URL).
		WithHTTPClient(server.Client()).
		Build()

	client := NewAPIClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := client.FetchRules(ctx)
	if err != nil {
		t.Fatalf("expected retries to eventually succeed: %v", err)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected three attempts, got %d", attempts.Load())
	}
}

func serverURL(t *testing.T, r *http.Request) string {
	t.Helper()
	if r.TLS != nil {
		return fmt.Sprintf("https://%s", r.Host)
	}
	return fmt.Sprintf("http://%s", r.Host)
}

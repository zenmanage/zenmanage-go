// Package testutil provides test-only helpers shared by the zenmanage-go SDK
// and its middleware submodules (net/http, gin, echo). It is not part of the
// public API.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	zenmanage "github.com/zenmanage/zenmanage-go"
)

const rulesJSON = `{"version":"1","flags":[` +
	`{"version":"1","type":"boolean","key":"feat","name":"Feature","target":{"value":{"value":{"boolean":true}}},"rules":[]},` +
	`{"version":"1","type":"string","key":"color","name":"Color","target":{"value":{"value":{"string":"hello"}}},"rules":[]},` +
	`{"version":"1","type":"json","key":"config","name":"Config","target":{"value":{"value":{"json":{"mode":"dark"}}}},"rules":[]}` +
	`]}`

// BuildPreloadedClient sets up a mock HTTP server that returns three flags
// and returns a Zenmanage client pointed at that server.
func BuildPreloadedClient(t *testing.T) *zenmanage.Zenmanage {
	t.Helper()

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]string{"cdn": srv.URL, "path": "/rules.json"},
			})
		case "/rules.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(rulesJSON))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)

	cfg, err := zenmanage.NewConfigBuilder().
		WithEnvironmentToken("srv_token").
		WithAPIEndpoint(srv.URL).
		WithHTTPClient(srv.Client()).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return zenmanage.New(cfg)
}

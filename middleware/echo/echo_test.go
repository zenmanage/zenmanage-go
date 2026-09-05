package echomiddleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	zenmanage "github.com/zenmanage/zenmanage-go"
	echomiddleware "github.com/zenmanage/zenmanage-go/middleware/echo"
)

// buildPreloadedClient sets up a mock HTTP server that returns two flags and
// returns a Zenmanage client pointed at that server.
func buildPreloadedClient(t *testing.T) *zenmanage.Zenmanage {
	t.Helper()

	const rulesJSON = `{"version":"1","flags":[` +
		`{"version":"1","type":"boolean","key":"feat","name":"Feature","target":{"value":{"value":{"boolean":true}}},"rules":[]},` +
		`{"version":"1","type":"string","key":"color","name":"Color","target":{"value":{"value":{"string":"hello"}}},"rules":[]}` +
		`]}`

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/flag-json":
			w.Header().Set("Content-Type", "application/json")
			metadataJSON := fmt.Sprintf(`{"data":{"cdn":"%s","path":"/rules.json"}}`, srv.URL)
			_, _ = w.Write([]byte(metadataJSON))
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

func newTestContext(e *echo.Echo, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestInjectFlags_NoUserID(t *testing.T) {
	client := buildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		gotFM = echomiddleware.FlagManagerFromContext(c)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager in context")
	}
}

func TestInjectFlags_WithUserID(t *testing.T) {
	client := buildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		gotFM = echomiddleware.FlagManagerFromContext(c)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user-42")
	e.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager with user context")
	}
}

func TestIsEnabled_ViaEchoContext(t *testing.T) {
	client := buildPreloadedClient(t)

	var enabled bool
	var evalErr error
	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		enabled, evalErr = echomiddleware.IsEnabled(c, "feat")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if !enabled {
		t.Fatal("expected flag to be enabled")
	}
}

func TestIsEnabled_NoManager(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(e, req)

	_, err := echomiddleware.IsEnabled(c, "feat")
	if err == nil {
		t.Fatal("expected error when no manager in context")
	}
}

func TestGetString_ViaEchoContext(t *testing.T) {
	client := buildPreloadedClient(t)

	var val string
	var evalErr error
	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		val, evalErr = echomiddleware.GetString(c, "color", "default")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}
}

func TestGetString_NoManager(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(e, req)

	v, err := echomiddleware.GetString(c, "color", "fallback")
	if err == nil {
		t.Fatal("expected error")
	}
	if v != "fallback" {
		t.Fatalf("expected fallback, got %q", v)
	}
}

func TestGetNumber_ViaEchoContext(t *testing.T) {
	client := buildPreloadedClient(t)

	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		n, err := echomiddleware.GetNumber(c, "missing-num", 3.14)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 3.14 {
			t.Errorf("expected default 3.14, got %v", n)
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetNumber_NoManager(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(e, req)

	v, err := echomiddleware.GetNumber(c, "n", 9.9)
	if err == nil {
		t.Fatal("expected error")
	}
	if v != 9.9 {
		t.Fatalf("expected 9.9, got %v", v)
	}
}

func TestFlagManagerFromContext_Nil(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(e, req)

	fm := echomiddleware.FlagManagerFromContext(c)
	if fm != nil {
		t.Fatal("expected nil when no manager injected")
	}
}

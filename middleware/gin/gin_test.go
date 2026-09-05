package ginmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	zenmanage "github.com/zenmanage/zenmanage-go"
	ginmiddleware "github.com/zenmanage/zenmanage-go/middleware/gin"
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
			_, _ = w.Write([]byte(`{"data":{"cdn":"` + srv.URL + `","path":"/rules.json"}}`))
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

func newTestContext(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	return c, rec
}

func TestInjectFlags_NoUserID(t *testing.T) {
	client := buildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		gotFM = ginmiddleware.FlagManagerFromContext(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager in context")
	}
}

func TestInjectFlags_WithUserID(t *testing.T) {
	client := buildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		gotFM = ginmiddleware.FlagManagerFromContext(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user-42")
	router.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager with user context")
	}
}

func TestIsEnabled_ViaGinContext(t *testing.T) {
	client := buildPreloadedClient(t)

	var enabled bool
	var evalErr error
	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		enabled, evalErr = ginmiddleware.IsEnabled(c, "feat")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if !enabled {
		t.Fatal("expected flag to be enabled")
	}
}

func TestIsEnabled_NoManager(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, rec := newTestContext(req)

	_, err := ginmiddleware.IsEnabled(c, "feat")
	if err == nil {
		t.Fatal("expected error when no manager in context")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetString_ViaGinContext(t *testing.T) {
	client := buildPreloadedClient(t)

	var val string
	var evalErr error
	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		val, evalErr = ginmiddleware.GetString(c, "color", "default")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}
}

func TestGetString_NoManager(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(req)

	v, err := ginmiddleware.GetString(c, "color", "fallback")
	if err == nil {
		t.Fatal("expected error")
	}
	if v != "fallback" {
		t.Fatalf("expected fallback, got %q", v)
	}
}

func TestGetNumber_ViaGinContext(t *testing.T) {
	client := buildPreloadedClient(t)

	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		n, err := ginmiddleware.GetNumber(c, "missing-num", 3.14)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 3.14 {
			t.Errorf("expected default 3.14, got %v", n)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetNumber_NoManager(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(req)

	v, err := ginmiddleware.GetNumber(c, "n", 9.9)
	if err == nil {
		t.Fatal("expected error")
	}
	if v != 9.9 {
		t.Fatalf("expected 9.9, got %v", v)
	}
}

func TestFlagManagerFromContext_Nil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(req)

	fm := ginmiddleware.FlagManagerFromContext(c)
	if fm != nil {
		t.Fatal("expected nil when no manager injected")
	}
}

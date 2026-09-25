package ginmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	zenmanage "github.com/zenmanage/zenmanage-go"
	"github.com/zenmanage/zenmanage-go/internal/testutil"
	ginmiddleware "github.com/zenmanage/zenmanage-go/middleware/gin"
)

func newTestContext(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	return c, rec
}

func TestInjectFlags_NoUserID(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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

func TestGetJSON_ViaGinContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	router := gin.New()
	router.Use(ginmiddleware.InjectFlags(client))
	router.GET("/", func(c *gin.Context) {
		v, err := ginmiddleware.GetJSON(c, "config", map[string]any{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		obj, ok := v.(map[string]any)
		if !ok || obj["mode"] != "dark" {
			t.Errorf("expected decoded config map, got %+v", v)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetJSON_NoManager(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(req)

	fallback := map[string]any{"default": true}
	v, err := ginmiddleware.GetJSON(c, "config", fallback)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, ok := v.(map[string]any); !ok || got["default"] != true {
		t.Fatalf("expected fallback default, got %+v", v)
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

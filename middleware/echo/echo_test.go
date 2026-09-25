package echomiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	zenmanage "github.com/zenmanage/zenmanage-go"
	"github.com/zenmanage/zenmanage-go/internal/testutil"
	echomiddleware "github.com/zenmanage/zenmanage-go/middleware/echo"
)

func newTestContext(e *echo.Echo, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestInjectFlags_NoUserID(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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
	client := testutil.BuildPreloadedClient(t)

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

func TestGetJSON_ViaEchoContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	e := echo.New()
	e.Use(echomiddleware.InjectFlags(client))
	e.GET("/", func(c echo.Context) error {
		v, err := echomiddleware.GetJSON(c, "config", map[string]any{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		obj, ok := v.(map[string]any)
		if !ok || obj["mode"] != "dark" {
			t.Errorf("expected decoded config map, got %+v", v)
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetJSON_NoManager(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := newTestContext(e, req)

	fallback := map[string]any{"default": true}
	v, err := echomiddleware.GetJSON(c, "config", fallback)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, ok := v.(map[string]any); !ok || got["default"] != true {
		t.Fatalf("expected fallback default, got %+v", v)
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

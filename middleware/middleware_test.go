package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	zenmanage "github.com/zenmanage/zenmanage-go"
	"github.com/zenmanage/zenmanage-go/internal/testutil"
	"github.com/zenmanage/zenmanage-go/middleware"
)

func TestInjectFlags_NoUserID(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotFM = middleware.FlagManagerFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager in context")
	}
}

func TestInjectFlags_WithUserID(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	var gotFM *zenmanage.FlagManager
	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotFM = middleware.FlagManagerFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user-42")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if gotFM == nil {
		t.Fatal("expected FlagManager with user context")
	}
}

func TestIsEnabled_ViaMWContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	var enabled bool
	var evalErr error
	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		enabled, evalErr = middleware.IsEnabled(r.Context(), "feat")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if !enabled {
		t.Fatal("expected flag to be enabled")
	}
}

func TestIsEnabled_NoManager(t *testing.T) {
	_, err := middleware.IsEnabled(context.Background(), "feat")
	if err == nil {
		t.Fatal("expected error when no manager in context")
	}
}

func TestGetString_ViaMWContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	var val string
	var evalErr error
	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		val, evalErr = middleware.GetString(r.Context(), "color", "default")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}
}

func TestGetString_NoManager(t *testing.T) {
	v, err := middleware.GetString(context.Background(), "color", "fallback")
	if err == nil {
		t.Fatal("expected error")
	}
	if v != "fallback" {
		t.Fatalf("expected fallback, got %q", v)
	}
}

func TestGetNumber_ViaMWContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		n, err := middleware.GetNumber(r.Context(), "missing-num", 3.14)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 3.14 {
			t.Errorf("expected default 3.14, got %v", n)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetNumber_NoManager(t *testing.T) {
	v, err := middleware.GetNumber(context.Background(), "n", 9.9)
	if err == nil {
		t.Fatal("expected error")
	}
	if v != 9.9 {
		t.Fatalf("expected 9.9, got %v", v)
	}
}

func TestGetJSON_ViaMWContext(t *testing.T) {
	client := testutil.BuildPreloadedClient(t)

	var v any
	var evalErr error
	handler := middleware.InjectFlags(client, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		v, evalErr = middleware.GetJSON(r.Context(), "config", map[string]any{})
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if evalErr != nil {
		t.Fatalf("unexpected error: %v", evalErr)
	}
	obj, ok := v.(map[string]any)
	if !ok || obj["mode"] != "dark" {
		t.Fatalf("expected decoded config map, got %+v", v)
	}
}

func TestGetJSON_NoManager(t *testing.T) {
	fallback := map[string]any{"default": true}
	v, err := middleware.GetJSON(context.Background(), "config", fallback)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, ok := v.(map[string]any); !ok || got["default"] != true {
		t.Fatalf("expected fallback default, got %+v", v)
	}
}

func TestFlagManagerFromContext_Nil(t *testing.T) {
	fm := middleware.FlagManagerFromContext(context.Background())
	if fm != nil {
		t.Fatal("expected nil when no manager injected")
	}
}

// Package middleware provides net/http middleware for the Zenmanage Go SDK.
//
// Middleware injects a [FlagManager] scoped to each request into the request
// context. Downstream handlers retrieve it with [FlagManagerFromContext] and
// call flag evaluation methods directly, or use the top-level helpers
// [IsEnabled], [GetString], and [GetNumber].
//
// Example — basic usage:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	    enabled, err := middleware.IsEnabled(r.Context(), "my-flag")
//	    if err != nil || !enabled {
//	        http.NotFound(w, r)
//	        return
//	    }
//	    fmt.Fprintln(w, "flag is on")
//	})
//	http.ListenAndServe(":8080", middleware.InjectFlags(zmClient, mux))
package middleware

import (
	"context"
	"net/http"

	zenmanage "github.com/zenmanage/zenmanage-go"
)

type contextKey struct{}

// InjectFlags returns an [http.Handler] middleware that places a request-scoped
// [zenmanage.FlagManager] into every request context. If the request contains
// a non-empty "X-User-ID" header the flag manager is pre-configured with a
// "user" context using that identifier, enabling per-user rollout evaluation.
//
// Use [FlagManagerFromContext] or the [IsEnabled]/[GetString]/[GetNumber]
// helpers inside your handlers to evaluate flags.
func InjectFlags(client *zenmanage.Zenmanage, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fm := client.Flags()
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			ctx := zenmanage.SingleContext("user", userID, "")
			fm = fm.WithContext(ctx)
		}
		r = r.WithContext(context.WithValue(r.Context(), contextKey{}, fm))
		next.ServeHTTP(w, r)
	})
}

// FlagManagerFromContext retrieves the [zenmanage.FlagManager] injected by
// [InjectFlags]. It returns nil when none is present.
func FlagManagerFromContext(ctx context.Context) *zenmanage.FlagManager {
	fm, _ := ctx.Value(contextKey{}).(*zenmanage.FlagManager)
	return fm
}

// IsEnabled evaluates a boolean flag from the context-injected flag manager.
// It returns false and a non-nil error when no manager is found in ctx or when
// flag evaluation fails.
func IsEnabled(ctx context.Context, key string) (bool, error) {
	fm := FlagManagerFromContext(ctx)
	if fm == nil {
		return false, &zenmanage.EvaluationError{Message: "no flag manager in context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(ctx, key, false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}

// GetString evaluates a string flag from the context-injected flag manager.
func GetString(ctx context.Context, key, defaultValue string) (string, error) {
	fm := FlagManagerFromContext(ctx)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(ctx, key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsString(), nil
}

// GetNumber evaluates a numeric flag from the context-injected flag manager.
func GetNumber(ctx context.Context, key string, defaultValue float64) (float64, error) {
	fm := FlagManagerFromContext(ctx)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(ctx, key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsNumber(), nil
}

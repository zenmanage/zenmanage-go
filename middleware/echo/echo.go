// Package echomiddleware provides an Echo middleware for the Zenmanage Go SDK.
//
// It injects a request-scoped [zenmanage.FlagManager] into the Echo context so
// handlers can evaluate feature flags without holding a reference to the SDK
// client directly.
//
// Example:
//
//	e := echo.New()
//	e.Use(echomiddleware.InjectFlags(zmClient))
//
//	e.GET("/dashboard", func(c echo.Context) error {
//	    enabled, err := echomiddleware.IsEnabled(c, "new-dashboard")
//	    if err != nil || !enabled {
//	        return echo.ErrNotFound
//	    }
//	    return c.JSON(http.StatusOK, map[string]string{"dashboard": "v2"})
//	})
package echomiddleware

import (
	"github.com/labstack/echo/v4"
	zenmanage "github.com/zenmanage/zenmanage-go"
)

const flagManagerKey = "zenmanage_flag_manager"

// InjectFlags returns an Echo middleware that places a request-scoped
// [zenmanage.FlagManager] into the Echo context. If the request contains a
// non-empty "X-User-ID" header, the manager is pre-configured with that user
// identifier for rollout and targeting evaluation.
func InjectFlags(client *zenmanage.Zenmanage) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fm := client.Flags()
			if userID := c.Request().Header.Get("X-User-ID"); userID != "" {
				ctx := zenmanage.SingleContext("user", userID, "")
				fm = fm.WithContext(ctx)
			}
			c.Set(flagManagerKey, fm)
			return next(c)
		}
	}
}

// FlagManagerFromContext retrieves the [zenmanage.FlagManager] injected by
// [InjectFlags]. Returns nil when not present.
func FlagManagerFromContext(c echo.Context) *zenmanage.FlagManager {
	v := c.Get(flagManagerKey)
	if v == nil {
		return nil
	}
	fm, _ := v.(*zenmanage.FlagManager)
	return fm
}

// IsEnabled evaluates a boolean flag using the flag manager stored in the Echo
// context.
func IsEnabled(c echo.Context, key string) (bool, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		return false, &zenmanage.EvaluationError{Message: "no flag manager in echo context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request().Context(), key, false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}

// GetString evaluates a string flag using the flag manager stored in the Echo
// context.
func GetString(c echo.Context, key, defaultValue string) (string, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in echo context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request().Context(), key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsString(), nil
}

// GetNumber evaluates a numeric flag using the flag manager stored in the Echo
// context.
func GetNumber(c echo.Context, key string, defaultValue float64) (float64, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in echo context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request().Context(), key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsNumber(), nil
}

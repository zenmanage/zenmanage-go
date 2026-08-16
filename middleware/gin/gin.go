// Package ginmiddleware provides a Gin middleware for the Zenmanage Go SDK.
//
// It injects a request-scoped [zenmanage.FlagManager] into the Gin context so
// handlers can evaluate feature flags without holding a reference to the SDK
// client directly.
//
// Example:
//
//	r := gin.Default()
//	r.Use(ginmiddleware.InjectFlags(zmClient))
//
//	r.GET("/dashboard", func(c *gin.Context) {
//	    enabled, err := ginmiddleware.IsEnabled(c, "new-dashboard")
//	    if err != nil || !enabled {
//	        c.AbortWithStatus(http.StatusNotFound)
//	        return
//	    }
//	    c.JSON(http.StatusOK, gin.H{"dashboard": "v2"})
//	})
package ginmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	zenmanage "github.com/zenmanage/zenmanage-go"
)

const flagManagerKey = "zenmanage_flag_manager"

// InjectFlags returns a Gin middleware that places a request-scoped
// [zenmanage.FlagManager] into the Gin context. If the request contains a
// non-empty "X-User-ID" header, the manager is pre-configured with that user
// identifier for rollout and targeting evaluation.
func InjectFlags(client *zenmanage.Zenmanage) gin.HandlerFunc {
	return func(c *gin.Context) {
		fm := client.Flags()
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			ctx := zenmanage.SingleContext("user", userID, "")
			fm = fm.WithContext(ctx)
		}
		c.Set(flagManagerKey, fm)
		c.Next()
	}
}

// FlagManagerFromContext retrieves the [zenmanage.FlagManager] injected by
// [InjectFlags]. Returns nil when not present.
func FlagManagerFromContext(c *gin.Context) *zenmanage.FlagManager {
	v, exists := c.Get(flagManagerKey)
	if !exists {
		return nil
	}
	fm, _ := v.(*zenmanage.FlagManager)
	return fm
}

// IsEnabled evaluates a boolean flag using the flag manager stored in the Gin
// context. Aborts with 500 when no manager is present.
func IsEnabled(c *gin.Context, key string) (bool, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return false, &zenmanage.EvaluationError{Message: "no flag manager in gin context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request.Context(), key, false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}

// GetString evaluates a string flag using the flag manager stored in the Gin
// context.
func GetString(c *gin.Context, key, defaultValue string) (string, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in gin context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request.Context(), key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsString(), nil
}

// GetNumber evaluates a numeric flag using the flag manager stored in the Gin
// context.
func GetNumber(c *gin.Context, key string, defaultValue float64) (float64, error) {
	fm := FlagManagerFromContext(c)
	if fm == nil {
		return defaultValue, &zenmanage.EvaluationError{Message: "no flag manager in gin context; use InjectFlags middleware"}
	}
	flag, err := fm.Single(c.Request.Context(), key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsNumber(), nil
}

// Package zenmanage is the Zenmanage feature-flag SDK for Go server
// applications. It evaluates flags locally against cached rules fetched
// from the Zenmanage API, supporting context-based targeting, percentage
// rollouts, and usage reporting.
//
// See the [Zenmanage] type for the main entry point, and the middleware
// sub-package for framework integrations.
package zenmanage

import "context"

// Zenmanage is the main SDK entry point.
type Zenmanage struct {
	flagManager *FlagManager
}

// New creates a Zenmanage SDK instance.
func New(config Config) *Zenmanage {
	apiClient := NewAPIClient(config)
	ruleEngine := NewRuleEngine()
	flagManager := NewFlagManager(apiClient, config.CustomCache, ruleEngine, config.CacheTTL, config.Logger)
	return &Zenmanage{flagManager: flagManager}
}

// Flags returns the flag manager for advanced use (context, defaults, bulk evaluation).
func (z *Zenmanage) Flags() *FlagManager {
	return z.flagManager
}

// IsEnabled returns whether a boolean flag is enabled for the given user ID.
// When userID is non-empty the flag is evaluated against a "user" context so
// rollout rules and targeting apply. Pass an empty userID to evaluate without
// context (e.g. kill-switch style flags).
func (z *Zenmanage) IsEnabled(ctx context.Context, key, userID string) (bool, error) {
	fm := z.flagManager
	if userID != "" {
		c := SingleContext("user", userID, "")
		fm = fm.WithContext(c)
	}
	flag, err := fm.Single(ctx, key, false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}

// GetString returns the string value of a flag for the given user ID.
// defaultValue is returned when the flag is not found.
func (z *Zenmanage) GetString(ctx context.Context, key, userID, defaultValue string) (string, error) {
	fm := z.flagManager
	if userID != "" {
		c := SingleContext("user", userID, "")
		fm = fm.WithContext(c)
	}
	flag, err := fm.Single(ctx, key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsString(), nil
}

// GetNumber returns the numeric value of a flag for the given user ID.
// defaultValue is returned when the flag is not found.
func (z *Zenmanage) GetNumber(ctx context.Context, key, userID string, defaultValue float64) (float64, error) {
	fm := z.flagManager
	if userID != "" {
		c := SingleContext("user", userID, "")
		fm = fm.WithContext(c)
	}
	flag, err := fm.Single(ctx, key, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	return flag.AsNumber(), nil
}

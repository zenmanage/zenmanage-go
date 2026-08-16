package zenmanage

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

const rulesCacheKey = "zenmanage_rules"

// FlagManager handles rule loading and flag evaluation.
type FlagManager struct {
	apiClient  *APIClient
	cache      Cache
	ruleEngine *RuleEngine
	cacheTTL   time.Duration
	logger     Logger

	mu       sync.RWMutex
	rules    *RulesResponse
	context  *Context
	defaults *DefaultsCollection
}

// NewFlagManager creates a flag manager.
func NewFlagManager(apiClient *APIClient, cache Cache, ruleEngine *RuleEngine, cacheTTL time.Duration, logger Logger) *FlagManager {
	return &FlagManager{
		apiClient:  apiClient,
		cache:      cache,
		ruleEngine: ruleEngine,
		cacheTTL:   cacheTTL,
		logger:     logger,
	}
}

// WithContext returns a new flag manager that shares rules/cache with the receiver but uses a different context.
func (m *FlagManager) WithContext(ctx Context) *FlagManager {
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()
	clone := NewFlagManager(m.apiClient, m.cache, m.ruleEngine, m.cacheTTL, m.logger)
	clone.context = &ctx
	clone.defaults = m.defaults
	clone.mu.Lock()
	clone.rules = rules
	clone.mu.Unlock()
	return clone
}

// WithDefaults returns a new flag manager that shares rules/cache with the receiver but uses a different defaults collection.
func (m *FlagManager) WithDefaults(defaults *DefaultsCollection) *FlagManager {
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()
	clone := NewFlagManager(m.apiClient, m.cache, m.ruleEngine, m.cacheTTL, m.logger)
	clone.context = m.context
	clone.defaults = defaults
	clone.mu.Lock()
	clone.rules = rules
	clone.mu.Unlock()
	return clone
}

// RefreshRules forces rules refresh from API.
func (m *FlagManager) RefreshRules(ctx context.Context) error {
	m.mu.Lock()
	m.rules = nil
	m.mu.Unlock()
	_ = m.cache.Delete(rulesCacheKey)
	_, err := m.loadRules(ctx)
	return err
}

// All returns all evaluated flags.
func (m *FlagManager) All(ctx context.Context) ([]Flag, error) {
	rules, err := m.loadRules(ctx)
	if err != nil {
		return nil, err
	}
	contextValue := m.getContext()
	out := make([]Flag, 0, len(rules.Flags))
	for _, f := range rules.Flags {
		flag, err := m.evaluateFlag(f, contextValue)
		if err != nil {
			return nil, err
		}
		out = append(out, flag)
		m.reportUsageAsync(f.Key, contextValue, nil)
	}
	return out, nil
}

// Single returns one evaluated flag by key.
func (m *FlagManager) Single(ctx context.Context, key string, inlineDefault ...any) (Flag, error) {
	rules, err := m.loadRules(ctx)
	if err != nil {
		return Flag{}, err
	}
	contextValue := m.getContext()
	for _, f := range rules.Flags {
		if f.Key != key {
			continue
		}
		flag, err := m.evaluateFlag(f, contextValue)
		if err != nil {
			return Flag{}, err
		}
		m.reportUsageAsync(key, contextValue, nil)
		return flag, nil
	}

	if len(inlineDefault) > 0 {
		flag := newDefaultFlag(key, inlineDefault[0])
		m.reportUsageAsync(key, contextValue, inlineDefault[0])
		return flag, nil
	}
	if m.defaults != nil {
		if v, ok := m.defaults.Get(key); ok {
			flag := newDefaultFlag(key, v)
			m.reportUsageAsync(key, contextValue, v)
			return flag, nil
		}
	}
	return Flag{}, &EvaluationError{Message: "flag not found and no default provided: " + key}
}

func (m *FlagManager) getContext() *Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.context == nil {
		return nil
	}
	ctx := *m.context
	return &ctx
}

func (m *FlagManager) evaluateFlag(data FlagData, ctx *Context) (Flag, error) {
	activeTarget := data.Target
	activeRules := data.Rules

	if data.Rollout != nil && ctx != nil && ctx.Identifier() != "" {
		if IsInBucket(data.Rollout.Salt, ctx.Identifier(), data.Rollout.Percentage) {
			activeTarget = data.Rollout.Target
			activeRules = data.Rollout.Rules
		}
	}

	if len(activeRules) > 0 {
		contextValue := Context{}
		if ctx != nil {
			contextValue = *ctx
		}
		ruleValue, err := m.ruleEngine.Evaluate(activeRules, contextValue)
		if err != nil {
			return Flag{}, err
		}
		if ruleValue != nil {
			activeTarget.Value = *ruleValue
		}
	}

	return newFlag(data, activeTarget, activeRules), nil
}

func (m *FlagManager) loadRules(ctx context.Context) (RulesResponse, error) {
	m.mu.RLock()
	if m.rules != nil {
		cached := *m.rules
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	if raw, found, err := m.cache.Get(rulesCacheKey); err == nil && found {
		var cached RulesResponse
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			m.mu.Lock()
			m.rules = &cached
			m.mu.Unlock()
			return cached, nil
		}
		m.logger.Warn("failed to decode cached rules", map[string]any{"error": err.Error()})
	}

	fresh, err := m.apiClient.FetchRules(ctx)
	if err != nil {
		return RulesResponse{}, err
	}

	serialized, err := json.Marshal(fresh)
	if err == nil {
		if err := m.cache.Set(rulesCacheKey, string(serialized), m.cacheTTL); err != nil {
			m.logger.Warn("failed to persist rules cache", map[string]any{"error": err.Error()})
		}
	}

	m.mu.Lock()
	m.rules = &fresh
	m.mu.Unlock()
	return fresh, nil
}

func (m *FlagManager) reportUsageAsync(key string, contextValue *Context, defaultValue any) {
	if key == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.apiClient.ReportUsage(ctx, key, contextValue, defaultValue); err != nil {
			m.logger.Debug("usage reporting failed", map[string]any{"flag": key, "error": err.Error()})
		}
	}()
}

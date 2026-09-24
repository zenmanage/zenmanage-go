package zenmanage

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

const rulesCacheKey = "zenmanage_rules"

// flagIndex pairs a loaded RulesResponse with a key->FlagData lookup map
// built once from it, so repeated Single() calls against the same loaded
// rules don't each re-scan the flag list. On duplicate keys within a
// payload, the first occurrence wins, matching the previous linear-scan
// behavior. Both fields are immutable after construction, so a *flagIndex
// can be shared across FlagManager clones (see clone()) without copying.
type flagIndex struct {
	rules RulesResponse
	byKey map[string]FlagData
}

func newFlagIndex(rules RulesResponse) *flagIndex {
	byKey := make(map[string]FlagData, len(rules.Flags))
	for _, f := range rules.Flags {
		if _, exists := byKey[f.Key]; !exists {
			byKey[f.Key] = f
		}
	}
	return &flagIndex{rules: rules, byKey: byKey}
}

// FlagManager handles rule loading and flag evaluation.
type FlagManager struct {
	apiClient  *APIClient
	cache      Cache
	ruleEngine *RuleEngine
	cacheTTL   time.Duration
	logger     Logger

	mu       sync.RWMutex
	rules    *flagIndex
	context  *Context
	defaults *DefaultsCollection

	// warnedTypes tracks which unrecognized FlagType values have already
	// been logged, so a flag type this SDK release doesn't know about
	// (e.g. "json") produces one warning, not one per lookup/evaluation.
	// It's a pointer shared across every clone derived from the same root
	// manager (via WithContext/WithDefaults) so the convenience methods on
	// Zenmanage, which clone a fresh FlagManager per call, still dedupe.
	warnedTypes *sync.Map
}

// NewFlagManager creates a flag manager.
func NewFlagManager(apiClient *APIClient, cache Cache, ruleEngine *RuleEngine, cacheTTL time.Duration, logger Logger) *FlagManager {
	return &FlagManager{
		apiClient:   apiClient,
		cache:       cache,
		ruleEngine:  ruleEngine,
		cacheTTL:    cacheTTL,
		logger:      logger,
		warnedTypes: &sync.Map{},
	}
}

// clone builds a new flag manager sharing the receiver's dependencies,
// warned-types dedup state, and cached rules. It's the shared base for
// WithContext/WithDefaults, which are called on every request in typical
// per-request middleware usage, so unlike NewFlagManager it doesn't
// allocate a fresh warnedTypes map just to immediately replace it.
func (m *FlagManager) clone() *FlagManager {
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()
	clone := &FlagManager{
		apiClient:   m.apiClient,
		cache:       m.cache,
		ruleEngine:  m.ruleEngine,
		cacheTTL:    m.cacheTTL,
		logger:      m.logger,
		warnedTypes: m.warnedTypes,
	}
	clone.rules = rules
	return clone
}

// WithContext returns a new flag manager that shares rules/cache with the receiver but uses a different context.
func (m *FlagManager) WithContext(ctx Context) *FlagManager {
	clone := m.clone()
	clone.context = &ctx
	clone.defaults = m.defaults
	return clone
}

// WithDefaults returns a new flag manager that shares rules/cache with the receiver but uses a different defaults collection.
func (m *FlagManager) WithDefaults(defaults *DefaultsCollection) *FlagManager {
	clone := m.clone()
	clone.context = m.context
	clone.defaults = defaults
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
	idx, err := m.loadRules(ctx)
	if err != nil {
		return nil, err
	}
	contextValue := m.getContext()
	out := make([]Flag, 0, len(idx.rules.Flags))
	for _, f := range idx.rules.Flags {
		if !isKnownFlagType(f.Type) {
			m.warnUnknownFlagType(f.Key, f.Type)
			continue
		}
		flag, err := m.evaluateFlag(f, contextValue)
		if err != nil {
			return nil, err
		}
		out = append(out, flag)
	}
	return out, nil
}

// Single returns one evaluated flag by key.
func (m *FlagManager) Single(ctx context.Context, key string, inlineDefault ...any) (Flag, error) {
	idx, err := m.loadRules(ctx)
	if err != nil {
		return Flag{}, err
	}
	contextValue := m.getContext()
	if f, ok := idx.byKey[key]; ok {
		if !isKnownFlagType(f.Type) {
			// A flag type this SDK release doesn't recognize yet (e.g. a
			// newer "json" flag served to an older release) can't be
			// evaluated meaningfully — degrade to the caller's default
			// exactly as if the flag were absent, rather than returning a
			// zero-value/garbage result for an unrecognized type.
			m.warnUnknownFlagType(f.Key, f.Type)
		} else {
			flag, err := m.evaluateFlag(f, contextValue)
			if err != nil {
				return Flag{}, err
			}
			m.reportUsageAsync(key, contextValue, m.resolveEffectiveDefault(key, inlineDefault...))
			return flag, nil
		}
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

// ReportUsage manually reports flag usage to the API, using the manager's
// current context (if any). Single already reports usage automatically on
// every evaluation; call this directly only when usage needs to be recorded
// outside of a Single/All evaluation path, matching the explicit reportUsage
// method exposed by the JavaScript, PHP, and Python SDKs.
func (m *FlagManager) ReportUsage(ctx context.Context, key string, defaultValue any) error {
	return m.apiClient.ReportUsage(ctx, key, m.getContext(), defaultValue)
}

// resolveEffectiveDefault returns the default value that would be used for key,
// preferring the inline default and falling back to a DefaultsCollection entry.
func (m *FlagManager) resolveEffectiveDefault(key string, inlineDefault ...any) any {
	if len(inlineDefault) > 0 {
		return inlineDefault[0]
	}
	if m.defaults != nil {
		if v, ok := m.defaults.Get(key); ok {
			return v
		}
	}
	return nil
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

func (m *FlagManager) loadRules(ctx context.Context) (*flagIndex, error) {
	m.mu.RLock()
	if m.rules != nil {
		idx := m.rules
		m.mu.RUnlock()
		return idx, nil
	}
	m.mu.RUnlock()

	if raw, found, err := m.cache.Get(rulesCacheKey); err == nil && found {
		var cached RulesResponse
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			idx := newFlagIndex(cached)
			m.mu.Lock()
			m.rules = idx
			m.mu.Unlock()
			return idx, nil
		}
		m.logger.Warn("failed to decode cached rules", map[string]any{"error": err.Error()})
	}

	fresh, err := m.apiClient.FetchRules(ctx)
	if err != nil {
		return nil, err
	}

	serialized, err := json.Marshal(fresh)
	if err == nil {
		if err := m.cache.Set(rulesCacheKey, string(serialized), m.cacheTTL); err != nil {
			m.logger.Warn("failed to persist rules cache", map[string]any{"error": err.Error()})
		}
	}

	idx := newFlagIndex(fresh)
	m.mu.Lock()
	m.rules = idx
	m.mu.Unlock()
	return idx, nil
}

// warnUnknownFlagType logs once (per distinct unrecognized FlagType value,
// for the lifetime of the root manager and every manager cloned from it via
// WithContext/WithDefaults) that a flag was skipped because this SDK
// release doesn't know its type.
func (m *FlagManager) warnUnknownFlagType(key string, flagType FlagType) {
	if _, alreadyWarned := m.warnedTypes.LoadOrStore(string(flagType), true); alreadyWarned {
		return
	}
	m.logger.Warn("skipping flag with unrecognized type; falling back to the caller's default", map[string]any{
		"flag": key,
		"type": string(flagType),
	})
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

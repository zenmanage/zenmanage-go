package zenmanage

import "encoding/json"

// FlagType is the primitive type for a flag.
type FlagType string

const (
	// FlagTypeBoolean represents boolean flags.
	FlagTypeBoolean FlagType = "boolean"
	// FlagTypeString represents string flags.
	FlagTypeString FlagType = "string"
	// FlagTypeNumber represents numeric flags.
	FlagTypeNumber FlagType = "number"
)

// ContextValue is a single context value.
type ContextValue struct {
	Value string `json:"value"`
}

// ContextAttribute is an attribute used for context targeting.
type ContextAttribute struct {
	Key    string         `json:"key"`
	Values []ContextValue `json:"values"`
}

// ContextData is the serialized context payload.
type ContextData struct {
	Type       string             `json:"type"`
	Name       string             `json:"name,omitempty"`
	Identifier string             `json:"identifier,omitempty"`
	Attributes []ContextAttribute `json:"attributes,omitempty"`
}

// RuleContextTarget is a typed context target used in context/segment clauses.
type RuleContextTarget struct {
	Identifier string  `json:"identifier"`
	Type       *string `json:"type,omitempty"`
}

// RuleCondition is a single clause for rule matching.
type RuleCondition struct {
	Attribute string
	Operator  string
	Value     any
}

// UnmarshalJSON maps both the legacy internal format (attribute/operator/value)
// and the CDN wire format (selector/selector_subtype/comparer/values) to the
// internal fields used by the rule engine.
func (rc *RuleCondition) UnmarshalJSON(data []byte) error {
	var raw struct {
		// CDN wire format
		Selector        string            `json:"selector"`
		SelectorSubtype *string           `json:"selector_subtype"`
		Comparer        string            `json:"comparer"`
		Values          []json.RawMessage `json:"values"`
		// Legacy / internal format
		Attribute string          `json:"attribute"`
		Operator  string          `json:"operator"`
		Value     json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Selector != "" {
		// CDN format
		rc.Operator = raw.Comparer
		switch raw.Selector {
		case "context", "segment":
			rc.Attribute = raw.Selector
		default: // "attribute"
			if raw.SelectorSubtype != nil {
				rc.Attribute = *raw.SelectorSubtype
			}
		}
		values := make([]any, 0, len(raw.Values))
		for _, v := range raw.Values {
			var parsed any
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			values = append(values, parsed)
		}
		rc.Value = values
		return nil
	}
	// Legacy format
	rc.Attribute = raw.Attribute
	rc.Operator = raw.Operator
	if raw.Value != nil {
		var v any
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return err
		}
		rc.Value = v
	}
	return nil
}

// ValueEnvelope is the nested API value format.
type ValueEnvelope struct {
	Version string `json:"version,omitempty"`
	Value   struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	} `json:"value"`
}

// Target contains a value payload and metadata.
type Target struct {
	Version     string        `json:"version,omitempty"`
	ExpiredAt   string        `json:"expired_at,omitempty"`
	PublishedAt string        `json:"published_at,omitempty"`
	ScheduledAt string        `json:"scheduled_at,omitempty"`
	Value       ValueEnvelope `json:"value"`
}

// Rule holds rule criteria and the resulting target value.
type Rule struct {
	Version     string          `json:"version,omitempty"`
	Description string          `json:"description,omitempty"`
	Criteria    *RuleCondition  `json:"criteria,omitempty"`
	Clauses     []RuleCondition `json:"clauses,omitempty"`
	Position    int             `json:"position,omitempty"`
	Value       ValueEnvelope   `json:"value"`
}

// RolloutData configures percentage rollout behavior for a flag.
type RolloutData struct {
	Target     Target `json:"target"`
	Rules      []Rule `json:"rules"`
	Percentage int    `json:"percentage"`
	Salt       string `json:"salt"`
	Status     string `json:"status"`
}

// FlagData is the API representation of a flag.
type FlagData struct {
	Version string       `json:"version"`
	Type    FlagType     `json:"type"`
	Key     string       `json:"key"`
	Name    string       `json:"name"`
	Target  Target       `json:"target"`
	Rules   []Rule       `json:"rules,omitempty"`
	Rollout *RolloutData `json:"rollout,omitempty"`
}

// RulesResponse is the top-level rule payload.
type RulesResponse struct {
	Version string     `json:"version"`
	Flags   []FlagData `json:"flags"`
}

package zenmanage

import (
	"strconv"
)

// Flag is an evaluated flag object with helper accessors.
type Flag struct {
	version string
	typ     FlagType
	key     string
	name    string
	target  Target
	rules   []Rule
	rollout *RolloutData
}

func newFlag(data FlagData, target Target, rules []Rule) Flag {
	return Flag{
		version: data.Version,
		typ:     data.Type,
		key:     data.Key,
		name:    data.Name,
		target:  target,
		rules:   append([]Rule{}, rules...),
		rollout: data.Rollout,
	}
}

func newDefaultFlag(key string, value any) Flag {
	data := FlagData{Version: "default", Key: key, Name: key}
	var envelope ValueEnvelope
	switch v := value.(type) {
	case bool:
		data.Type = FlagTypeBoolean
		envelope.Value.Boolean = &v
	case string:
		data.Type = FlagTypeString
		envelope.Value.String = &v
	case int:
		data.Type = FlagTypeNumber
		f := float64(v)
		envelope.Value.Number = &f
	case int64:
		data.Type = FlagTypeNumber
		f := float64(v)
		envelope.Value.Number = &f
	case float64:
		data.Type = FlagTypeNumber
		envelope.Value.Number = &v
	case float32:
		data.Type = FlagTypeNumber
		f := float64(v)
		envelope.Value.Number = &f
	default:
		// Preserve compatibility by coercing unknown defaults to string.
		s := ""
		data.Type = FlagTypeString
		envelope.Value.String = &s
	}
	return Flag{
		version: data.Version,
		typ:     data.Type,
		key:     data.Key,
		name:    data.Name,
		target:  Target{Value: envelope},
	}
}

// Version returns flag version.
func (f Flag) Version() string { return f.version }

// Type returns flag primitive type.
func (f Flag) Type() FlagType { return f.typ }

// Key returns flag key.
func (f Flag) Key() string { return f.key }

// Name returns display name.
func (f Flag) Name() string { return f.name }

// Target returns raw target.
func (f Flag) Target() Target { return f.target }

// Rules returns the effective rules.
func (f Flag) Rules() []Rule { return append([]Rule{}, f.rules...) }

// Rollout returns rollout metadata if present.
func (f Flag) Rollout() *RolloutData { return f.rollout }

// Value returns the raw typed value.
func (f Flag) Value() any {
	switch f.typ {
	case FlagTypeBoolean:
		if f.target.Value.Value.Boolean != nil {
			return *f.target.Value.Value.Boolean
		}
		return false
	case FlagTypeNumber:
		if f.target.Value.Value.Number != nil {
			return *f.target.Value.Value.Number
		}
		return float64(0)
	default:
		if f.target.Value.Value.String != nil {
			return *f.target.Value.Value.String
		}
		return ""
	}
}

// AsBool coerces value to boolean.
func (f Flag) AsBool() bool {
	switch v := f.Value().(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case float64:
		return v != 0
	default:
		return false
	}
}

// AsString coerces value to string.
func (f Flag) AsString() string {
	switch v := f.Value().(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

// AsNumber coerces value to float64.
func (f Flag) AsNumber() float64 {
	switch v := f.Value().(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return n
	case float64:
		return v
	default:
		return 0
	}
}

// IsEnabled is true when the flag resolves to boolean true.
func (f Flag) IsEnabled() bool {
	return f.typ == FlagTypeBoolean && f.AsBool()
}

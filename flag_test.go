package zenmanage

import (
	"reflect"
	"testing"
)

func TestFlagConversions(t *testing.T) {
	b := true
	s := "hello"
	n := 42.5

	fb := Flag{typ: FlagTypeBoolean, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{Boolean: &b}}}}
	if !fb.IsEnabled() || fb.AsNumber() != 1 || fb.AsString() != "true" {
		t.Fatalf("boolean conversions failed")
	}

	fs := Flag{typ: FlagTypeString, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{String: &s}}}}
	if fs.AsString() != "hello" || fs.AsBool() != true {
		t.Fatalf("string conversions failed")
	}

	fn := Flag{typ: FlagTypeNumber, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{Number: &n}}}}
	if fn.AsNumber() != 42.5 || fn.AsString() != "42.5" {
		t.Fatalf("number conversions failed")
	}
}

func TestFlagAsJSON(t *testing.T) {
	obj := map[string]any{"mode": "dark", "count": float64(3)}
	list := []any{float64(1), float64(2), float64(3)}

	fj := Flag{typ: FlagTypeJSON, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{JSON: obj}}}}
	if got := fj.AsJSON(); !reflect.DeepEqual(got, obj) {
		t.Fatalf("expected AsJSON() to return the decoded object %+v, got %+v", obj, got)
	}
	// Calling a mismatched accessor on a json flag must fall back to the
	// safe zero value, never a lossy conversion.
	if fj.AsString() != "" || fj.AsNumber() != 0 || fj.AsBool() {
		t.Fatalf("expected mismatched accessors on a json flag to return safe zero values")
	}

	fl := Flag{typ: FlagTypeJSON, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{JSON: list}}}}
	if got := fl.AsJSON(); !reflect.DeepEqual(got, list) {
		t.Fatalf("expected AsJSON() to return the decoded array %+v, got %+v", list, got)
	}

	// AsJSON() on a non-json flag returns an empty map, not the flag's own
	// value coerced into a structure.
	fb := Flag{typ: FlagTypeBoolean, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{Boolean: boolPtr(true)}}}}
	if got := fb.AsJSON(); !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("expected AsJSON() on a non-json flag to return an empty map, got %+v", got)
	}

	// A json flag with no populated value falls back to the same empty map.
	fEmpty := Flag{typ: FlagTypeJSON}
	if got := fEmpty.AsJSON(); !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("expected AsJSON() on an empty json flag to return an empty map, got %+v", got)
	}
}

func boolPtr(b bool) *bool { return &b }

func TestDefaultFlagTypeInference(t *testing.T) {
	if f := newDefaultFlag("a", true); f.Type() != FlagTypeBoolean {
		t.Fatalf("expected boolean default type")
	}
	if f := newDefaultFlag("a", "v"); f.Type() != FlagTypeString {
		t.Fatalf("expected string default type")
	}
	if f := newDefaultFlag("a", 2); f.Type() != FlagTypeNumber {
		t.Fatalf("expected number default type")
	}
	if f := newDefaultFlag("a", map[string]any{"x": 1}); f.Type() != FlagTypeJSON {
		t.Fatalf("expected json default type for a map default")
	}
	if f := newDefaultFlag("a", []any{1, 2}); f.Type() != FlagTypeJSON {
		t.Fatalf("expected json default type for a slice default")
	}
}

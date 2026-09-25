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
	// safe zero value, never a lossy conversion — except AsBool(), which per
	// the documented cross-SDK coercion contract returns true for every
	// non-boolean type regardless of the underlying value.
	if fj.AsString() != "" || fj.AsNumber() != 0 || !fj.AsBool() {
		t.Fatalf("expected mismatched string/number accessors on a json flag to return safe zero values and AsBool() to return true")
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

// TestFlagAsBoolJSONAlwaysTrue confirms ZEN-1752: AsBool() on a json-typed
// flag must return true regardless of the underlying decoded value —
// including falsy-looking values like an empty object/array, or no value at
// all — matching the documented cross-SDK coercion contract (every
// non-boolean type is truthy for AsBool()).
func TestFlagAsBoolJSONAlwaysTrue(t *testing.T) {
	newJSONFlag := func(value any) Flag {
		return Flag{typ: FlagTypeJSON, target: Target{Value: ValueEnvelope{Value: struct {
			Boolean *bool    `json:"boolean,omitempty"`
			String  *string  `json:"string,omitempty"`
			Number  *float64 `json:"number,omitempty"`
			JSON    any      `json:"json,omitempty"`
		}{JSON: value}}}}
	}

	cases := []struct {
		name  string
		value any
	}{
		{"populated object", map[string]any{"mode": "dark"}},
		{"populated array", []any{1, 2, 3}},
		{"empty object", map[string]any{}},
		{"empty array", []any{}},
		{"nil value", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !newJSONFlag(tc.value).AsBool() {
				t.Fatalf("expected AsBool() on a json flag with value %+v to return true", tc.value)
			}
		})
	}

	// A json flag with no value set at all (zero-value Target) must also be
	// truthy.
	if !(Flag{typ: FlagTypeJSON}).AsBool() {
		t.Fatalf("expected AsBool() on an empty json flag to return true")
	}
}

// TestFlagAsBoolNumberAndStringAlwaysTrue confirms AsBool() returns true
// unconditionally for number/string flags too, per the same coercion
// contract as TestFlagAsBoolJSONAlwaysTrue — this was a pre-existing gap
// (value-dependent truthy checks) that predates json support entirely.
func TestFlagAsBoolNumberAndStringAlwaysTrue(t *testing.T) {
	zero := 0.0
	empty := ""
	falseStr := "false"

	fn := Flag{typ: FlagTypeNumber, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{Number: &zero}}}}
	if !fn.AsBool() {
		t.Fatalf("expected AsBool() on a number flag set to 0 to return true")
	}

	fs := Flag{typ: FlagTypeString, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{String: &empty}}}}
	if !fs.AsBool() {
		t.Fatalf("expected AsBool() on a string flag set to \"\" to return true")
	}

	ffalse := Flag{typ: FlagTypeString, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
		JSON    any      `json:"json,omitempty"`
	}{String: &falseStr}}}}
	if !ffalse.AsBool() {
		t.Fatalf(`expected AsBool() on a string flag set to "false" to return true`)
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
	// A concretely-typed map/slice (not exactly map[string]any/[]any) must
	// still be typed as json and preserved unchanged, not discarded to an
	// empty string/map — GetJSON()/AsJSON() must hand the caller back their
	// own default value, whatever its concrete Go type.
	typedMapDefault := map[string]string{"mode": "dark"}
	f := newDefaultFlag("a", typedMapDefault)
	if f.Type() != FlagTypeJSON {
		t.Fatalf("expected json default type for a map[string]string default")
	}
	if got, ok := f.AsJSON().(map[string]string); !ok || !reflect.DeepEqual(got, typedMapDefault) {
		t.Fatalf("expected AsJSON() to return the typed map default unchanged, got %+v", f.AsJSON())
	}

	typedSliceDefault := []int{1, 2, 3}
	f = newDefaultFlag("a", typedSliceDefault)
	if f.Type() != FlagTypeJSON {
		t.Fatalf("expected json default type for a []int default")
	}
	if got, ok := f.AsJSON().([]int); !ok || !reflect.DeepEqual(got, typedSliceDefault) {
		t.Fatalf("expected AsJSON() to return the typed slice default unchanged, got %+v", f.AsJSON())
	}
}

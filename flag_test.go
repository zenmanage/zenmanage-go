package zenmanage

import "testing"

func TestFlagConversions(t *testing.T) {
	b := true
	s := "hello"
	n := 42.5

	fb := Flag{typ: FlagTypeBoolean, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{Boolean: &b}}}}
	if !fb.IsEnabled() || fb.AsNumber() != 1 || fb.AsString() != "true" {
		t.Fatalf("boolean conversions failed")
	}

	fs := Flag{typ: FlagTypeString, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{String: &s}}}}
	if fs.AsString() != "hello" || fs.AsBool() != true {
		t.Fatalf("string conversions failed")
	}

	fn := Flag{typ: FlagTypeNumber, target: Target{Value: ValueEnvelope{Value: struct {
		Boolean *bool    `json:"boolean,omitempty"`
		String  *string  `json:"string,omitempty"`
		Number  *float64 `json:"number,omitempty"`
	}{Number: &n}}}}
	if fn.AsNumber() != 42.5 || fn.AsString() != "42.5" {
		t.Fatalf("number conversions failed")
	}
}

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
}

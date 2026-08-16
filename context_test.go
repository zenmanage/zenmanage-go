package zenmanage

import "testing"

func TestContextSingleAndAttributes(t *testing.T) {
	ctx := SingleContext("user", "u-1", "Test User")
	if ctx.Type() != "user" || ctx.Identifier() != "u-1" || ctx.Name() != "Test User" {
		t.Fatalf("unexpected context values")
	}

	attr := NewAttribute("country", []string{"US"})
	ctx.AddAttribute(attr)

	got, ok := ctx.GetAttribute("country")
	if !ok || len(got.Values) != 1 || got.Values[0].Value != "US" {
		t.Fatalf("expected country attribute")
	}

	payload, err := ctx.JSON()
	if err != nil || len(payload) == 0 {
		t.Fatalf("expected context json")
	}
}

func TestContextIsEmpty(t *testing.T) {
	if !(Context{}).IsEmpty() {
		t.Fatalf("expected zero context to be empty")
	}
	if SingleContext("user", "u-2", "").IsEmpty() {
		t.Fatalf("expected context with type/id to be non-empty")
	}
}

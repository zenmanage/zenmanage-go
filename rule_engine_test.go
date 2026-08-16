package zenmanage

import "testing"

func TestRuleEngineEqualsAndIn(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", []Attribute{
		NewAttribute("country", []string{"US"}),
	})

	str := "enabled"
	rules := []Rule{
		{
			Clauses: []RuleCondition{{Attribute: "country", Operator: "equal", Value: "US"}},
			Value: ValueEnvelope{Value: struct {
				Boolean *bool    `json:"boolean,omitempty"`
				String  *string  `json:"string,omitempty"`
				Number  *float64 `json:"number,omitempty"`
			}{String: &str}},
		},
	}

	v, err := engine.Evaluate(rules, ctx)
	if err != nil || v == nil || v.Value.String == nil || *v.Value.String != "enabled" {
		t.Fatalf("expected matching equals rule")
	}
}

func TestRuleEngineNumericComparison(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", []Attribute{NewAttribute("age", []string{"21"})})

	rule := Rule{Clauses: []RuleCondition{{Attribute: "age", Operator: "gte", Value: "18"}}}
	ok, err := engine.matchesRule(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("expected gte match")
	}
}

func TestRuleEngineRegexAndUnsupported(t *testing.T) {
	engine := NewRuleEngine()
	ctx := NewContext("user", "u-1", "", []Attribute{NewAttribute("email", []string{"a@b.com"})})

	ok, err := engine.matchesCondition(RuleCondition{Attribute: "email", Operator: "regex", Value: `.+@.+`}, ctx)
	if err != nil || !ok {
		t.Fatalf("expected regex to match")
	}

	_, err = engine.matchesCondition(RuleCondition{Attribute: "email", Operator: "made_up", Value: "x"}, ctx)
	if err == nil {
		t.Fatalf("expected unsupported operator error")
	}
}

func TestRuleEngineContextTarget(t *testing.T) {
	engine := NewRuleEngine()
	typ := "user"
	ctx := SingleContext("user", "u-1", "")

	value := []any{map[string]any{"identifier": "u-1", "type": typ}}
	ok, err := engine.matchesCondition(RuleCondition{Attribute: "context", Operator: "equal", Value: value}, ctx)
	if err != nil || !ok {
		t.Fatalf("expected context target match")
	}
}

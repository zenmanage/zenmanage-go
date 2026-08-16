package zenmanage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// RuleEngine evaluates flag rules against a context.
type RuleEngine struct{}

// NewRuleEngine creates a rule engine.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// Evaluate finds the first matching rule and returns its value envelope.
func (e *RuleEngine) Evaluate(rules []Rule, context Context) (*ValueEnvelope, error) {
	for i := range rules {
		matched, err := e.matchesRule(rules[i], context)
		if err != nil {
			return nil, err
		}
		if matched {
			return &rules[i].Value, nil
		}
	}
	return nil, nil
}

func (e *RuleEngine) matchesRule(rule Rule, context Context) (bool, error) {
	if len(rule.Clauses) == 0 && rule.Criteria == nil {
		return true, nil
	}
	if rule.Criteria != nil {
		return e.matchesCondition(*rule.Criteria, context)
	}
	for _, clause := range rule.Clauses {
		ok, err := e.matchesCondition(clause, context)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (e *RuleEngine) matchesCondition(condition RuleCondition, context Context) (bool, error) {
	op := normalizeOperator(condition.Operator)

	if condition.Attribute == "context" || condition.Attribute == "segment" {
		matched, err := matchContextTarget(condition.Value, context)
		if err != nil {
			return false, err
		}
		if op == "notequal" || op == "notin" {
			return !matched, nil
		}
		return matched, nil
	}

	attr, ok := context.GetAttribute(condition.Attribute)
	if !ok || len(attr.Values) == 0 {
		switch op {
		case "isnull", "notequal", "notin", "notcontains", "notstartswith", "notendswith":
			return true, nil
		default:
			return false, nil
		}
	}

	candidate := attr.Values[0].Value
	all := make([]string, 0, len(attr.Values))
	for _, v := range attr.Values {
		all = append(all, v.Value)
	}

	return applyOperator(op, candidate, all, condition.Value)
}

func normalizeOperator(op string) string {
	op = strings.TrimSpace(strings.ToLower(op))
	op = strings.ReplaceAll(op, "_", "")
	op = strings.ReplaceAll(op, "-", "")
	return op
}

func matchContextTarget(value any, context Context) (bool, error) {
	if context.Identifier() == "" {
		return false, nil
	}
	targets := extractContextTargets(value)
	if len(targets) == 0 {
		return false, nil
	}
	for _, target := range targets {
		if target.Identifier != context.Identifier() {
			continue
		}
		if target.Type == nil || *target.Type == "" || *target.Type == context.Type() {
			return true, nil
		}
	}
	return false, nil
}

func extractContextTargets(value any) []RuleContextTarget {
	result := []RuleContextTarget{}
	switch v := value.(type) {
	case map[string]any:
		t := toContextTarget(v)
		if t.Identifier != "" {
			result = append(result, t)
		}
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			t := toContextTarget(m)
			if t.Identifier != "" {
				result = append(result, t)
			}
		}
	}
	return result
}

func toContextTarget(m map[string]any) RuleContextTarget {
	t := RuleContextTarget{}
	if id, ok := m["identifier"].(string); ok {
		t.Identifier = id
	}
	if typ, ok := m["type"].(string); ok {
		t.Type = &typ
	}
	return t
}

func applyOperator(op, candidate string, allValues []string, rawValue any) (bool, error) {
	switch op {
	case "equal":
		return candidate == asString(rawValue), nil
	case "notequal":
		return candidate != asString(rawValue), nil
	case "in":
		targets := asStringSlice(rawValue)
		for _, value := range allValues {
			for _, target := range targets {
				if value == target {
					return true, nil
				}
			}
		}
		return false, nil
	case "notin":
		ok, _ := applyOperator("in", candidate, allValues, rawValue)
		return !ok, nil
	case "contains":
		substr := asString(rawValue)
		for _, v := range allValues {
			if strings.Contains(v, substr) {
				return true, nil
			}
		}
		return false, nil
	case "notcontains":
		ok, _ := applyOperator("contains", candidate, allValues, rawValue)
		return !ok, nil
	case "startswith":
		prefix := asString(rawValue)
		for _, v := range allValues {
			if strings.HasPrefix(v, prefix) {
				return true, nil
			}
		}
		return false, nil
	case "notstartswith":
		ok, _ := applyOperator("startswith", candidate, allValues, rawValue)
		return !ok, nil
	case "endswith":
		suffix := asString(rawValue)
		for _, v := range allValues {
			if strings.HasSuffix(v, suffix) {
				return true, nil
			}
		}
		return false, nil
	case "notendswith":
		ok, _ := applyOperator("endswith", candidate, allValues, rawValue)
		return !ok, nil
	case "gt", "gte", "lt", "lte":
		left, err := strconv.ParseFloat(candidate, 64)
		if err != nil {
			return false, nil
		}
		right, err := strconv.ParseFloat(asString(rawValue), 64)
		if err != nil {
			return false, nil
		}
		switch op {
		case "gt":
			return left > right, nil
		case "gte":
			return left >= right, nil
		case "lt":
			return left < right, nil
		default:
			return left <= right, nil
		}
	case "isnull":
		return candidate == "", nil
	case "notnull":
		return candidate != "", nil
	case "regex":
		rx, err := regexp.Compile(asString(rawValue))
		if err != nil {
			return false, &EvaluationError{Message: "invalid regex: " + err.Error()}
		}
		return rx.MatchString(candidate), nil
	default:
		return false, &EvaluationError{Message: fmt.Sprintf("unsupported operator: %s", op)}
	}
}

func asString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case []any:
		if len(x) > 0 {
			return asString(x[0])
		}
		return ""
	case []string:
		if len(x) > 0 {
			return x[0]
		}
		return ""
	default:
		return ""
	}
}

func asStringSlice(v any) []string {
	switch x := v.(type) {
	case []string:
		return append([]string{}, x...)
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			out = append(out, asString(item))
		}
		return out
	case string:
		return []string{x}
	default:
		return nil
	}
}

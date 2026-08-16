package zenmanage

import "fmt"

// ConfigurationError indicates invalid SDK configuration.
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

// EvaluationError indicates a rule or flag evaluation failure.
type EvaluationError struct {
	Message string
}

func (e *EvaluationError) Error() string {
	return e.Message
}

// FetchRulesError indicates a failure while loading rules from remote APIs.
type FetchRulesError struct {
	Message    string
	StatusCode int
}

func (e *FetchRulesError) Error() string {
	if e.StatusCode == 0 {
		return e.Message
	}
	return fmt.Sprintf("%s (status: %d)", e.Message, e.StatusCode)
}

// InvalidRulesError indicates malformed or semantically invalid rules payloads.
type InvalidRulesError struct {
	Message string
}

func (e *InvalidRulesError) Error() string {
	return e.Message
}

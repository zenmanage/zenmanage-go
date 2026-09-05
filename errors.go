package zenmanage

import "fmt"

// Error is implemented by every error type this SDK returns, mirroring the
// shared ZenmanageError base class in the JavaScript, PHP, and Python SDKs.
// Use errors.As to distinguish SDK errors from other errors:
//
//	var zmErr zenmanage.Error
//	if errors.As(err, &zmErr) { ... }
type Error interface {
	error
	zenmanageError()
}

// baseError is embedded by every concrete error type in this package so they
// all satisfy the Error interface above.
type baseError struct{}

func (baseError) zenmanageError() {}

// ConfigurationError indicates invalid SDK configuration.
type ConfigurationError struct {
	baseError
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

// EvaluationError indicates a rule or flag evaluation failure.
type EvaluationError struct {
	baseError
	Message string
}

func (e *EvaluationError) Error() string {
	return e.Message
}

// FetchRulesError indicates a failure while loading rules from remote APIs.
type FetchRulesError struct {
	baseError
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
	baseError
	Message string
}

func (e *InvalidRulesError) Error() string {
	return e.Message
}

package executor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/vkrishna03/orchestr/internal/models"
	"github.com/vkrishna03/orchestr/internal/parser"
)

// ResponseValidator validates HTTP responses against expectations.
type ResponseValidator struct {
	matcher *parser.PatternMatcher
	vars    *VariableStore
}

// NewResponseValidator creates a new response validator.
func NewResponseValidator(vars *VariableStore) *ResponseValidator {
	return &ResponseValidator{
		matcher: parser.NewPatternMatcher(),
		vars:    vars,
	}
}

// ResponseInfo contains details about the response for reporting.
type ResponseInfo struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

// ValidationError provides detailed information about validation failures.
type ValidationError struct {
	Field    string `json:"field,omitempty"`
	Expected any    `json:"expected"`
	Actual   any    `json:"actual"`
	Message  string `json:"message"`
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// Validate checks if a response matches expectations.
func (v *ResponseValidator) Validate(resp *http.Response, body []byte, expect models.ExpectSpec) error {
	// Check status code
	if resp.StatusCode != expect.Status {
		return &ValidationError{
			Expected: expect.Status,
			Actual:   resp.StatusCode,
			Message:  fmt.Sprintf("expected status %d, got %d", expect.Status, resp.StatusCode),
		}
	}

	// Check body pattern if specified
	if expect.Pattern != nil && len(expect.Pattern) > 0 {
		if err := v.validateBody(body, expect.Pattern); err != nil {
			return err
		}
	}

	return nil
}

// validateBody checks if the response body matches the expected pattern.
func (v *ResponseValidator) validateBody(body []byte, pattern map[string]any) error {
	if len(body) == 0 {
		return &ValidationError{
			Expected: pattern,
			Actual:   nil,
			Message:  "expected response body, got empty",
		}
	}

	var actual map[string]any
	if err := json.Unmarshal(body, &actual); err != nil {
		// Try as array
		var arr []any
		if err2 := json.Unmarshal(body, &arr); err2 == nil {
			// If pattern has a single key matching array pattern
			for key, patternVal := range pattern {
				if patternStr, ok := patternVal.(string); ok && strings.HasPrefix(patternStr, "<array") {
					if !v.matcher.MatchValue(patternStr, arr) {
						return &ValidationError{
							Field:    key,
							Expected: patternStr,
							Actual:   arr,
							Message:  fmt.Sprintf("array constraint mismatch for %q", key),
						}
					}
					return nil
				}
			}
		}
		return &ValidationError{
			Expected: "JSON object",
			Actual:   string(body),
			Message:  fmt.Sprintf("invalid JSON: %v", err),
		}
	}

	return v.matchPattern(pattern, actual, "")
}

// matchPattern recursively matches pattern against actual data.
func (v *ResponseValidator) matchPattern(pattern, actual map[string]any, prefix string) error {
	for key, expected := range pattern {
		fieldPath := key
		if prefix != "" {
			fieldPath = prefix + "." + key
		}

		// Check for optional fields
		actualKey, optional := parser.IsOptionalField(key)
		if optional {
			key = actualKey
			fieldPath = prefix
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += actualKey
		}

		actualVal, exists := actual[key]

		// Handle optional fields
		if !exists {
			if optional {
				continue // Optional field not present, that's okay
			}
			return &ValidationError{
				Field:    fieldPath,
				Expected: expected,
				Actual:   nil,
				Message:  fmt.Sprintf("field %q not found", fieldPath),
			}
		}

		if err := v.matchValue(expected, actualVal, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

// matchValue checks if a single value matches the expected pattern.
func (v *ResponseValidator) matchValue(expected, actual any, fieldPath string) error {
	switch exp := expected.(type) {
	case string:
		// Check for variable reference {{varName}}
		if varName, isRef := parser.IsVarRef(exp); isRef {
			// Resolve variable and compare
			if savedVal, ok := v.vars.Get(varName); ok {
				if !deepEqual(savedVal, actual) {
					return &ValidationError{
						Field:    fieldPath,
						Expected: fmt.Sprintf("{{%s}} = %v", varName, savedVal),
						Actual:   actual,
						Message:  fmt.Sprintf("variable %q mismatch", varName),
					}
				}
				return nil
			}
			// Variable not set, skip validation
			return nil
		}

		// Check for type pattern <type>
		if v.matcher.IsPattern(exp) {
			if !v.matcher.MatchValue(exp, actual) {
				return &ValidationError{
					Field:    fieldPath,
					Expected: exp,
					Actual:   actual,
					Message:  fmt.Sprintf("pattern %q mismatch", exp),
				}
			}
			return nil
		}

		// Literal string comparison
		actualStr, ok := actual.(string)
		if !ok {
			return &ValidationError{
				Field:    fieldPath,
				Expected: exp,
				Actual:   actual,
				Message:  "expected string value",
			}
		}
		if exp != actualStr {
			return &ValidationError{
				Field:    fieldPath,
				Expected: exp,
				Actual:   actual,
				Message:  "value mismatch",
			}
		}

	case map[string]any:
		// Nested object
		actualMap, ok := actual.(map[string]any)
		if !ok {
			return &ValidationError{
				Field:    fieldPath,
				Expected: exp,
				Actual:   actual,
				Message:  "expected object",
			}
		}
		if err := v.matchPattern(exp, actualMap, fieldPath); err != nil {
			return err
		}

	case []any:
		// Array validation
		actualArr, ok := actual.([]any)
		if !ok {
			return &ValidationError{
				Field:    fieldPath,
				Expected: exp,
				Actual:   actual,
				Message:  "expected array",
			}
		}
		if err := v.matchArray(exp, actualArr, fieldPath); err != nil {
			return err
		}

	case float64, int, bool:
		// Numeric/boolean comparison
		if !deepEqual(expected, actual) {
			return &ValidationError{
				Field:    fieldPath,
				Expected: expected,
				Actual:   actual,
				Message:  "value mismatch",
			}
		}

	default:
		return &ValidationError{
			Field:    fieldPath,
			Expected: expected,
			Actual:   actual,
			Message:  fmt.Sprintf("unsupported pattern type: %T", expected),
		}
	}

	return nil
}

// matchArray validates array values.
func (v *ResponseValidator) matchArray(expected, actual []any, fieldPath string) error {
	// If expected is empty, just check actual is an array (already done)
	if len(expected) == 0 {
		return nil
	}

	// If expected has a single pattern element, apply to all
	if len(expected) == 1 {
		pattern := expected[0]
		for i, item := range actual {
			itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			if err := v.matchValue(pattern, item, itemPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Otherwise, match element by element
	if len(actual) < len(expected) {
		return &ValidationError{
			Field:    fieldPath,
			Expected: fmt.Sprintf("array with at least %d elements", len(expected)),
			Actual:   fmt.Sprintf("array with %d elements", len(actual)),
			Message:  "array too short",
		}
	}

	for i, exp := range expected {
		itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := v.matchValue(exp, actual[i], itemPath); err != nil {
			return err
		}
	}

	return nil
}

// deepEqual compares two values for equality, handling type differences.
func deepEqual(a, b any) bool {
	// Handle numeric comparison (JSON numbers are float64)
	switch av := a.(type) {
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		}
	case int:
		switch bv := b.(type) {
		case float64:
			return float64(av) == bv
		case int:
			return av == bv
		}
	}
	return reflect.DeepEqual(a, b)
}

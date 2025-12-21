// Package parser implements the Orchestrfile DSL parser.
package parser

import (
	"errors"
	"fmt"
)

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	// Line is the line number where the error occurred (1-based).
	Line int

	// Column is the column number (optional, 0 if unknown).
	Column int

	// Message is a human-readable error description.
	Message string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Column > 0 {
		return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// Unwrap returns nil (ParseError is a leaf error).
func (e *ParseError) Unwrap() error {
	return nil
}

// ValidationError represents a semantic validation error.
type ValidationError struct {
	// TestIndex is the index of the test case (0-based).
	TestIndex int

	// Field is the field that failed validation.
	Field string

	// Message is the validation error message.
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("test %d: %s: %s", e.TestIndex+1, e.Field, e.Message)
}

// Sentinel errors for common parsing failures.
var (
	ErrEmptyFile       = errors.New("orchestrfile is empty")
	ErrNoTests         = errors.New("no test cases found")
	ErrInvalidJSON     = errors.New("invalid JSON syntax")
	ErrMissingEndpoint = errors.New("missing required ENDPOINT")
	ErrMissingMethod   = errors.New("missing required METHOD")
	ErrMissingExpect   = errors.New("missing required EXPECT")
)

// IsParseError checks if an error is a ParseError.
func IsParseError(err error) bool {
	var pe *ParseError
	return errors.As(err, &pe)
}

// GetLineNumber extracts the line number from a ParseError.
// Returns 0 and false if the error is not a ParseError.
func GetLineNumber(err error) (int, bool) {
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Line, true
	}
	return 0, false
}

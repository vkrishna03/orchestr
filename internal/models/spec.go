// Package models contains shared types used across the orchestr codebase.
package models

// TestSpec represents a single test case from an Orchestrfile.
type TestSpec struct {
	// Entity is the logical entity being tested (e.g., "User", "Payment").
	Entity string

	// Endpoint is the API path to call (required).
	Endpoint string

	// Method is the HTTP method (required): GET, POST, PUT, PATCH, DELETE.
	Method string

	// Data is the parsed request body as a map.
	Data map[string]any

	// RawData is the original DATA string for deferred variable substitution.
	RawData string

	// Auth is the authorization header value (e.g., "Bearer {{token}}").
	Auth string

	// Expect defines the expected response (required).
	Expect ExpectSpec

	// Save contains directives to save response values to variables.
	Save []SaveSpec

	// ExpectCalls defines expected mock service calls (for multi-service testing).
	ExpectCalls []ExpectCallSpec

	// LineNumber is the starting line of this test case (1-based, for errors).
	LineNumber int
}

// ExpectSpec represents the expected response for a test case.
type ExpectSpec struct {
	// Status is the expected HTTP status code.
	Status int

	// Pattern is the parsed response body pattern with type placeholders.
	Pattern map[string]any

	// RawBody is the original pattern string for reference.
	RawBody string
}

// SaveSpec represents a directive to save a response value to a variable.
type SaveSpec struct {
	// VarName is the variable name to save to.
	VarName string

	// JSONPath is the path to extract from the response (e.g., "body.token").
	JSONPath string
}

// ExpectCallSpec represents an expected call to a mock service.
type ExpectCallSpec struct {
	// Service is the mock service name (e.g., "stripe", "auth-service").
	Service string

	// Path is the expected request path.
	Path string

	// Method is the expected HTTP method (optional, empty means any).
	Method string

	// Count is the expected number of calls (default: 1).
	Count int
}

// TypePattern represents a type placeholder in EXPECT patterns.
type TypePattern string

// Supported type patterns for response validation.
const (
	PatternUUID    TypePattern = "<uuid>"
	PatternJWT     TypePattern = "<jwt>"
	PatternString  TypePattern = "<string>"
	PatternInt     TypePattern = "<int>"
	PatternFloat   TypePattern = "<float>"
	PatternBool    TypePattern = "<bool>"
	PatternEmail   TypePattern = "<email>"
	PatternISO8601 TypePattern = "<iso8601>"
	PatternAny     TypePattern = "<any>"
)

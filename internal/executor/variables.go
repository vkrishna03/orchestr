// Package executor provides test execution capabilities.
package executor

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// variablePattern matches {{variableName}} placeholders.
var variablePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// VariableStore manages test variables with thread-safe access.
type VariableStore struct {
	vars map[string]any
	mu   sync.RWMutex
}

// NewVariableStore creates a new variable store.
func NewVariableStore() *VariableStore {
	return &VariableStore{
		vars: make(map[string]any),
	}
}

// Set stores a variable value.
func (s *VariableStore) Set(name string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[name] = value
}

// Get retrieves a variable value.
func (s *VariableStore) Get(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.vars[name]
	return val, ok
}

// Substitute replaces {{variable}} placeholders in a string.
func (s *VariableStore) Substitute(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		name := match[2 : len(match)-2] // Extract name from {{name}}
		if val, ok := s.vars[name]; ok {
			return fmt.Sprint(val)
		}
		return match // Keep original if not found
	})
}

// SubstituteMap replaces variables in all string values of a map.
func (s *VariableStore) SubstituteMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	result := make(map[string]any, len(input))
	for k, v := range input {
		result[k] = s.substituteValue(v)
	}
	return result
}

// substituteValue recursively substitutes variables in a value.
func (s *VariableStore) substituteValue(val any) any {
	switch v := val.(type) {
	case string:
		return s.Substitute(v)
	case map[string]any:
		return s.SubstituteMap(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = s.substituteValue(item)
		}
		return result
	default:
		return val
	}
}

// ExtractAndSave extracts values from JSON body and saves them as variables.
// Path format: "body.field.nested" or just "field"
func (s *VariableStore) ExtractAndSave(body []byte, varName, jsonPath string) error {
	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	// Parse path and extract value
	value, err := extractJSONPath(data, jsonPath)
	if err != nil {
		return err
	}

	s.Set(varName, value)
	return nil
}

// extractJSONPath navigates a JSON structure using dot notation.
// Supports: "body.token", "data.user.id", "items.0.name"
func extractJSONPath(data any, path string) (any, error) {
	// Remove "body." prefix if present (it's implicit)
	path = strings.TrimPrefix(path, "body.")

	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found in path %q", part, path)
			}
			current = val

		case []any:
			// Try to parse as array index
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err != nil {
				return nil, fmt.Errorf("expected array index, got %q", part)
			}
			if idx < 0 || idx >= len(v) {
				return nil, fmt.Errorf("array index %d out of bounds (len=%d)", idx, len(v))
			}
			current = v[idx]

		default:
			return nil, fmt.Errorf("cannot navigate into %T at %q", current, part)
		}
	}

	return current, nil
}

// All returns a copy of all variables.
func (s *VariableStore) All() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any, len(s.vars))
	for k, v := range s.vars {
		result[k] = v
	}
	return result
}

// Clear removes all variables.
func (s *VariableStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars = make(map[string]any)
}

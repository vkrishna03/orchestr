package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// PatternMatcher validates values against type patterns like <uuid>, <jwt>, etc.
// Extended patterns support constraints: <array:min=1>, <int:gte=0>, <string:nonempty>
type PatternMatcher struct {
	patterns map[string]*regexp.Regexp
}

// NewPatternMatcher creates a matcher with built-in patterns.
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		patterns: map[string]*regexp.Regexp{
			"uuid":    regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
			"jwt":     regexp.MustCompile(`^eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*$`),
			"string":  regexp.MustCompile(`^.+$`),
			"int":     regexp.MustCompile(`^-?\d+$`),
			"float":   regexp.MustCompile(`^-?\d+(\.\d+)?$`),
			"bool":    regexp.MustCompile(`^(true|false)$`),
			"email":   regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
			"iso8601": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?)?$`),
			"url":     regexp.MustCompile(`^https?://[^\s]+$`),
			"any":     regexp.MustCompile(`.*`),
		},
	}
}

// Pattern represents a parsed pattern with type and constraints.
type Pattern struct {
	Type        string            // e.g., "uuid", "int", "array"
	Constraints map[string]string // e.g., {"min": "1", "gte": "0"}
	Optional    bool              // true if field is optional (? suffix)
	VarRef      string            // variable reference for {{varName}}
	Raw         string            // original pattern string
}

// patternRegex matches extended patterns like <type:constraint=value>
var extendedPatternRegex = regexp.MustCompile(`^<(\w+)(?::(.+))?>$`)
var constraintRegex = regexp.MustCompile(`(\w+)=([^,]+)`)
var varRefRegex = regexp.MustCompile(`^\{\{(\w+)\}\}$`)

// ParsePattern parses a pattern string into a Pattern struct.
// Supports: <uuid>, <int:gte=0>, <array:min=1,max=10>, <string:nonempty>, {{varName}}
func ParsePattern(s string) *Pattern {
	s = strings.TrimSpace(s)

	// Check for variable reference {{varName}}
	if matches := varRefRegex.FindStringSubmatch(s); matches != nil {
		return &Pattern{
			Type:   "varref",
			VarRef: matches[1],
			Raw:    s,
		}
	}

	// Check for extended pattern <type:constraints>
	matches := extendedPatternRegex.FindStringSubmatch(s)
	if matches == nil {
		// Not a pattern, return nil (literal value)
		return nil
	}

	p := &Pattern{
		Type:        matches[1],
		Constraints: make(map[string]string),
		Raw:         s,
	}

	// Parse constraints if present
	if matches[2] != "" {
		// Handle special case: <string:nonempty>
		if matches[2] == "nonempty" {
			p.Constraints["nonempty"] = "true"
		} else {
			// Parse key=value constraints
			for _, cm := range constraintRegex.FindAllStringSubmatch(matches[2], -1) {
				p.Constraints[cm[1]] = cm[2]
			}
		}
	}

	return p
}

// IsPattern checks if a string is a recognized type pattern.
func (m *PatternMatcher) IsPattern(s string) bool {
	p := ParsePattern(s)
	if p == nil {
		return false
	}
	if p.Type == "varref" {
		return true
	}
	// Check if base type is known
	_, ok := m.patterns[p.Type]
	if ok {
		return true
	}
	// Check for structural types
	return p.Type == "array" || p.Type == "object"
}

// Match validates a string value against a pattern.
// For complex types (array, object), use MatchValue instead.
func (m *PatternMatcher) Match(pattern string, value string) bool {
	// Handle patterns (angle brackets or variable refs)
	if strings.HasPrefix(pattern, "<") && strings.HasSuffix(pattern, ">") {
		p := ParsePattern(pattern)
		if p == nil {
			return pattern == value
		}
		return m.matchParsed(p, value)
	}

	// Handle variable references {{varName}}
	if strings.HasPrefix(pattern, "{{") && strings.HasSuffix(pattern, "}}") {
		// Variable references always match - actual value resolved at runtime
		return true
	}

	// Try as literal
	return pattern == value
}

// matchParsed validates a value against a parsed pattern.
func (m *PatternMatcher) matchParsed(p *Pattern, value string) bool {
	// Variable reference - always matches (resolved at runtime)
	if p.Type == "varref" {
		return true
	}

	// Check base type regex first
	if re, ok := m.patterns[p.Type]; ok {
		if !re.MatchString(value) {
			return false
		}
	}

	// Apply constraints
	return m.checkConstraints(p, value)
}

// checkConstraints validates constraints on a value.
func (m *PatternMatcher) checkConstraints(p *Pattern, value string) bool {
	for k, v := range p.Constraints {
		switch k {
		case "nonempty":
			if value == "" {
				return false
			}
		case "min", "minlen":
			minLen, _ := strconv.Atoi(v)
			if len(value) < minLen {
				return false
			}
		case "max", "maxlen":
			maxLen, _ := strconv.Atoi(v)
			if len(value) > maxLen {
				return false
			}
		case "len":
			exactLen, _ := strconv.Atoi(v)
			if len(value) != exactLen {
				return false
			}
		case "gt":
			// For numeric comparisons
			threshold, _ := strconv.ParseFloat(v, 64)
			val, err := strconv.ParseFloat(value, 64)
			if err != nil || val <= threshold {
				return false
			}
		case "gte":
			threshold, _ := strconv.ParseFloat(v, 64)
			val, err := strconv.ParseFloat(value, 64)
			if err != nil || val < threshold {
				return false
			}
		case "lt":
			threshold, _ := strconv.ParseFloat(v, 64)
			val, err := strconv.ParseFloat(value, 64)
			if err != nil || val >= threshold {
				return false
			}
		case "lte":
			threshold, _ := strconv.ParseFloat(v, 64)
			val, err := strconv.ParseFloat(value, 64)
			if err != nil || val > threshold {
				return false
			}
		}
	}
	return true
}

// MatchValue validates any value (including arrays/objects) against a pattern.
func (m *PatternMatcher) MatchValue(pattern string, value any) bool {
	p := ParsePattern(pattern)
	if p == nil {
		// Literal comparison
		if strVal, ok := value.(string); ok {
			return pattern == strVal
		}
		return false
	}

	switch p.Type {
	case "array":
		return m.matchArray(p, value)
	case "object":
		return m.matchObject(p, value)
	case "varref":
		return true // Variable references matched at runtime
	default:
		// Convert to string for regex match
		if strVal, ok := value.(string); ok {
			return m.matchParsed(p, strVal)
		}
		// Handle numeric types
		switch v := value.(type) {
		case float64:
			return m.matchParsed(p, strconv.FormatFloat(v, 'f', -1, 64))
		case int:
			return m.matchParsed(p, strconv.Itoa(v))
		case bool:
			return m.matchParsed(p, strconv.FormatBool(v))
		}
		return false
	}
}

// matchArray validates an array value against array pattern constraints.
func (m *PatternMatcher) matchArray(p *Pattern, value any) bool {
	arr, ok := value.([]any)
	if !ok {
		return false
	}

	length := len(arr)

	for k, v := range p.Constraints {
		switch k {
		case "min":
			minLen, _ := strconv.Atoi(v)
			if length < minLen {
				return false
			}
		case "max":
			maxLen, _ := strconv.Atoi(v)
			if length > maxLen {
				return false
			}
		case "len":
			exactLen, _ := strconv.Atoi(v)
			if length != exactLen {
				return false
			}
		case "nonempty":
			if length == 0 {
				return false
			}
		}
	}

	return true
}

// matchObject validates an object value against object pattern.
func (m *PatternMatcher) matchObject(p *Pattern, value any) bool {
	_, ok := value.(map[string]any)
	if !ok {
		return false
	}

	// Object constraints could be added here (e.g., required keys)
	// For now, just check it's a valid object

	for k, v := range p.Constraints {
		switch k {
		case "nonempty":
			obj := value.(map[string]any)
			if v == "true" && len(obj) == 0 {
				return false
			}
		}
	}

	return true
}

// SupportedPatterns returns all supported pattern type names.
func (m *PatternMatcher) SupportedPatterns() []string {
	patterns := make([]string, 0, len(m.patterns)+2)
	for p := range m.patterns {
		patterns = append(patterns, "<"+p+">")
	}
	// Add structural types
	patterns = append(patterns, "<array>", "<object>")
	return patterns
}

// IsOptionalField checks if a field name ends with ? indicating optional.
func IsOptionalField(fieldName string) (string, bool) {
	if strings.HasSuffix(fieldName, "?") {
		return strings.TrimSuffix(fieldName, "?"), true
	}
	return fieldName, false
}

// IsVarRef checks if a pattern is a variable reference {{varName}}.
func IsVarRef(s string) (string, bool) {
	if matches := varRefRegex.FindStringSubmatch(s); matches != nil {
		return matches[1], true
	}
	return "", false
}

// DefaultMatcher is a shared instance of PatternMatcher with built-in patterns.
var DefaultMatcher = NewPatternMatcher()

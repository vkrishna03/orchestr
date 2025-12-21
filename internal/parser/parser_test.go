package parser

import (
	"fmt"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSpecs int
		wantErr   bool
		errLine   int // Expected error line, if wantErr
	}{
		{
			name: "basic test case",
			input: `
ENTITY: User
ENDPOINT: /api/users
METHOD: GET
EXPECT: 200 with {id: <uuid>}
`,
			wantSpecs: 1,
		},
		{
			name: "multiple test cases",
			input: `
ENDPOINT: /api/health
METHOD: GET
EXPECT: 200

---

ENDPOINT: /api/users
METHOD: POST
DATA: {name: "test"}
EXPECT: 201
`,
			wantSpecs: 2,
		},
		{
			name: "missing endpoint",
			input: `
METHOD: GET
EXPECT: 200
`,
			wantErr: true,
			errLine: 2,
		},
		{
			name: "missing method",
			input: `
ENDPOINT: /api/users
EXPECT: 200
`,
			wantErr: true,
			errLine: 2,
		},
		{
			name: "missing expect",
			input: `
ENDPOINT: /api/users
METHOD: GET
`,
			wantErr: true,
			errLine: 2,
		},
		{
			name: "invalid method",
			input: `
ENDPOINT: /api/users
METHOD: GETT
EXPECT: 200
`,
			wantErr: true,
			errLine: 3,
		},
		{
			name: "with comments",
			input: `
# This is a comment
ENDPOINT: /api/users
# Another comment
METHOD: GET
EXPECT: 200
`,
			wantSpecs: 1,
		},
		{
			name: "with save directive",
			input: `
ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@example.com", password: "secret"}
EXPECT: 200 with {token: <jwt>}
SAVE: token from body.token
`,
			wantSpecs: 1,
		},
		{
			name: "with auth header",
			input: `
ENDPOINT: /api/me
METHOD: GET
AUTH: Bearer {{token}}
EXPECT: 200
`,
			wantSpecs: 1,
		},
		{
			name: "empty file",
			input: "",
			wantSpecs: 0,
		},
		{
			name: "only comments",
			input: `
# comment 1
# comment 2
`,
			wantSpecs: 0,
		},
		{
			name: "status only expect",
			input: `
ENDPOINT: /api/health
METHOD: GET
EXPECT: 204
`,
			wantSpecs: 1,
		},
		{
			name: "multiple saves",
			input: `
ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@example.com"}
EXPECT: 200 with {token: <jwt>, userId: <uuid>}
SAVE: token from body.token
SAVE: userId from body.userId
`,
			wantSpecs: 1,
		},
		{
			name: "with expect_calls",
			input: `
ENDPOINT: /api/payments
METHOD: POST
DATA: {amount: 100}
EXPECT: 200
EXPECT_CALLS: stripe:/v1/charges
`,
			wantSpecs: 1,
		},
		{
			name: "lowercase method normalized",
			input: `
ENDPOINT: /api/users
METHOD: get
EXPECT: 200
`,
			wantSpecs: 1,
		},
		{
			name: "standard JSON format",
			input: `
ENDPOINT: /api/users
METHOD: POST
DATA: {"name": "test", "age": 30}
EXPECT: 201
`,
			wantSpecs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			specs, err := p.ParseString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errLine > 0 {
					line, ok := GetLineNumber(err)
					if !ok {
						t.Errorf("expected ParseError with line number, got %T", err)
					} else if line != tt.errLine {
						t.Errorf("expected error on line %d, got line %d", tt.errLine, line)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(specs) != tt.wantSpecs {
				t.Errorf("expected %d specs, got %d", tt.wantSpecs, len(specs))
			}
		})
	}
}

func TestParser_ParseExpect(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStatus int
		wantBody   bool
		wantErr    bool
	}{
		{"status only", "200", 200, false, false},
		{"status with pattern", "201 with {id: <uuid>}", 201, true, false},
		{"no body pattern", "204", 204, false, false},
		{"500 error", "500 with {error: <string>}", 500, true, false},
		{"invalid status", "abc", 0, false, true},
		{"invalid format", "200 using {}", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseExpect(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec.Status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, spec.Status)
			}
			if (spec.Pattern != nil) != tt.wantBody {
				t.Errorf("expected body=%v, got pattern=%v", tt.wantBody, spec.Pattern)
			}
		})
	}
}

func TestParser_ParseSave(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantVar  string
		wantPath string
		wantErr  bool
	}{
		{"simple", "token from body.token", "token", "body.token", false},
		{"nested path", "userId from body.data.user.id", "userId", "body.data.user.id", false},
		{"underscore var", "auth_token from body.auth.token", "auth_token", "body.auth.token", false},
		{"invalid format", "token body.token", "", "", true},
		{"missing path", "token from", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseSave(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec.VarName != tt.wantVar {
				t.Errorf("expected var %q, got %q", tt.wantVar, spec.VarName)
			}
			if spec.JSONPath != tt.wantPath {
				t.Errorf("expected path %q, got %q", tt.wantPath, spec.JSONPath)
			}
		})
	}
}

func TestParser_ParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"standard JSON", `{"name": "test"}`, false},
		{"relaxed JSON", `{name: "test"}`, false},
		{"nested object", `{user: {name: "test", age: 30}}`, false},
		{"with array", `{tags: ["a", "b"]}`, false},
		{"mixed quotes", `{name: "test", "age": 30}`, false},
		{"invalid JSON", `{name: }`, true},
		{"unclosed brace", `{name: "test"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseJSON(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_Strict(t *testing.T) {
	input := `
ENDPOINT: /api/users
METHOD: GET
UNKNOWN: something
EXPECT: 200
`

	// Non-strict mode: should succeed
	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("non-strict mode should succeed: %v", err)
	}
	if len(specs) != 1 {
		t.Errorf("expected 1 spec, got %d", len(specs))
	}

	// Strict mode: should fail
	pStrict := NewParser(WithStrict(true))
	_, err = pStrict.ParseString(input)
	if err == nil {
		t.Fatal("strict mode should fail on unknown keyword")
	}
}

func TestPatternMatcher_Match(t *testing.T) {
	m := NewPatternMatcher()

	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"<uuid>", "550e8400-e29b-41d4-a716-446655440000", true},
		{"<uuid>", "not-a-uuid", false},
		{"<uuid>", "550e8400e29b41d4a716446655440000", false}, // No dashes
		{"<jwt>", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", true},
		{"<jwt>", "not.a.jwt", false},
		{"<jwt>", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig", true},
		{"<int>", "42", true},
		{"<int>", "-42", true},
		{"<int>", "42.5", false},
		{"<int>", "abc", false},
		{"<float>", "3.14", true},
		{"<float>", "-2.5", true},
		{"<float>", "42", true}, // Integers are valid floats
		{"<float>", "abc", false},
		{"<bool>", "true", true},
		{"<bool>", "false", true},
		{"<bool>", "TRUE", false}, // Case sensitive
		{"<email>", "test@example.com", true},
		{"<email>", "user.name+tag@domain.co.uk", true},
		{"<email>", "invalid", false},
		{"<email>", "@domain.com", false},
		{"<string>", "anything", true},
		{"<string>", "", false}, // Empty string doesn't match <string>
		{"<any>", "", true},
		{"<any>", "anything", true},
		{"<iso8601>", "2024-12-21", true},
		{"<iso8601>", "2024-12-21T14:30:00Z", true},
		{"<iso8601>", "2024-12-21T14:30:00+05:30", true},
		{"<iso8601>", "not-a-date", false},
		// Literal match for unknown patterns
		{"exact value", "exact value", true},
		{"exact value", "different value", false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/%s", tt.pattern, tt.value)
		t.Run(name, func(t *testing.T) {
			got := m.Match(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestPatternMatcher_IsPattern(t *testing.T) {
	m := NewPatternMatcher()

	tests := []struct {
		input string
		want  bool
	}{
		{"<uuid>", true},
		{"<jwt>", true},
		{"<string>", true},
		{"<int>", true},
		{"<float>", true},
		{"<bool>", true},
		{"<email>", true},
		{"<iso8601>", true},
		{"<any>", true},
		{"<unknown>", false},
		{"uuid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := m.IsPattern(tt.input)
			if got != tt.want {
				t.Errorf("IsPattern(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsParseError(t *testing.T) {
	parseErr := &ParseError{Line: 1, Message: "test"}
	otherErr := fmt.Errorf("not a parse error")

	if !IsParseError(parseErr) {
		t.Error("IsParseError should return true for ParseError")
	}
	if IsParseError(otherErr) {
		t.Error("IsParseError should return false for non-ParseError")
	}
	if IsParseError(nil) {
		t.Error("IsParseError should return false for nil")
	}
}

func TestParser_ComplexJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "nested objects",
			input:   `{user: {profile: {name: "test", settings: {theme: "dark"}}}}`,
			wantErr: false,
		},
		{
			name:    "array of objects",
			input:   `{users: [{name: "alice"}, {name: "bob"}]}`,
			wantErr: false,
		},
		{
			name:    "mixed types",
			input:   `{name: "test", age: 30, active: true, score: 95.5, tags: ["a", "b"]}`,
			wantErr: false,
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   `{items: []}`,
			wantErr: false,
		},
		{
			name:    "null value",
			input:   `{value: null}`,
			wantErr: false,
		},
		{
			name:    "special characters in string",
			input:   `{msg: "hello \"world\" \n\t"}`,
			wantErr: false,
		},
		{
			name:    "unicode",
			input:   `{name: "日本語"}`,
			wantErr: false,
		},
		{
			name:    "numbers",
			input:   `{int: 42, neg: -10, float: 3.14, exp: 1e10}`,
			wantErr: false,
		},
		{
			name:    "deeply nested",
			input:   `{a: {b: {c: {d: {e: "deep"}}}}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseJSON(tt.input)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			input := fmt.Sprintf(`
ENDPOINT: /api/test
METHOD: %s
EXPECT: 200
`, method)
			p := NewParser()
			specs, err := p.ParseString(input)
			if err != nil {
				t.Fatalf("unexpected error for method %s: %v", method, err)
			}
			if len(specs) != 1 {
				t.Fatalf("expected 1 spec, got %d", len(specs))
			}
			if specs[0].Method != method {
				t.Errorf("expected method %s, got %s", method, specs[0].Method)
			}
		})
	}
}

func TestParser_MultipleTypePatterns(t *testing.T) {
	input := `
ENDPOINT: /api/users
METHOD: POST
DATA: {email: "test@example.com", password: "secret123"}
EXPECT: 201 with {id: <uuid>, email: <email>, token: <jwt>, createdAt: <iso8601>, verified: <bool>}
`
	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}

	pattern := specs[0].Expect.Pattern
	expectedKeys := []string{"id", "email", "token", "createdAt", "verified"}
	for _, key := range expectedKeys {
		if _, ok := pattern[key]; !ok {
			t.Errorf("expected key %q in pattern", key)
		}
	}
}

func TestParser_VariableSubstitution(t *testing.T) {
	input := `
ENDPOINT: /api/users/{{userId}}
METHOD: GET
AUTH: Bearer {{token}}
EXPECT: 200 with {id: <uuid>}
`
	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if specs[0].Endpoint != "/api/users/{{userId}}" {
		t.Errorf("expected endpoint with variable, got %s", specs[0].Endpoint)
	}
	if specs[0].Auth != "Bearer {{token}}" {
		t.Errorf("expected auth with variable, got %s", specs[0].Auth)
	}
}

func TestParser_ExpectCallsFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantService string
		wantPath    string
		wantMethod  string
		wantCount   int
	}{
		{
			name:        "basic",
			input:       "stripe:/v1/charges",
			wantService: "stripe",
			wantPath:    "/v1/charges",
			wantMethod:  "",
			wantCount:   1,
		},
		{
			name:        "with method",
			input:       "stripe:/v1/charges:POST",
			wantService: "stripe",
			wantPath:    "/v1/charges",
			wantMethod:  "POST",
			wantCount:   1,
		},
		{
			name:        "with count",
			input:       "stripe:/v1/charges:POST:2",
			wantService: "stripe",
			wantPath:    "/v1/charges",
			wantMethod:  "POST",
			wantCount:   2,
		},
		{
			name:        "service with dash",
			input:       "auth-service:/api/validate",
			wantService: "auth-service",
			wantPath:    "/api/validate",
			wantMethod:  "",
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseExpectCall(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec.Service != tt.wantService {
				t.Errorf("service: got %q, want %q", spec.Service, tt.wantService)
			}
			if spec.Path != tt.wantPath {
				t.Errorf("path: got %q, want %q", spec.Path, tt.wantPath)
			}
			if spec.Method != tt.wantMethod {
				t.Errorf("method: got %q, want %q", spec.Method, tt.wantMethod)
			}
			if spec.Count != tt.wantCount {
				t.Errorf("count: got %d, want %d", spec.Count, tt.wantCount)
			}
		})
	}
}

func TestParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSpecs int
		wantErr   bool
	}{
		{
			name: "trailing separator",
			input: `
ENDPOINT: /api/test
METHOD: GET
EXPECT: 200
---
`,
			wantSpecs: 1,
		},
		{
			name: "multiple separators",
			input: `
ENDPOINT: /api/a
METHOD: GET
EXPECT: 200
---
---
ENDPOINT: /api/b
METHOD: GET
EXPECT: 200
`,
			wantSpecs: 2,
		},
		{
			name: "whitespace lines",
			input: `

ENDPOINT: /api/test

METHOD: GET

EXPECT: 200

`,
			wantSpecs: 1,
		},
		{
			name: "endpoint with query params",
			input: `
ENDPOINT: /api/users?page=1&limit=10
METHOD: GET
EXPECT: 200
`,
			wantSpecs: 1,
		},
		{
			name: "all status codes",
			input: `
ENDPOINT: /api/test
METHOD: GET
EXPECT: 418
`,
			wantSpecs: 1,
		},
		{
			name: "empty entity",
			input: `
ENTITY:
ENDPOINT: /api/test
METHOD: GET
EXPECT: 200
`,
			wantSpecs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			specs, err := p.ParseString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(specs) != tt.wantSpecs {
				t.Errorf("expected %d specs, got %d", tt.wantSpecs, len(specs))
			}
		})
	}
}

func TestParser_FullWorkflow(t *testing.T) {
	// Simulates a complete auth flow
	input := `
# Health check
ENTITY: System
ENDPOINT: /health
METHOD: GET
EXPECT: 200

---

# Register new user
ENTITY: User
ENDPOINT: /api/register
METHOD: POST
DATA: {email: "test@example.com", password: "secret123", name: "Test User"}
EXPECT: 201 with {id: <uuid>, email: <email>}
SAVE: userId from body.id

---

# Login
ENTITY: User
ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@example.com", password: "secret123"}
EXPECT: 200 with {token: <jwt>, expiresIn: <int>}
SAVE: token from body.token

---

# Get profile with auth
ENTITY: User
ENDPOINT: /api/users/{{userId}}
METHOD: GET
AUTH: Bearer {{token}}
EXPECT: 200 with {id: <uuid>, email: <email>, name: <string>}

---

# Update profile
ENTITY: User
ENDPOINT: /api/users/{{userId}}
METHOD: PUT
AUTH: Bearer {{token}}
DATA: {name: "Updated Name"}
EXPECT: 200 with {id: <uuid>, name: <string>}

---

# Delete account
ENTITY: User
ENDPOINT: /api/users/{{userId}}
METHOD: DELETE
AUTH: Bearer {{token}}
EXPECT: 204
`

	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(specs) != 6 {
		t.Fatalf("expected 6 specs, got %d", len(specs))
	}

	// Verify each spec
	expected := []struct {
		entity   string
		method   string
		endpoint string
		status   int
	}{
		{"System", "GET", "/health", 200},
		{"User", "POST", "/api/register", 201},
		{"User", "POST", "/api/login", 200},
		{"User", "GET", "/api/users/{{userId}}", 200},
		{"User", "PUT", "/api/users/{{userId}}", 200},
		{"User", "DELETE", "/api/users/{{userId}}", 204},
	}

	for i, exp := range expected {
		spec := specs[i]
		if spec.Entity != exp.entity {
			t.Errorf("spec %d: entity got %q, want %q", i, spec.Entity, exp.entity)
		}
		if spec.Method != exp.method {
			t.Errorf("spec %d: method got %q, want %q", i, spec.Method, exp.method)
		}
		if spec.Endpoint != exp.endpoint {
			t.Errorf("spec %d: endpoint got %q, want %q", i, spec.Endpoint, exp.endpoint)
		}
		if spec.Expect.Status != exp.status {
			t.Errorf("spec %d: status got %d, want %d", i, spec.Expect.Status, exp.status)
		}
	}

	// Verify saves
	if len(specs[1].Save) != 1 || specs[1].Save[0].VarName != "userId" {
		t.Error("spec 1: expected SAVE userId")
	}
	if len(specs[2].Save) != 1 || specs[2].Save[0].VarName != "token" {
		t.Error("spec 2: expected SAVE token")
	}
}

func TestParser_JSONPatternValues(t *testing.T) {
	// Verify that type patterns are stored correctly as strings
	input := `
ENDPOINT: /api/test
METHOD: GET
EXPECT: 200 with {id: <uuid>, count: <int>, price: <float>, active: <bool>}
`
	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pattern := specs[0].Expect.Pattern

	tests := []struct {
		key  string
		want string
	}{
		{"id", "<uuid>"},
		{"count", "<int>"},
		{"price", "<float>"},
		{"active", "<bool>"},
	}

	for _, tt := range tests {
		got, ok := pattern[tt.key].(string)
		if !ok {
			t.Errorf("pattern[%q] is not a string", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("pattern[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// ============================================================
// Extended Pattern Tests
// ============================================================

func TestParsePattern(t *testing.T) {
	tests := []struct {
		input       string
		wantType    string
		wantConstr  map[string]string
		wantVarRef  string
		wantNil     bool
	}{
		// Basic types
		{"<uuid>", "uuid", nil, "", false},
		{"<int>", "int", nil, "", false},
		{"<string>", "string", nil, "", false},
		{"<array>", "array", nil, "", false},
		{"<object>", "object", nil, "", false},

		// With constraints
		{"<int:gte=0>", "int", map[string]string{"gte": "0"}, "", false},
		{"<int:gt=0>", "int", map[string]string{"gt": "0"}, "", false},
		{"<int:lt=100>", "int", map[string]string{"lt": "100"}, "", false},
		{"<int:lte=100>", "int", map[string]string{"lte": "100"}, "", false},
		{"<float:gte=0.0>", "float", map[string]string{"gte": "0.0"}, "", false},
		{"<string:nonempty>", "string", map[string]string{"nonempty": "true"}, "", false},
		{"<string:min=3>", "string", map[string]string{"min": "3"}, "", false},
		{"<string:max=50>", "string", map[string]string{"max": "50"}, "", false},
		{"<array:min=1>", "array", map[string]string{"min": "1"}, "", false},
		{"<array:max=10>", "array", map[string]string{"max": "10"}, "", false},
		{"<array:len=5>", "array", map[string]string{"len": "5"}, "", false},
		{"<object:nonempty>", "object", map[string]string{"nonempty": "true"}, "", false},

		// Multiple constraints
		{"<int:gte=0,lte=100>", "int", map[string]string{"gte": "0", "lte": "100"}, "", false},
		{"<array:min=1,max=10>", "array", map[string]string{"min": "1", "max": "10"}, "", false},

		// Variable references
		{"{{userId}}", "varref", nil, "userId", false},
		{"{{token}}", "varref", nil, "token", false},

		// Not patterns (literals)
		{"hello", "", nil, "", true},
		{"123", "", nil, "", true},
		{"", "", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := ParsePattern(tt.input)

			if tt.wantNil {
				if p != nil {
					t.Errorf("expected nil, got %+v", p)
				}
				return
			}

			if p == nil {
				t.Fatalf("expected pattern, got nil")
			}

			if p.Type != tt.wantType {
				t.Errorf("type: got %q, want %q", p.Type, tt.wantType)
			}

			if tt.wantVarRef != "" && p.VarRef != tt.wantVarRef {
				t.Errorf("varref: got %q, want %q", p.VarRef, tt.wantVarRef)
			}

			for k, v := range tt.wantConstr {
				if p.Constraints[k] != v {
					t.Errorf("constraint %q: got %q, want %q", k, p.Constraints[k], v)
				}
			}
		})
	}
}

func TestPatternMatcher_ExtendedMatch(t *testing.T) {
	m := NewPatternMatcher()

	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		// Basic patterns still work
		{"<uuid>", "550e8400-e29b-41d4-a716-446655440000", true},
		{"<int>", "42", true},
		{"<string>", "hello", true},

		// String constraints
		{"<string:nonempty>", "hello", true},
		{"<string:nonempty>", "", false},
		{"<string:min=3>", "hello", true},
		{"<string:min=3>", "hi", false},
		{"<string:max=5>", "hello", true},
		{"<string:max=5>", "hello world", false},
		{"<string:len=5>", "hello", true},
		{"<string:len=5>", "hi", false},

		// Numeric constraints
		{"<int:gte=0>", "0", true},
		{"<int:gte=0>", "42", true},
		{"<int:gte=0>", "-1", false},
		{"<int:gt=0>", "1", true},
		{"<int:gt=0>", "0", false},
		{"<int:lt=100>", "99", true},
		{"<int:lt=100>", "100", false},
		{"<int:lte=100>", "100", true},
		{"<int:lte=100>", "101", false},

		// Range constraints
		{"<int:gte=0,lte=100>", "50", true},
		{"<int:gte=0,lte=100>", "-1", false},
		{"<int:gte=0,lte=100>", "101", false},

		// Float constraints
		{"<float:gte=0.0>", "3.14", true},
		{"<float:gt=0>", "0.01", true},
		{"<float:gt=0>", "0", false},

		// URL pattern
		{"<url>", "https://example.com", true},
		{"<url>", "http://localhost:8080/api", true},
		{"<url>", "not-a-url", false},

		// Variable references (always match - resolved at runtime)
		{"{{userId}}", "anything", true},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/%s", tt.pattern, tt.value)
		t.Run(name, func(t *testing.T) {
			got := m.Match(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestPatternMatcher_MatchValue(t *testing.T) {
	m := NewPatternMatcher()

	tests := []struct {
		pattern string
		value   any
		want    bool
	}{
		// Array matching
		{"<array>", []any{1, 2, 3}, true},
		{"<array>", []any{}, true},
		{"<array>", "not an array", false},
		{"<array:min=1>", []any{1}, true},
		{"<array:min=1>", []any{}, false},
		{"<array:max=2>", []any{1, 2}, true},
		{"<array:max=2>", []any{1, 2, 3}, false},
		{"<array:len=3>", []any{1, 2, 3}, true},
		{"<array:len=3>", []any{1, 2}, false},
		{"<array:nonempty>", []any{1}, true},
		{"<array:nonempty>", []any{}, false},
		{"<array:min=1,max=5>", []any{1, 2, 3}, true},
		{"<array:min=1,max=5>", []any{}, false},

		// Object matching
		{"<object>", map[string]any{"key": "value"}, true},
		{"<object>", map[string]any{}, true},
		{"<object>", "not an object", false},
		{"<object:nonempty>", map[string]any{"key": "value"}, true},
		{"<object:nonempty>", map[string]any{}, false},

		// Numeric values (from JSON)
		{"<int>", float64(42), true},
		{"<int:gte=0>", float64(42), true},
		{"<int:gte=0>", float64(-1), false},
		{"<float>", float64(3.14), true},

		// Boolean values
		{"<bool>", true, true},
		{"<bool>", false, true},

		// String values
		{"<string>", "hello", true},
		{"<uuid>", "550e8400-e29b-41d4-a716-446655440000", true},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s/%v", tt.pattern, tt.value)
		t.Run(name, func(t *testing.T) {
			got := m.MatchValue(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("MatchValue(%q, %v) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestIsOptionalField(t *testing.T) {
	tests := []struct {
		input      string
		wantName   string
		wantOptional bool
	}{
		{"name", "name", false},
		{"name?", "name", true},
		{"nickname?", "nickname", true},
		{"id", "id", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, optional := IsOptionalField(tt.input)
			if name != tt.wantName {
				t.Errorf("name: got %q, want %q", name, tt.wantName)
			}
			if optional != tt.wantOptional {
				t.Errorf("optional: got %v, want %v", optional, tt.wantOptional)
			}
		})
	}
}

func TestIsVarRef(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantOk   bool
	}{
		{"{{userId}}", "userId", true},
		{"{{token}}", "token", true},
		{"{{auth_token}}", "auth_token", true},
		{"<uuid>", "", false},
		{"userId", "", false},
		{"{userId}", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, ok := IsVarRef(tt.input)
			if name != tt.wantName || ok != tt.wantOk {
				t.Errorf("IsVarRef(%q) = (%q, %v), want (%q, %v)", tt.input, name, ok, tt.wantName, tt.wantOk)
			}
		})
	}
}

func TestParser_ExtendedPatterns(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSpecs int
		wantErr   bool
	}{
		{
			name: "array pattern",
			input: `
ENDPOINT: /api/users
METHOD: GET
EXPECT: 200 with {users: <array>, count: <int>}
`,
			wantSpecs: 1,
		},
		{
			name: "array with constraints",
			input: `
ENDPOINT: /api/users
METHOD: GET
EXPECT: 200 with {users: <array:min=1>, total: <int:gte=0>}
`,
			wantSpecs: 1,
		},
		{
			name: "object pattern",
			input: `
ENDPOINT: /api/user
METHOD: GET
EXPECT: 200 with {user: <object>, meta: <object:nonempty>}
`,
			wantSpecs: 1,
		},
		{
			name: "variable reference in pattern",
			input: `
ENDPOINT: /api/users/{{userId}}
METHOD: GET
EXPECT: 200 with {id: {{userId}}, name: <string>}
`,
			wantSpecs: 1,
		},
		{
			name: "string constraints",
			input: `
ENDPOINT: /api/validate
METHOD: POST
DATA: {name: "test"}
EXPECT: 200 with {name: <string:nonempty>, code: <string:len=6>}
`,
			wantSpecs: 1,
		},
		{
			name: "numeric constraints",
			input: `
ENDPOINT: /api/products
METHOD: GET
EXPECT: 200 with {price: <float:gt=0>, quantity: <int:gte=0,lte=1000>}
`,
			wantSpecs: 1,
		},
		{
			name: "url pattern",
			input: `
ENDPOINT: /api/links
METHOD: GET
EXPECT: 200 with {url: <url>}
`,
			wantSpecs: 1,
		},
		{
			name: "complex response with all pattern types",
			input: `
ENDPOINT: /api/order
METHOD: POST
DATA: {items: [{id: 1}]}
EXPECT: 201 with {id: <uuid>, items: <array:min=1>, total: <float:gte=0>, status: <string:nonempty>, createdAt: <iso8601>}
`,
			wantSpecs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			specs, err := p.ParseString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(specs) != tt.wantSpecs {
				t.Errorf("expected %d specs, got %d", tt.wantSpecs, len(specs))
			}
		})
	}
}

func TestParser_ExtendedPatternValues(t *testing.T) {
	input := `
ENDPOINT: /api/test
METHOD: GET
EXPECT: 200 with {items: <array:min=1>, count: <int:gte=0>, id: {{savedId}}}
`
	p := NewParser()
	specs, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pattern := specs[0].Expect.Pattern

	// Verify patterns are stored correctly
	tests := []struct {
		key  string
		want string
	}{
		{"items", "<array:min=1>"},
		{"count", "<int:gte=0>"},
		{"id", "{{savedId}}"},
	}

	for _, tt := range tests {
		got, ok := pattern[tt.key].(string)
		if !ok {
			t.Errorf("pattern[%q] is not a string", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("pattern[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

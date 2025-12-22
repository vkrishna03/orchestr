package executor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vkrishna03/orchestr/internal/models"
)

func TestVariableStore_SetGet(t *testing.T) {
	store := NewVariableStore()

	store.Set("token", "abc123")
	store.Set("userId", 42)

	val, ok := store.Get("token")
	if !ok || val != "abc123" {
		t.Errorf("Get(token) = %v, %v; want abc123, true", val, ok)
	}

	val, ok = store.Get("userId")
	if !ok || val != 42 {
		t.Errorf("Get(userId) = %v, %v; want 42, true", val, ok)
	}

	_, ok = store.Get("missing")
	if ok {
		t.Error("Get(missing) should return false")
	}
}

func TestVariableStore_Substitute(t *testing.T) {
	store := NewVariableStore()
	store.Set("userId", "abc-123")
	store.Set("token", "jwt-xyz")

	tests := []struct {
		input    string
		expected string
	}{
		{"/api/users/{{userId}}", "/api/users/abc-123"},
		{"Bearer {{token}}", "Bearer jwt-xyz"},
		{"{{userId}}/{{token}}", "abc-123/jwt-xyz"},
		{"no variables here", "no variables here"},
		{"{{missing}}", "{{missing}}"}, // Keep unresolved
	}

	for _, tt := range tests {
		result := store.Substitute(tt.input)
		if result != tt.expected {
			t.Errorf("Substitute(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestVariableStore_SubstituteMap(t *testing.T) {
	store := NewVariableStore()
	store.Set("email", "test@test.com")

	input := map[string]any{
		"email":    "{{email}}",
		"password": "secret",
		"nested": map[string]any{
			"ref": "{{email}}",
		},
	}

	result := store.SubstituteMap(input)

	if result["email"] != "test@test.com" {
		t.Errorf("email = %v; want test@test.com", result["email"])
	}
	if result["password"] != "secret" {
		t.Errorf("password = %v; want secret", result["password"])
	}

	nested := result["nested"].(map[string]any)
	if nested["ref"] != "test@test.com" {
		t.Errorf("nested.ref = %v; want test@test.com", nested["ref"])
	}
}

func TestVariableStore_ExtractAndSave(t *testing.T) {
	store := NewVariableStore()

	body := []byte(`{
		"token": "jwt-abc",
		"user": {
			"id": "user-123",
			"name": "Test"
		},
		"items": [{"id": "item-1"}, {"id": "item-2"}]
	}`)

	tests := []struct {
		varName  string
		jsonPath string
		expected any
	}{
		{"token", "body.token", "jwt-abc"},
		{"token2", "token", "jwt-abc"},
		{"userId", "user.id", "user-123"},
		{"firstItem", "items.0.id", "item-1"},
	}

	for _, tt := range tests {
		err := store.ExtractAndSave(body, tt.varName, tt.jsonPath)
		if err != nil {
			t.Errorf("ExtractAndSave(%s, %s) error: %v", tt.varName, tt.jsonPath, err)
			continue
		}

		val, ok := store.Get(tt.varName)
		if !ok || val != tt.expected {
			t.Errorf("After save, Get(%s) = %v; want %v", tt.varName, val, tt.expected)
		}
	}
}

func TestVariableStore_ExtractAndSave_Errors(t *testing.T) {
	store := NewVariableStore()

	tests := []struct {
		name     string
		body     []byte
		jsonPath string
	}{
		{"empty body", []byte{}, "token"},
		{"invalid json", []byte("not json"), "token"},
		{"missing field", []byte(`{"other": 1}`), "token"},
		{"invalid array index", []byte(`{"items": []}`), "items.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.ExtractAndSave(tt.body, "var", tt.jsonPath)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestRequestBuilder_Build(t *testing.T) {
	vars := NewVariableStore()
	vars.Set("userId", "abc-123")
	vars.Set("token", "jwt-xyz")

	builder := NewRequestBuilder(vars)

	spec := models.TestSpec{
		Method:   "POST",
		Endpoint: "/api/users/{{userId}}",
		Data: map[string]any{
			"name": "Test",
		},
		Auth: "Bearer {{token}}",
	}

	req, info, err := builder.Build(spec, "http://localhost:8080")
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	// Check URL
	expectedURL := "http://localhost:8080/api/users/abc-123"
	if req.URL.String() != expectedURL {
		t.Errorf("URL = %s; want %s", req.URL, expectedURL)
	}

	// Check method
	if req.Method != "POST" {
		t.Errorf("Method = %s; want POST", req.Method)
	}

	// Check auth header
	auth := req.Header.Get("Authorization")
	if auth != "Bearer jwt-xyz" {
		t.Errorf("Authorization = %s; want Bearer jwt-xyz", auth)
	}

	// Check info
	if info.Method != "POST" {
		t.Errorf("info.Method = %s; want POST", info.Method)
	}
	if info.Headers["Authorization"] != "Bearer jwt-xyz" {
		t.Errorf("info.Headers[Authorization] = %s; want Bearer jwt-xyz", info.Headers["Authorization"])
	}
}

func TestResponseValidator_Validate_StatusCode(t *testing.T) {
	vars := NewVariableStore()
	validator := NewResponseValidator(vars)

	tests := []struct {
		name         string
		statusCode   int
		expectStatus int
		wantErr      bool
	}{
		{"exact match", 200, 200, false},
		{"mismatch", 401, 200, true},
		{"201 created", 201, 201, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			expect := models.ExpectSpec{Status: tt.expectStatus}

			err := validator.Validate(resp, nil, expect)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResponseValidator_Validate_Pattern(t *testing.T) {
	vars := NewVariableStore()
	validator := NewResponseValidator(vars)

	tests := []struct {
		name    string
		body    string
		pattern map[string]any
		wantErr bool
	}{
		{
			name: "literal match",
			body: `{"status": "ok"}`,
			pattern: map[string]any{
				"status": "ok",
			},
			wantErr: false,
		},
		{
			name: "literal mismatch",
			body: `{"status": "error"}`,
			pattern: map[string]any{
				"status": "ok",
			},
			wantErr: true,
		},
		{
			name: "uuid pattern",
			body: `{"id": "550e8400-e29b-41d4-a716-446655440000"}`,
			pattern: map[string]any{
				"id": "<uuid>",
			},
			wantErr: false,
		},
		{
			name: "invalid uuid",
			body: `{"id": "not-a-uuid"}`,
			pattern: map[string]any{
				"id": "<uuid>",
			},
			wantErr: true,
		},
		{
			name: "nested object",
			body: `{"user": {"id": "123", "name": "Test"}}`,
			pattern: map[string]any{
				"user": map[string]any{
					"id":   "<string>",
					"name": "Test",
				},
			},
			wantErr: false,
		},
		{
			name: "optional field present",
			body: `{"id": "123", "meta": "data"}`,
			pattern: map[string]any{
				"id":    "<string>",
				"meta?": "<string>",
			},
			wantErr: false,
		},
		{
			name: "optional field missing",
			body: `{"id": "123"}`,
			pattern: map[string]any{
				"id":    "<string>",
				"meta?": "<string>",
			},
			wantErr: false,
		},
		{
			name: "required field missing",
			body: `{"id": "123"}`,
			pattern: map[string]any{
				"id":   "<string>",
				"name": "<string>",
			},
			wantErr: true,
		},
		{
			name: "int constraint",
			body: `{"count": 5}`,
			pattern: map[string]any{
				"count": "<int:gte=0>",
			},
			wantErr: false,
		},
		{
			name: "array constraint",
			body: `{"items": [1, 2, 3]}`,
			pattern: map[string]any{
				"items": "<array:min=1>",
			},
			wantErr: false,
		},
		{
			name: "empty array fails min",
			body: `{"items": []}`,
			pattern: map[string]any{
				"items": "<array:min=1>",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: 200}
			expect := models.ExpectSpec{
				Status:  200,
				Pattern: tt.pattern,
			}

			err := validator.Validate(resp, []byte(tt.body), expect)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResponseValidator_Validate_VariableRef(t *testing.T) {
	vars := NewVariableStore()
	vars.Set("savedId", "abc-123")
	validator := NewResponseValidator(vars)

	tests := []struct {
		name    string
		body    string
		pattern map[string]any
		wantErr bool
	}{
		{
			name: "variable matches",
			body: `{"id": "abc-123"}`,
			pattern: map[string]any{
				"id": "{{savedId}}",
			},
			wantErr: false,
		},
		{
			name: "variable mismatch",
			body: `{"id": "different"}`,
			pattern: map[string]any{
				"id": "{{savedId}}",
			},
			wantErr: true,
		},
		{
			name: "unset variable passes",
			body: `{"id": "anything"}`,
			pattern: map[string]any{
				"id": "{{unknownVar}}",
			},
			wantErr: false, // Unset variables pass
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: 200}
			expect := models.ExpectSpec{
				Status:  200,
				Pattern: tt.pattern,
			}

			err := validator.Validate(resp, []byte(tt.body), expect)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExecutor_Run(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/users":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":    "user-123",
					"email": "test@test.com",
				})
				return
			}
		case "/api/users/user-123":
			if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":    "user-123",
					"email": "test@test.com",
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create executor
	opts := DefaultOptions()
	opts.BaseURL = server.URL
	exec := New(opts)

	// Define tests
	specs := []models.TestSpec{
		{
			Entity:   "User",
			Method:   "POST",
			Endpoint: "/api/users",
			Data: map[string]any{
				"email": "test@test.com",
			},
			Expect: models.ExpectSpec{
				Status: 201,
				Pattern: map[string]any{
					"id":    "<string>",
					"email": "test@test.com",
				},
			},
			Save: []models.SaveSpec{
				{VarName: "userId", JSONPath: "id"},
			},
		},
		{
			Entity:   "User",
			Method:   "GET",
			Endpoint: "/api/users/{{userId}}",
			Expect: models.ExpectSpec{
				Status: 200,
				Pattern: map[string]any{
					"id": "{{userId}}",
				},
			},
		},
	}

	// Run tests
	results, summary, err := exec.Run(context.Background(), specs)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Check summary
	if summary.Total != 2 {
		t.Errorf("summary.Total = %d; want 2", summary.Total)
	}
	if summary.Passed != 2 {
		t.Errorf("summary.Passed = %d; want 2", summary.Passed)
	}
	if summary.Failed != 0 {
		t.Errorf("summary.Failed = %d; want 0", summary.Failed)
	}

	// Check individual results
	for i, r := range results {
		if !r.Passed {
			t.Errorf("results[%d].Passed = false; error: %s", i, r.Error)
		}
	}

	// Check variable was saved and used
	val, ok := exec.Variables().Get("userId")
	if !ok || val != "user-123" {
		t.Errorf("Variables().Get(userId) = %v, %v; want user-123, true", val, ok)
	}
}

func TestExecutor_Run_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "unauthorized",
		})
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.BaseURL = server.URL
	exec := New(opts)

	specs := []models.TestSpec{
		{
			Method:   "GET",
			Endpoint: "/api/protected",
			Expect: models.ExpectSpec{
				Status: 200,
			},
		},
	}

	results, summary, err := exec.Run(context.Background(), specs)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if summary.Passed != 0 {
		t.Errorf("summary.Passed = %d; want 0", summary.Passed)
	}
	if summary.Failed != 1 {
		t.Errorf("summary.Failed = %d; want 1", summary.Failed)
	}
	if results[0].Passed {
		t.Error("results[0].Passed = true; want false")
	}
	if results[0].Error == "" {
		t.Error("results[0].Error is empty; want error message")
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	opts := DefaultOptions()
	opts.BaseURL = server.URL
	opts.Timeout = 10 * time.Millisecond
	exec := New(opts)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	specs := []models.TestSpec{
		{
			Method:   "GET",
			Endpoint: "/api/slow",
			Expect:   models.ExpectSpec{Status: 200},
		},
	}

	_, _, err := exec.Run(ctx, specs)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:    "user.id",
		Expected: "<uuid>",
		Actual:   "not-uuid",
		Message:  "pattern mismatch",
	}

	expected := "user.id: pattern mismatch"
	if err.Error() != expected {
		t.Errorf("Error() = %q; want %q", err.Error(), expected)
	}

	// Without field
	err2 := &ValidationError{
		Message: "status mismatch",
	}
	if err2.Error() != "status mismatch" {
		t.Errorf("Error() = %q; want %q", err2.Error(), "status mismatch")
	}
}

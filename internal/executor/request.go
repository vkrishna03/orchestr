package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vkrishna03/orchestr/internal/models"
)

// RequestBuilder constructs HTTP requests from test specs.
type RequestBuilder struct {
	vars *VariableStore
}

// NewRequestBuilder creates a new request builder.
func NewRequestBuilder(vars *VariableStore) *RequestBuilder {
	return &RequestBuilder{vars: vars}
}

// RequestInfo contains details about the built request for reporting.
type RequestInfo struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// Build creates an HTTP request from a test spec.
func (b *RequestBuilder) Build(spec models.TestSpec, baseURL string) (*http.Request, RequestInfo, error) {
	info := RequestInfo{
		Method:  spec.Method,
		Headers: make(map[string]string),
	}

	// Build URL with variable substitution
	endpoint := b.vars.Substitute(spec.Endpoint)
	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(endpoint, "/")
	info.URL = url

	// Build request body
	var body []byte
	if spec.Data != nil || spec.RawData != "" {
		data := b.prepareData(spec)
		var err error
		body, err = json.Marshal(data)
		if err != nil {
			return nil, info, fmt.Errorf("marshal request body: %w", err)
		}
		info.Body = string(body)
	}

	// Create request
	req, err := http.NewRequest(spec.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, info, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
		info.Headers["Content-Type"] = "application/json"
	}

	req.Header.Set("Accept", "application/json")
	info.Headers["Accept"] = "application/json"

	// Add trace header for debugging
	traceID := uuid.New().String()
	req.Header.Set("X-Request-ID", traceID)
	info.Headers["X-Request-ID"] = traceID

	// Set auth header with variable substitution
	if spec.Auth != "" {
		auth := b.vars.Substitute(spec.Auth)
		req.Header.Set("Authorization", auth)
		info.Headers["Authorization"] = auth
	}

	return req, info, nil
}

// prepareData processes request data with variable substitution.
func (b *RequestBuilder) prepareData(spec models.TestSpec) map[string]any {
	// Use parsed Data if available
	if spec.Data != nil {
		return b.vars.SubstituteMap(spec.Data)
	}

	// Fall back to parsing RawData with substitution
	if spec.RawData != "" {
		substituted := b.vars.Substitute(spec.RawData)
		var data map[string]any
		if err := json.Unmarshal([]byte(substituted), &data); err != nil {
			// Return as-is if not valid JSON
			return map[string]any{"_raw": substituted}
		}
		return data
	}

	return nil
}

package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vkrishna03/orchestr/internal/models"
)

// Executor runs test specs against a service.
type Executor struct {
	client    *http.Client
	baseURL   string
	vars      *VariableStore
	builder   *RequestBuilder
	validator *ResponseValidator
}

// Options configures the executor.
type Options struct {
	// BaseURL is the target service URL (e.g., "http://localhost:8080").
	BaseURL string

	// Timeout is the maximum time for each request.
	Timeout time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		BaseURL: "http://localhost:8080",
		Timeout: 30 * time.Second,
	}
}

// New creates a new executor with the given options.
func New(opts Options) *Executor {
	vars := NewVariableStore()
	return &Executor{
		client: &http.Client{
			Timeout: opts.Timeout,
		},
		baseURL:   opts.BaseURL,
		vars:      vars,
		builder:   NewRequestBuilder(vars),
		validator: NewResponseValidator(vars),
	}
}

// TestResult contains the outcome of a single test.
type TestResult struct {
	Spec     models.TestSpec `json:"spec"`
	Passed   bool            `json:"passed"`
	Duration time.Duration   `json:"duration"`
	Error    string          `json:"error,omitempty"`
	Request  RequestInfo     `json:"request"`
	Response ResponseInfo    `json:"response"`
}

// Summary contains aggregate test results.
type Summary struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Duration time.Duration `json:"duration"`
}

// Run executes all test specs and returns results.
func (e *Executor) Run(ctx context.Context, specs []models.TestSpec) ([]TestResult, Summary, error) {
	results := make([]TestResult, 0, len(specs))
	summary := Summary{Total: len(specs)}
	start := time.Now()

	for _, spec := range specs {
		select {
		case <-ctx.Done():
			return results, summary, ctx.Err()
		default:
		}

		result := e.RunOne(ctx, spec)
		results = append(results, result)

		if result.Passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}

	summary.Duration = time.Since(start)
	return results, summary, nil
}

// RunOne executes a single test spec.
func (e *Executor) RunOne(ctx context.Context, spec models.TestSpec) TestResult {
	result := TestResult{
		Spec: spec,
	}
	start := time.Now()

	// Build request
	req, reqInfo, err := e.builder.Build(spec, e.baseURL)
	if err != nil {
		result.Error = fmt.Sprintf("build request: %v", err)
		result.Request = reqInfo
		result.Duration = time.Since(start)
		return result
	}
	result.Request = reqInfo

	// Execute request
	req = req.WithContext(ctx)
	resp, err := e.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("execute request: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("read response: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Response = ResponseInfo{
		StatusCode: resp.StatusCode,
		Headers:    extractHeaders(resp.Header),
		Body:       string(body),
	}

	// Validate response
	if err := e.validator.Validate(resp, body, spec.Expect); err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// Process SAVE directives
	for _, save := range spec.Save {
		if err := e.vars.ExtractAndSave(body, save.VarName, save.JSONPath); err != nil {
			result.Error = fmt.Sprintf("save %q: %v", save.VarName, err)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// Variables returns the variable store for inspection.
func (e *Executor) Variables() *VariableStore {
	return e.vars
}

// SetVariable sets a variable for use in subsequent tests.
func (e *Executor) SetVariable(name string, value any) {
	e.vars.Set(name, value)
}

// extractHeaders converts http.Header to a simple map.
func extractHeaders(h http.Header) map[string]string {
	result := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

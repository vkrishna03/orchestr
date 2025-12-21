package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vkrishna03/orchestr/internal/executor"
	"github.com/vkrishna03/orchestr/internal/models"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"console", FormatConsole},
		{"json", FormatJSON},
		{"markdown", FormatMarkdown},
		{"md", FormatMarkdown},
		{"invalid", FormatConsole},
		{"", FormatConsole},
	}

	for _, tt := range tests {
		result := ParseFormat(tt.input)
		if result != tt.expected {
			t.Errorf("ParseFormat(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []Format{FormatConsole, FormatJSON, FormatMarkdown}

	for _, format := range tests {
		rep := New(format)
		if rep == nil {
			t.Errorf("New(%q) returned nil", format)
		}
	}

	// Verify correct types
	if _, ok := New(FormatConsole).(*ConsoleReporter); !ok {
		t.Error("FormatConsole should return *ConsoleReporter")
	}
	if _, ok := New(FormatJSON).(*JSONReporter); !ok {
		t.Error("FormatJSON should return *JSONReporter")
	}
	if _, ok := New(FormatMarkdown).(*MarkdownReporter); !ok {
		t.Error("FormatMarkdown should return *MarkdownReporter")
	}
}

func sampleResults() ([]executor.TestResult, executor.Summary, *models.Service) {
	results := []executor.TestResult{
		{
			Spec: models.TestSpec{
				Entity:   "Health",
				Method:   "GET",
				Endpoint: "/health",
				Expect:   models.ExpectSpec{Status: 200},
			},
			Passed:   true,
			Duration: 25 * time.Millisecond,
			Request:  executor.RequestInfo{Method: "GET", URL: "http://localhost:8080/health"},
			Response: executor.ResponseInfo{StatusCode: 200, Body: `{"status":"ok"}`},
		},
		{
			Spec: models.TestSpec{
				Entity:   "Auth",
				Method:   "POST",
				Endpoint: "/auth/login",
				Expect:   models.ExpectSpec{Status: 200},
			},
			Passed:   false,
			Duration: 50 * time.Millisecond,
			Error:    "expected status 200, got 401",
			Request:  executor.RequestInfo{Method: "POST", URL: "http://localhost:8080/auth/login"},
			Response: executor.ResponseInfo{StatusCode: 401, Body: `{"error":"unauthorized"}`},
		},
	}

	summary := executor.Summary{
		Total:    2,
		Passed:   1,
		Failed:   1,
		Duration: 75 * time.Millisecond,
	}

	service := &models.Service{
		Name: "test-service",
		Type: models.ServiceSpringBoot,
		Port: 8080,
		Config: map[string]string{
			"profile": "local",
		},
	}

	return results, summary, service
}

func TestConsoleReporter_Report(t *testing.T) {
	results, summary, service := sampleResults()
	rep := &ConsoleReporter{NoColor: true}

	var buf bytes.Buffer
	err := rep.Report(&buf, results, summary, service)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	output := buf.String()

	// Check key elements
	checks := []string{
		"test-service",
		"Running 2 test(s)",
		"✓",
		"✗",
		"[Health]",
		"[Auth]",
		"1/2 passed",
		"expected status 200, got 401",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q", check)
		}
	}
}

func TestJSONReporter_Report(t *testing.T) {
	results, summary, service := sampleResults()
	rep := &JSONReporter{Pretty: false}

	var buf bytes.Buffer
	err := rep.Report(&buf, results, summary, service)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	// Parse JSON
	var report JSONReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check structure
	if report.Service == nil {
		t.Error("service is nil")
	} else if report.Service.Name != "test-service" {
		t.Errorf("service.name = %q; want test-service", report.Service.Name)
	}

	if report.Summary.Total != 2 {
		t.Errorf("summary.total = %d; want 2", report.Summary.Total)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("summary.passed = %d; want 1", report.Summary.Passed)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("summary.failed = %d; want 1", report.Summary.Failed)
	}

	if len(report.Results) != 2 {
		t.Errorf("len(results) = %d; want 2", len(report.Results))
	}

	// Check first result
	if !report.Results[0].Passed {
		t.Error("results[0].passed = false; want true")
	}
	if report.Results[1].Passed {
		t.Error("results[1].passed = true; want false")
	}
	if report.Results[1].Error == "" {
		t.Error("results[1].error is empty")
	}
}

func TestMarkdownReporter_Report(t *testing.T) {
	results, summary, service := sampleResults()
	rep := &MarkdownReporter{}

	var buf bytes.Buffer
	err := rep.Report(&buf, results, summary, service)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	output := buf.String()

	// Check Markdown structure
	checks := []string{
		"# Orchestr Test Report",
		"## Service",
		"## Summary",
		"## Results",
		"## Failures",
		"| test-service |",
		"1/2 passed",
		"✅",
		"❌",
		"`/health`",
		"`/auth/login`",
		"expected status 200, got 401",
		"<details>",
		"</details>",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q", check)
		}
	}
}

func TestConsoleReporter_NoService(t *testing.T) {
	results, summary, _ := sampleResults()
	rep := &ConsoleReporter{NoColor: true}

	var buf bytes.Buffer
	err := rep.Report(&buf, results, summary, nil)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	// Should work without service info
	if !strings.Contains(buf.String(), "Running 2 test(s)") {
		t.Error("output missing test count")
	}
}

func TestJSONReporter_Pretty(t *testing.T) {
	results, summary, service := sampleResults()
	rep := &JSONReporter{Pretty: true}

	var buf bytes.Buffer
	err := rep.Report(&buf, results, summary, service)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}

	// Pretty output should have indentation
	if !strings.Contains(buf.String(), "  ") {
		t.Error("pretty output should have indentation")
	}
}

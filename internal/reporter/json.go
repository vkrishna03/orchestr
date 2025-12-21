package reporter

import (
	"encoding/json"
	"io"
	"time"

	"github.com/vkrishna03/orchestr/internal/executor"
	"github.com/vkrishna03/orchestr/internal/models"
)

// JSONReporter outputs results as JSON.
type JSONReporter struct {
	Pretty bool // Pretty print with indentation
}

// JSONReport is the structured output format.
type JSONReport struct {
	Timestamp string          `json:"timestamp"`
	Service   *ServiceInfo    `json:"service,omitempty"`
	Summary   SummaryInfo     `json:"summary"`
	Results   []JSONResult    `json:"results"`
}

// ServiceInfo contains service metadata.
type ServiceInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Port    int    `json:"port"`
	Profile string `json:"profile,omitempty"`
}

// SummaryInfo contains test summary.
type SummaryInfo struct {
	Total    int    `json:"total"`
	Passed   int    `json:"passed"`
	Failed   int    `json:"failed"`
	PassRate int    `json:"passRate"`
	Duration string `json:"duration"`
}

// JSONResult represents a single test result.
type JSONResult struct {
	Entity   string              `json:"entity"`
	Method   string              `json:"method"`
	Endpoint string              `json:"endpoint"`
	Passed   bool                `json:"passed"`
	Duration string              `json:"duration"`
	Expected ExpectedInfo        `json:"expected"`
	Actual   *ActualInfo         `json:"actual,omitempty"`
	Error    string              `json:"error,omitempty"`
	Request  executor.RequestInfo  `json:"request,omitempty"`
	Response executor.ResponseInfo `json:"response,omitempty"`
}

// ExpectedInfo contains expected values.
type ExpectedInfo struct {
	Status  int            `json:"status"`
	Pattern map[string]any `json:"pattern,omitempty"`
}

// ActualInfo contains actual values.
type ActualInfo struct {
	Status int    `json:"status"`
	Body   string `json:"body,omitempty"`
}

// Report writes JSON output.
func (r *JSONReporter) Report(w io.Writer, results []executor.TestResult, summary executor.Summary, service *models.Service) error {
	report := JSONReport{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Summary: SummaryInfo{
			Total:    summary.Total,
			Passed:   summary.Passed,
			Failed:   summary.Failed,
			PassRate: calculatePassRate(summary),
			Duration: summary.Duration.String(),
		},
		Results: make([]JSONResult, 0, len(results)),
	}

	// Add service info
	if service != nil {
		report.Service = &ServiceInfo{
			Name:    service.Name,
			Type:    string(service.Type),
			Port:    service.Port,
			Profile: service.Config["profile"],
		}
	}

	// Convert results
	for _, result := range results {
		entity := result.Spec.Entity
		if entity == "" {
			entity = "Test"
		}

		jr := JSONResult{
			Entity:   entity,
			Method:   result.Spec.Method,
			Endpoint: result.Spec.Endpoint,
			Passed:   result.Passed,
			Duration: result.Duration.String(),
			Expected: ExpectedInfo{
				Status:  result.Spec.Expect.Status,
				Pattern: result.Spec.Expect.Pattern,
			},
			Request:  result.Request,
			Response: result.Response,
		}

		if !result.Passed {
			jr.Error = result.Error
			jr.Actual = &ActualInfo{
				Status: result.Response.StatusCode,
				Body:   result.Response.Body,
			}
		}

		report.Results = append(report.Results, jr)
	}

	// Encode
	encoder := json.NewEncoder(w)
	if r.Pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(report)
}

func calculatePassRate(summary executor.Summary) int {
	if summary.Total == 0 {
		return 0
	}
	return (summary.Passed * 100) / summary.Total
}

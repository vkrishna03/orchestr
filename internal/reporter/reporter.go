// Package reporter provides test result output in various formats.
package reporter

import (
	"io"

	"github.com/vkrishna03/orchestr/internal/executor"
	"github.com/vkrishna03/orchestr/internal/models"
)

// Reporter outputs test results in a specific format.
type Reporter interface {
	// Report writes the test results to the given writer.
	Report(w io.Writer, results []executor.TestResult, summary executor.Summary, service *models.Service) error
}

// Format represents the output format type.
type Format string

const (
	FormatConsole  Format = "console"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// New creates a reporter for the given format.
func New(format Format) Reporter {
	switch format {
	case FormatJSON:
		return &JSONReporter{}
	case FormatMarkdown:
		return &MarkdownReporter{}
	default:
		return &ConsoleReporter{}
	}
}

// ParseFormat converts a string to a Format, defaulting to console.
func ParseFormat(s string) Format {
	switch s {
	case "json":
		return FormatJSON
	case "markdown", "md":
		return FormatMarkdown
	default:
		return FormatConsole
	}
}

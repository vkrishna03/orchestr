package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/vkrishna03/orchestr/internal/executor"
	"github.com/vkrishna03/orchestr/internal/models"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// ConsoleReporter outputs results to console with colors.
type ConsoleReporter struct {
	NoColor bool // Disable colors for non-TTY output
}

// Report writes colored console output.
func (r *ConsoleReporter) Report(w io.Writer, results []executor.TestResult, summary executor.Summary, service *models.Service) error {
	// Service info
	if service != nil {
		fmt.Fprintf(w, "\n%sService:%s %s (%s) on port %d\n",
			r.color(colorCyan), r.color(colorReset),
			service.Name, service.Type, service.Port)
	}

	// Results header
	fmt.Fprintf(w, "\n%sRunning %d test(s):%s\n\n", r.color(colorBold), summary.Total, r.color(colorReset))

	// Individual results
	for _, result := range results {
		r.printResult(w, result)
	}

	// Summary
	fmt.Fprintln(w)
	r.printSummary(w, summary)

	return nil
}

// printResult outputs a single test result.
func (r *ConsoleReporter) printResult(w io.Writer, result executor.TestResult) {
	entity := result.Spec.Entity
	if entity == "" {
		entity = "Test"
	}

	var status, statusColor string
	if result.Passed {
		status = "✓"
		statusColor = colorGreen
	} else {
		status = "✗"
		statusColor = colorRed
	}

	fmt.Fprintf(w, "  %s%s%s [%s] %s %s → %d %s(%s)%s\n",
		r.color(statusColor),
		status,
		r.color(colorReset),
		entity,
		result.Spec.Method,
		result.Spec.Endpoint,
		result.Spec.Expect.Status,
		r.color(colorGray),
		formatDuration(result.Duration),
		r.color(colorReset),
	)

	// Show error for failed tests
	if !result.Passed && result.Error != "" {
		fmt.Fprintf(w, "    %s%s%s\n", r.color(colorRed), result.Error, r.color(colorReset))
	}
}

// printSummary outputs the test summary.
func (r *ConsoleReporter) printSummary(w io.Writer, summary executor.Summary) {
	passRate := 0
	if summary.Total > 0 {
		passRate = (summary.Passed * 100) / summary.Total
	}

	var summaryColor string
	if summary.Failed == 0 {
		summaryColor = colorGreen
	} else if summary.Passed == 0 {
		summaryColor = colorRed
	} else {
		summaryColor = colorYellow
	}

	fmt.Fprintf(w, "%s%s%d/%d passed (%d%%)%s in %s\n",
		r.color(colorBold),
		r.color(summaryColor),
		summary.Passed,
		summary.Total,
		passRate,
		r.color(colorReset),
		formatDuration(summary.Duration),
	)

	// List failures
	if summary.Failed > 0 {
		fmt.Fprintf(w, "\n%s%d test(s) failed%s\n", r.color(colorRed), summary.Failed, r.color(colorReset))
	}
}

// color returns the color code if colors are enabled.
func (r *ConsoleReporter) color(code string) string {
	if r.NoColor {
		return ""
	}
	return code
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

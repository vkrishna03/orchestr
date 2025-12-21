package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vkrishna03/orchestr/internal/discovery"
	"github.com/vkrishna03/orchestr/internal/executor"
	"github.com/vkrishna03/orchestr/internal/parser"
	"github.com/vkrishna03/orchestr/internal/reporter"
	"github.com/vkrishna03/orchestr/internal/runner"
)

var testCmd = &cobra.Command{
	Use:   "test [path]",
	Short: "Run tests against services",
	Long: `Run functional tests defined in Orchestrfile against your services.

Examples:
  orchestr test .
  orchestr test ./my-project
  orchestr test . --format json
  orchestr test . --spec ./tests/Orchestrfile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringP("format", "f", "console", "output format (console, json, markdown)")
	testCmd.Flags().StringP("spec", "s", "", "path to Orchestrfile (default: ./Orchestrfile)")
	testCmd.Flags().Bool("no-start", false, "skip starting the service (assume already running)")

	_ = viper.BindPFlag("format", testCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("spec", testCmd.Flags().Lookup("spec"))
	_ = viper.BindPFlag("no-start", testCmd.Flags().Lookup("no-start"))
}

func runTest(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Get config from Viper (flags, env, or config file)
	format := viper.GetString("format")
	specPath := viper.GetString("spec")
	noStart := viper.GetBool("no-start")

	if specPath == "" {
		specPath = filepath.Join(path, "Orchestrfile")
	}

	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("orchestrfile not found: %s", specPath)
	}

	slog.Info("starting test run",
		"path", path,
		"spec", specPath,
		"format", format,
	)

	// Parse Orchestrfile
	p := parser.NewParser()
	specs, err := p.ParseFile(specPath)
	if err != nil {
		return fmt.Errorf("parse orchestrfile: %w", err)
	}

	slog.Info("parsed orchestrfile", "tests", len(specs))

	// Discover service
	discoverer := discovery.NewDiscoverer()
	svc, err := discoverer.DiscoverOne(path)
	if err != nil {
		return fmt.Errorf("discover service: %w", err)
	}

	if svc == nil {
		slog.Warn("no service detected, running tests against existing server")
		noStart = true
	} else {
		slog.Info("discovered service",
			"name", svc.Name,
			"type", svc.Type,
			"port", svc.Port,
		)
	}

	// Set up context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("received shutdown signal")
		cancel()
	}()

	// Start service if needed
	var svcRunner runner.Runner
	if svc != nil && !noStart {
		opts := runner.DefaultOptions()
		opts.Verbose = verbose

		svcRunner = runner.NewNativeRunner(opts)

		fmt.Printf("Starting %s...", svc.Name)
		if err := svcRunner.Start(ctx, svc); err != nil {
			fmt.Println(" failed")
			slog.Error("failed to start service", "error", err)
			if logs := svcRunner.Logs(); logs != "" {
				slog.Debug("service logs", "logs", logs)
			}
			return fmt.Errorf("start service: %w", err)
		}
		fmt.Println(" ready")

		// Ensure cleanup on exit
		defer func() {
			fmt.Printf("Stopping %s...", svc.Name)
			if err := svcRunner.Stop(context.Background()); err != nil {
				slog.Error("failed to stop service", "error", err)
			}
			fmt.Println(" done")
		}()
	}

	// Determine base URL
	port := 8080
	if svc != nil && svc.Port > 0 {
		port = svc.Port
	}
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Create executor
	execOpts := executor.DefaultOptions()
	execOpts.BaseURL = baseURL
	exec := executor.New(execOpts)

	// Execute tests
	results, summary, err := exec.Run(ctx, specs)
	if err != nil {
		return fmt.Errorf("run tests: %w", err)
	}

	// Create reporter based on format
	rep := reporter.New(reporter.ParseFormat(format))

	// For markdown, write to file
	if format == "markdown" || format == "md" {
		reportPath := filepath.Join(path, "orchestr-report.md")
		f, err := os.Create(reportPath)
		if err != nil {
			return fmt.Errorf("create report file: %w", err)
		}
		defer f.Close()

		if err := rep.Report(f, results, summary, svc); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		fmt.Printf("Report written to: %s\n", reportPath)

		// Also print summary to console
		fmt.Printf("\n%d/%d passed (%d%%)\n", summary.Passed, summary.Total,
			calculatePassRate(summary.Passed, summary.Total))
	} else {
		// Console or JSON - write to stdout
		if err := rep.Report(os.Stdout, results, summary, svc); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}

	// Return error if any tests failed
	if summary.Failed > 0 {
		return fmt.Errorf("%d test(s) failed", summary.Failed)
	}

	return nil
}

func calculatePassRate(passed, total int) int {
	if total == 0 {
		return 0
	}
	return (passed * 100) / total
}

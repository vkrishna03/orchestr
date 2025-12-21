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
	"github.com/vkrishna03/orchestr/internal/parser"
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

	// Display test cases
	fmt.Printf("\nRunning %d test(s):\n\n", len(specs))
	for i, spec := range specs {
		entity := spec.Entity
		if entity == "" {
			entity = "Test"
		}
		fmt.Printf("  %d. [%s] %s %s → %d\n",
			i+1,
			entity,
			spec.Method,
			spec.Endpoint,
			spec.Expect.Status,
		)
	}

	// TODO: Implement test execution
	// 1. Build HTTP requests from specs
	// 2. Execute requests against svc.Port
	// 3. Validate responses against expectations
	// 4. Handle SAVE directives
	// 5. Generate report

	fmt.Println()
	slog.Warn("test execution not yet implemented")

	return nil
}

package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vkrishna03/orchestr/internal/parser"
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

	_ = viper.BindPFlag("format", testCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("spec", testCmd.Flags().Lookup("spec"))
}

func runTest(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Get config from Viper (flags, env, or config file)
	format := viper.GetString("format")
	specPath := viper.GetString("spec")
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

	for i, spec := range specs {
		slog.Info("test case",
			"num", i+1,
			"method", spec.Method,
			"endpoint", spec.Endpoint,
			"entity", spec.Entity,
			"expect", spec.Expect.Status,
		)
	}

	// TODO: Implement test runner
	// 2. Discover services
	// 3. Start services
	// 4. Execute tests
	// 5. Generate report

	slog.Warn("execution not yet implemented")
	return nil
}

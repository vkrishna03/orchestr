package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	formatFlag string
	specFlag   string
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

	testCmd.Flags().StringVarP(&formatFlag, "format", "f", "console", "output format (console, json, markdown)")
	testCmd.Flags().StringVarP(&specFlag, "spec", "s", "", "path to Orchestrfile (default: ./Orchestrfile)")
}

func runTest(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	fmt.Printf("Testing project at: %s\n", path)
	fmt.Printf("Format: %s\n", formatFlag)

	// TODO: Implement test runner
	// 1. Parse Orchestrfile
	// 2. Discover services
	// 3. Start services
	// 4. Execute tests
	// 5. Generate report

	fmt.Println("\n[Not yet implemented]")
	return nil
}

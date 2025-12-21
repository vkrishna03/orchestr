package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover [path]",
	Short: "Show detected services",
	Long: `Scan a directory and show all detected services with their dependencies.

Examples:
  orchestr discover .
  orchestr discover ./microservices`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDiscover,
}

func init() {
	rootCmd.AddCommand(discoverCmd)
}

func runDiscover(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	fmt.Printf("Discovering services in: %s\n", path)

	// TODO: Implement service discovery
	// 1. Walk directory tree
	// 2. Detect service types (Spring Boot, Vite, etc.)
	// 3. Build dependency graph
	// 4. Display results

	fmt.Println("\n[Not yet implemented]")
	return nil
}

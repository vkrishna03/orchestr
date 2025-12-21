package cli

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vkrishna03/orchestr/internal/discovery"
)

var discoverCmd = &cobra.Command{
	Use:   "discover [path]",
	Short: "Show detected services",
	Long: `Scan a directory and show all detected services with their configuration.

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

	slog.Debug("discovering services", "path", path)

	discoverer := discovery.NewDiscoverer()
	services, err := discoverer.Discover(path)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(services) == 0 {
		fmt.Println("No services found.")
		fmt.Println("\nSupported service types:")
		for _, t := range discoverer.SupportedTypes() {
			fmt.Printf("  - %s\n", t)
		}
		return nil
	}

	// Print header
	fmt.Printf("\nDiscovered %d service(s):\n\n", len(services))

	// Print table header
	fmt.Printf("%-20s %-15s %-8s %-10s %s\n",
		"NAME", "TYPE", "PORT", "BUILD", "START COMMAND")
	fmt.Println(strings.Repeat("-", 80))

	// Print services
	for _, svc := range services {
		startCmd := strings.Join(svc.StartCmd, " ")
		if len(startCmd) > 25 {
			startCmd = startCmd[:22] + "..."
		}

		fmt.Printf("%-20s %-15s %-8d %-10s %s\n",
			truncate(svc.Name, 20),
			svc.Type,
			svc.Port,
			svc.BuildTool,
			startCmd,
		)
	}

	// Print details if verbose
	if verbose {
		fmt.Println("\nDetails:")
		for _, svc := range services {
			fmt.Printf("\n%s:\n", svc.Name)
			fmt.Printf("  Path:      %s\n", svc.Path)
			fmt.Printf("  Type:      %s\n", svc.Type)
			fmt.Printf("  Port:      %d\n", svc.Port)
			fmt.Printf("  Build:     %s\n", svc.BuildTool)
			fmt.Printf("  Start:     %s\n", strings.Join(svc.StartCmd, " "))
			fmt.Printf("  Health:    %s\n", svc.HealthURL)
			if len(svc.BuildCmd) > 0 {
				fmt.Printf("  BuildCmd:  %s\n", strings.Join(svc.BuildCmd, " "))
			}
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate Orchestrfile template",
	Long: `Generate a starter Orchestrfile template in the current directory.

Example:
  orchestr init`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	orchestrfile := filepath.Join(".", "Orchestrfile")

	if _, err := os.Stat(orchestrfile); err == nil {
		return fmt.Errorf("Orchestrfile already exists")
	}

	template := `# Orchestrfile - Test specifications
# See: https://github.com/vkrishna03/orchestr

ENTITY: User
ENDPOINT: /api/health
METHOD: GET
EXPECT: 200 with {status: "ok"}

---

# Example: User registration
# ENTITY: User
# ENDPOINT: /api/register
# METHOD: POST
# DATA: {email: "test@example.com", password: "test123"}
# EXPECT: 201 with {id: <uuid>, email: "test@example.com"}

---

# Example: User login with variable save
# ENTITY: User
# ENDPOINT: /api/login
# METHOD: POST
# DATA: {email: "test@example.com", password: "test123"}
# EXPECT: 200 with {token: <jwt>}
# SAVE: token from body.token

---

# Example: Authenticated request
# ENTITY: User
# ENDPOINT: /api/me
# METHOD: GET
# AUTH: Bearer {{token}}
# EXPECT: 200 with {email: "test@example.com"}
`

	if err := os.WriteFile(orchestrfile, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write Orchestrfile: %w", err)
	}

	fmt.Println("Created Orchestrfile")
	return nil
}

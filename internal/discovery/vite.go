package discovery

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/vkrishna03/orchestr/internal/models"
)

const defaultVitePort = 5173

// ViteDetector detects Vite-based frontend projects.
type ViteDetector struct{}

// NewViteDetector creates a new Vite detector.
func NewViteDetector() *ViteDetector {
	return &ViteDetector{}
}

// Type returns the service type.
func (d *ViteDetector) Type() models.ServiceType {
	return models.ServiceVite
}

// Detect checks if the directory contains a Vite project.
func (d *ViteDetector) Detect(dir string) bool {
	// Check for vite config files
	viteConfigs := []string{
		"vite.config.ts",
		"vite.config.js",
		"vite.config.mts",
		"vite.config.mjs",
	}

	for _, config := range viteConfigs {
		if _, err := os.Stat(filepath.Join(dir, config)); err == nil {
			// Verify package.json exists
			if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
				return true
			}
		}
	}

	return false
}

// Analyze extracts Vite service configuration.
func (d *ViteDetector) Analyze(dir string) (*models.Service, error) {
	svc := &models.Service{
		Name:   filepath.Base(dir),
		Type:   models.ServiceVite,
		Path:   dir,
		Port:   defaultVitePort,
		Config: make(map[string]string),
	}

	// Determine package manager
	svc.BuildTool = d.detectPackageManager(dir)

	// Set commands based on package manager
	switch svc.BuildTool {
	case models.BuildPNPM:
		svc.StartCmd = []string{"pnpm", "dev"}
		svc.BuildCmd = []string{"pnpm", "build"}
	case models.BuildYarn:
		svc.StartCmd = []string{"yarn", "dev"}
		svc.BuildCmd = []string{"yarn", "build"}
	default:
		svc.BuildTool = models.BuildNPM
		svc.StartCmd = []string{"npm", "run", "dev"}
		svc.BuildCmd = []string{"npm", "run", "build"}
	}

	// Extract port from vite config
	port := d.extractPort(dir)
	if port > 0 {
		svc.Port = port
	}

	// Set health check URL
	svc.HealthURL = fmt.Sprintf("http://localhost:%d/", svc.Port)

	return svc, nil
}

// detectPackageManager determines which package manager is used.
func (d *ViteDetector) detectPackageManager(dir string) models.BuildTool {
	// Check for lockfiles in order of preference
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		return models.BuildPNPM
	}
	if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		return models.BuildYarn
	}
	if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); err == nil {
		return models.BuildNPM
	}

	// Default to npm
	return models.BuildNPM
}

// extractPort reads server.port from vite config.
func (d *ViteDetector) extractPort(dir string) int {
	configs := []string{
		"vite.config.ts",
		"vite.config.js",
		"vite.config.mts",
		"vite.config.mjs",
	}

	for _, config := range configs {
		if port := d.extractPortFromConfig(filepath.Join(dir, config)); port > 0 {
			return port
		}
	}

	return 0
}

// extractPortFromConfig parses server.port from vite config file.
// Uses simple regex matching to avoid adding a JS parser dependency.
func (d *ViteDetector) extractPortFromConfig(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Patterns to match port configuration in vite config
	// Matches: port: 3000, port: 3000, "port": 3000
	portRegex := regexp.MustCompile(`['"]?port['"]?\s*:\s*(\d+)`)

	scanner := bufio.NewScanner(file)
	inServer := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're entering server config block
		if strings.Contains(line, "server:") || strings.Contains(line, "server :") {
			inServer = true
		}

		// Look for port in server block
		if inServer {
			if matches := portRegex.FindStringSubmatch(line); len(matches) > 1 {
				port, _ := strconv.Atoi(matches[1])
				return port
			}
		}

		// Also check for inline server config: server: { port: 3000 }
		if strings.Contains(line, "server") && strings.Contains(line, "port") {
			if matches := portRegex.FindStringSubmatch(line); len(matches) > 1 {
				port, _ := strconv.Atoi(matches[1])
				return port
			}
		}
	}

	return 0
}

// Ensure ViteDetector implements Detector.
var _ Detector = (*ViteDetector)(nil)

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

const defaultSpringBootPort = 8080

// SpringBootDetector detects Spring Boot services.
type SpringBootDetector struct{}

// NewSpringBootDetector creates a new Spring Boot detector.
func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{}
}

// Type returns the service type.
func (d *SpringBootDetector) Type() models.ServiceType {
	return models.ServiceSpringBoot
}

// Detect checks if the directory contains a Spring Boot project.
func (d *SpringBootDetector) Detect(dir string) bool {
	// Check for Maven pom.xml
	if d.isPomSpringBoot(filepath.Join(dir, "pom.xml")) {
		return true
	}

	// Check for Gradle build files
	if d.isGradleSpringBoot(dir) {
		return true
	}

	return false
}

// Analyze extracts Spring Boot service configuration.
func (d *SpringBootDetector) Analyze(dir string) (*models.Service, error) {
	svc := &models.Service{
		Name:   filepath.Base(dir),
		Type:   models.ServiceSpringBoot,
		Path:   dir,
		Port:   defaultSpringBootPort,
		Config: make(map[string]string),
	}

	// Determine build tool and commands
	if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
		svc.BuildTool = models.BuildMaven
		svc.StartCmd = []string{"mvn", "spring-boot:run"}
		svc.BuildCmd = []string{"mvn", "package", "-DskipTests"}
	} else if d.hasGradleWrapper(dir) {
		svc.BuildTool = models.BuildGradle
		svc.StartCmd = []string{"./gradlew", "bootRun"}
		svc.BuildCmd = []string{"./gradlew", "build", "-x", "test"}
	} else if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
		svc.BuildTool = models.BuildGradle
		svc.StartCmd = []string{"gradle", "bootRun"}
		svc.BuildCmd = []string{"gradle", "build", "-x", "test"}
	} else if _, err := os.Stat(filepath.Join(dir, "build.gradle.kts")); err == nil {
		svc.BuildTool = models.BuildGradle
		svc.StartCmd = []string{"gradle", "bootRun"}
		svc.BuildCmd = []string{"gradle", "build", "-x", "test"}
	}

	// Extract port from application config
	port := d.extractPort(dir)
	if port > 0 {
		svc.Port = port
	}

	// Set health check URL
	svc.HealthURL = fmt.Sprintf("http://localhost:%d/actuator/health", svc.Port)

	return svc, nil
}

// isPomSpringBoot checks if pom.xml contains Spring Boot dependencies.
func (d *SpringBootDetector) isPomSpringBoot(pomPath string) bool {
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return false
	}

	s := string(content)
	return strings.Contains(s, "spring-boot-starter") ||
		strings.Contains(s, "spring-boot-maven-plugin") ||
		strings.Contains(s, "org.springframework.boot")
}

// isGradleSpringBoot checks if Gradle files contain Spring Boot.
func (d *SpringBootDetector) isGradleSpringBoot(dir string) bool {
	// Check build.gradle
	if content, err := os.ReadFile(filepath.Join(dir, "build.gradle")); err == nil {
		s := string(content)
		if strings.Contains(s, "org.springframework.boot") ||
			strings.Contains(s, "spring-boot-starter") {
			return true
		}
	}

	// Check build.gradle.kts
	if content, err := os.ReadFile(filepath.Join(dir, "build.gradle.kts")); err == nil {
		s := string(content)
		if strings.Contains(s, "org.springframework.boot") ||
			strings.Contains(s, "spring-boot-starter") {
			return true
		}
	}

	return false
}

// hasGradleWrapper checks if gradlew exists.
func (d *SpringBootDetector) hasGradleWrapper(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "gradlew"))
	return err == nil
}

// extractPort reads server.port from application.yml or application.properties.
// Checks main config files first, then profile-specific files.
func (d *SpringBootDetector) extractPort(dir string) int {
	resourcesDir := filepath.Join(dir, "src", "main", "resources")

	// Try main application.yml/yaml
	for _, name := range []string{"application.yml", "application.yaml"} {
		if port := d.extractPortFromYAML(filepath.Join(resourcesDir, name)); port > 0 {
			return port
		}
	}

	// Try main application.properties
	if port := d.extractPortFromProperties(filepath.Join(resourcesDir, "application.properties")); port > 0 {
		return port
	}

	// Try common profile files
	profiles := []string{"dev", "local", "default"}
	for _, profile := range profiles {
		for _, ext := range []string{"yml", "yaml"} {
			path := filepath.Join(resourcesDir, fmt.Sprintf("application-%s.%s", profile, ext))
			if port := d.extractPortFromYAML(path); port > 0 {
				return port
			}
		}
		path := filepath.Join(resourcesDir, fmt.Sprintf("application-%s.properties", profile))
		if port := d.extractPortFromProperties(path); port > 0 {
			return port
		}
	}

	return 0
}

// extractPortFromYAML parses server.port from YAML file.
// Uses simple line parsing to avoid adding a YAML dependency.
func (d *SpringBootDetector) extractPortFromYAML(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Simple YAML parsing for server.port
	// Handles:
	//   server:
	//     port: 8081
	//   server.port: 8081
	//   port: ${PORT:8081}  (extracts default value)
	scanner := bufio.NewScanner(file)
	inServer := false
	portRegex := regexp.MustCompile(`^\s*port:\s*(.+)`)
	serverPortRegex := regexp.MustCompile(`^\s*server\.port:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for server.port: directly
		if matches := serverPortRegex.FindStringSubmatch(line); len(matches) > 1 {
			if port := d.parsePortValue(matches[1]); port > 0 {
				return port
			}
		}

		// Check for server: block
		if strings.HasPrefix(strings.TrimSpace(line), "server:") {
			inServer = true
			continue
		}

		// Check for port: under server:
		if inServer {
			if matches := portRegex.FindStringSubmatch(line); len(matches) > 1 {
				if port := d.parsePortValue(matches[1]); port > 0 {
					return port
				}
			}
			// Exit server block if we hit another top-level key
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && !strings.HasPrefix(line, "#") {
				inServer = false
			}
		}
	}

	return 0
}

// extractPortFromProperties parses server.port from properties file.
func (d *SpringBootDetector) extractPortFromProperties(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "server.port=") {
			value := strings.TrimPrefix(line, "server.port=")
			return d.parsePortValue(value)
		}
	}

	return 0
}

// parsePortValue extracts a port number from a value string.
// Handles plain numbers (8080) and Spring placeholders (${PORT:8080}).
var envPlaceholderRegex = regexp.MustCompile(`\$\{[^:}]+:(\d+)\}`)

func (d *SpringBootDetector) parsePortValue(value string) int {
	value = strings.TrimSpace(value)

	// Try plain number first
	if port, err := strconv.Atoi(value); err == nil {
		return port
	}

	// Try ${VAR:default} pattern
	if matches := envPlaceholderRegex.FindStringSubmatch(value); len(matches) > 1 {
		port, _ := strconv.Atoi(matches[1])
		return port
	}

	return 0
}

// Ensure SpringBootDetector implements Detector.
var _ Detector = (*SpringBootDetector)(nil)

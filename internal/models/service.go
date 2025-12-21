package models

// ServiceType identifies the type of service (e.g., Spring Boot, Vite).
type ServiceType string

// Supported service types.
const (
	ServiceSpringBoot ServiceType = "spring-boot"
	ServiceVite       ServiceType = "vite"
)

// BuildTool identifies the build tool used by a service.
type BuildTool string

// Supported build tools.
const (
	BuildMaven  BuildTool = "maven"
	BuildGradle BuildTool = "gradle"
	BuildNPM    BuildTool = "npm"
	BuildPNPM   BuildTool = "pnpm"
	BuildYarn   BuildTool = "yarn"
)

// Service represents a discovered service in the project.
type Service struct {
	// Name is the service name, derived from directory name.
	Name string

	// Type is the detected service type.
	Type ServiceType

	// Path is the absolute path to the service root directory.
	Path string

	// Port is the detected or default port number.
	Port int

	// BuildTool is the detected build tool (maven, gradle, npm, etc.).
	BuildTool BuildTool

	// StartCmd is the command to start the service.
	StartCmd []string

	// BuildCmd is the optional command to build the service.
	BuildCmd []string

	// HealthURL is the health check endpoint.
	HealthURL string

	// Config holds additional configuration extracted from the service.
	Config map[string]string
}

// String returns a human-readable representation of the service type.
func (t ServiceType) String() string {
	return string(t)
}

// String returns a human-readable representation of the build tool.
func (t BuildTool) String() string {
	return string(t)
}

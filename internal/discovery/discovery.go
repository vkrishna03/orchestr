package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vkrishna03/orchestr/internal/models"
)

// Discoverer finds services in a directory tree.
type Discoverer struct {
	detectors []Detector
}

// NewDiscoverer creates a discoverer with default detectors.
func NewDiscoverer() *Discoverer {
	return &Discoverer{
		detectors: []Detector{
			NewSpringBootDetector(),
			NewViteDetector(),
		},
	}
}

// Option configures the discoverer.
type Option func(*Discoverer)

// WithDetectors sets custom detectors.
func WithDetectors(detectors ...Detector) Option {
	return func(d *Discoverer) {
		d.detectors = detectors
	}
}

// NewDiscovererWithOptions creates a discoverer with options.
func NewDiscovererWithOptions(opts ...Option) *Discoverer {
	d := NewDiscoverer()
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// DiscoverOne checks a single directory for a service.
// Returns nil if no service is detected.
func (d *Discoverer) DiscoverOne(dir string) (*models.Service, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absDir)
	}

	for _, detector := range d.detectors {
		if detector.Detect(absDir) {
			svc, err := detector.Analyze(absDir)
			if err != nil {
				return nil, fmt.Errorf("analyze %s: %w", detector.Type(), err)
			}
			// Ensure path is absolute
			svc.Path = absDir
			return svc, nil
		}
	}

	return nil, nil
}

// Discover recursively scans a directory for services.
// Returns all discovered services.
func (d *Discoverer) Discover(root string) ([]*models.Service, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absRoot)
	}

	var services []*models.Service

	err = filepath.WalkDir(absRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories
		if !entry.IsDir() {
			return nil
		}

		// Skip common non-service directories
		name := entry.Name()
		if shouldSkipDir(name) {
			return filepath.SkipDir
		}

		// Try to detect a service in this directory
		svc, err := d.DiscoverOne(path)
		if err != nil {
			// Log but don't fail on detection errors
			return nil
		}

		if svc != nil {
			services = append(services, svc)
			// Don't recurse into detected service directories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return services, nil
}

// shouldSkipDir returns true for directories that should not be scanned.
func shouldSkipDir(name string) bool {
	skipDirs := map[string]bool{
		".git":         true,
		".svn":         true,
		".hg":          true,
		"node_modules": true,
		"target":       true,
		"build":        true,
		"dist":         true,
		".gradle":      true,
		".mvn":         true,
		"vendor":       true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
	}
	return skipDirs[name]
}

// SupportedTypes returns all service types that can be detected.
func (d *Discoverer) SupportedTypes() []models.ServiceType {
	types := make([]models.ServiceType, len(d.detectors))
	for i, detector := range d.detectors {
		types[i] = detector.Type()
	}
	return types
}

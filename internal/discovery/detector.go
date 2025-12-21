// Package discovery provides service detection for various frameworks.
package discovery

import "github.com/vkrishna03/orchestr/internal/models"

// Detector defines the interface for service type detection.
type Detector interface {
	// Detect checks if the given directory contains this service type.
	// Returns true if the service type is detected.
	Detect(dir string) bool

	// Analyze extracts service configuration from the directory.
	// Should only be called after Detect returns true.
	Analyze(dir string) (*models.Service, error)

	// Type returns the service type this detector handles.
	Type() models.ServiceType
}

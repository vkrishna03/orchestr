// Package runner provides service lifecycle management.
package runner

import (
	"context"
	"time"

	"github.com/vkrishna03/orchestr/internal/models"
)

// Runner manages the lifecycle of a service.
type Runner interface {
	// Start launches the service and waits for it to become healthy.
	Start(ctx context.Context, svc *models.Service) error

	// Stop gracefully shuts down the service.
	Stop(ctx context.Context) error

	// Logs returns captured stdout/stderr from the service.
	Logs() string

	// Status returns the current service state.
	Status() ServiceStatus
}

// ServiceStatus represents the current state of a service.
type ServiceStatus string

const (
	StatusStopped  ServiceStatus = "stopped"
	StatusStarting ServiceStatus = "starting"
	StatusHealthy  ServiceStatus = "healthy"
	StatusFailed   ServiceStatus = "failed"
)

// Options configures runner behavior.
type Options struct {
	// StartTimeout is the maximum time to wait for the service to start.
	StartTimeout time.Duration

	// HealthTimeout is the maximum time to wait for health checks.
	HealthTimeout time.Duration

	// HealthRetries is the number of health check attempts.
	HealthRetries int

	// HealthInterval is the initial delay between health checks.
	HealthInterval time.Duration

	// StopTimeout is the maximum time to wait for graceful shutdown.
	StopTimeout time.Duration

	// Verbose enables printing service output to stdout/stderr.
	Verbose bool
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		StartTimeout:   60 * time.Second,
		HealthTimeout:  30 * time.Second,
		HealthRetries:  10,
		HealthInterval: 500 * time.Millisecond,
		StopTimeout:    10 * time.Second,
		Verbose:        false,
	}
}

// String returns the status as a string.
func (s ServiceStatus) String() string {
	return string(s)
}

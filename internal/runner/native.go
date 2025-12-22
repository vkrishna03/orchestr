package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/vkrishna03/orchestr/internal/models"
)

// NativeRunner runs services using os/exec.
type NativeRunner struct {
	opts   Options
	cmd    *exec.Cmd
	status ServiceStatus
	logs   *bytes.Buffer
	mu     sync.RWMutex
	done   chan struct{}
}

// NewNativeRunner creates a new native runner with the given options.
func NewNativeRunner(opts Options) *NativeRunner {
	return &NativeRunner{
		opts:   opts,
		status: StatusStopped,
		logs:   &bytes.Buffer{},
		done:   make(chan struct{}),
	}
}

// Start launches the service and waits for it to become healthy.
func (r *NativeRunner) Start(ctx context.Context, svc *models.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusStopped {
		return fmt.Errorf("service already running (status: %s)", r.status)
	}

	if len(svc.StartCmd) == 0 {
		return fmt.Errorf("no start command for service %s", svc.Name)
	}

	r.status = StatusStarting
	r.logs.Reset()
	r.done = make(chan struct{})

	// Build command
	r.cmd = exec.CommandContext(ctx, svc.StartCmd[0], svc.StartCmd[1:]...)
	r.cmd.Dir = svc.Path

	// Set up log capture
	if r.opts.Verbose {
		r.cmd.Stdout = io.MultiWriter(r.logs, os.Stdout)
		r.cmd.Stderr = io.MultiWriter(r.logs, os.Stderr)
	} else {
		r.cmd.Stdout = r.logs
		r.cmd.Stderr = r.logs
	}

	// Set process group for clean shutdown (platform-specific)
	setProcGroup(r.cmd)

	// Start process
	if err := r.cmd.Start(); err != nil {
		r.status = StatusFailed
		return fmt.Errorf("start process: %w", err)
	}

	// Monitor process in background
	go func() {
		_ = r.cmd.Wait()
		close(r.done)
	}()

	// Wait for health check
	healthCtx, cancel := context.WithTimeout(ctx, r.opts.HealthTimeout)
	defer cancel()

	checker := NewHealthChecker(r.opts.HealthTimeout, r.opts.HealthRetries, r.opts.HealthInterval)
	tcpAddr := fmt.Sprintf("localhost:%d", svc.Port)

	err := checker.Check(healthCtx, svc.HealthURL, tcpAddr)
	if err != nil {
		// Check if process died
		select {
		case <-r.done:
			r.status = StatusFailed
			return fmt.Errorf("service exited during startup: %w", err)
		default:
		}

		r.status = StatusFailed
		// Try to stop the failed process
		_ = r.stopProcess(context.Background())
		return fmt.Errorf("health check failed: %w", err)
	}

	r.status = StatusHealthy
	return nil
}

// Stop gracefully shuts down the service.
func (r *NativeRunner) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.stopProcess(ctx)
}

// stopProcess stops the running process (must be called with lock held).
func (r *NativeRunner) stopProcess(ctx context.Context) error {
	if r.cmd == nil || r.cmd.Process == nil {
		r.status = StatusStopped
		return nil
	}

	// Check if already exited
	select {
	case <-r.done:
		r.status = StatusStopped
		return nil
	default:
	}

	// Send interrupt signal (platform-specific)
	_ = interruptProcess(r.cmd)

	// Wait for graceful shutdown
	select {
	case <-r.done:
		r.status = StatusStopped
		return nil
	case <-time.After(r.opts.StopTimeout):
		// Force kill (platform-specific)
		_ = forceKillProcess(r.cmd)
		<-r.done
		r.status = StatusStopped
		return nil
	case <-ctx.Done():
		// Force kill on context cancel
		_ = r.cmd.Process.Kill()
		r.status = StatusStopped
		return ctx.Err()
	}
}

// Logs returns captured stdout/stderr from the service.
func (r *NativeRunner) Logs() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.logs.String()
}

// Status returns the current service state.
func (r *NativeRunner) Status() ServiceStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

// Ensure NativeRunner implements Runner.
var _ Runner = (*NativeRunner)(nil)

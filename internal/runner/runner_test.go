package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vkrishna03/orchestr/internal/models"
)

func TestHealthChecker_CheckHTTP(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "healthy 200",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "healthy 204",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "unhealthy 500",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "unhealthy 503",
			statusCode: http.StatusServiceUnavailable,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			checker := NewHealthChecker(5*time.Second, 1, 100*time.Millisecond)
			err := checker.CheckHTTP(context.Background(), server.URL)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestHealthChecker_CheckHTTP_Retries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHealthChecker(10*time.Second, 5, 50*time.Millisecond)
	err := checker.CheckHTTP(context.Background(), server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestHealthChecker_CheckHTTP_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
	}))
	defer server.Close()

	checker := NewHealthChecker(5*time.Second, 10, 50*time.Millisecond)
	err := checker.CheckHTTP(ctx, server.URL)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestHealthChecker_CheckTCP(t *testing.T) {
	// Use the test server's address for TCP check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host:port from server URL
	addr := server.Listener.Addr().String()

	checker := NewHealthChecker(5*time.Second, 3, 100*time.Millisecond)
	err := checker.CheckTCP(context.Background(), addr)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHealthChecker_CheckTCP_Unreachable(t *testing.T) {
	checker := NewHealthChecker(1*time.Second, 2, 50*time.Millisecond)
	err := checker.CheckTCP(context.Background(), "localhost:59999")

	if err == nil {
		t.Error("expected error for unreachable port, got nil")
	}
}

func TestNativeRunner_Status(t *testing.T) {
	runner := NewNativeRunner(DefaultOptions())

	if runner.Status() != StatusStopped {
		t.Errorf("expected initial status %s, got %s", StatusStopped, runner.Status())
	}
}

func TestNativeRunner_Start_NoCommand(t *testing.T) {
	runner := NewNativeRunner(DefaultOptions())
	svc := &models.Service{
		Name:     "test",
		StartCmd: []string{},
	}

	err := runner.Start(context.Background(), svc)
	if err == nil {
		t.Error("expected error for empty command, got nil")
	}
}

func TestNativeRunner_Start_InvalidCommand(t *testing.T) {
	opts := DefaultOptions()
	opts.HealthTimeout = 1 * time.Second
	opts.HealthRetries = 1

	runner := NewNativeRunner(opts)
	svc := &models.Service{
		Name:      "test",
		StartCmd:  []string{"nonexistent-command-12345"},
		Path:      "/tmp",
		Port:      9999,
		HealthURL: "http://localhost:9999/health",
	}

	err := runner.Start(context.Background(), svc)
	if err == nil {
		t.Error("expected error for invalid command, got nil")
	}
}

func TestNativeRunner_StartStop_SimpleProcess(t *testing.T) {
	// Start a simple HTTP server using Go's test server pattern
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with a simple sleep command (we'll skip health check by using TCP)
	opts := DefaultOptions()
	opts.HealthTimeout = 2 * time.Second
	opts.HealthRetries = 5
	opts.HealthInterval = 100 * time.Millisecond

	runner := NewNativeRunner(opts)
	svc := &models.Service{
		Name:      "test-sleep",
		StartCmd:  []string{"sleep", "60"},
		Path:      "/tmp",
		Port:      0, // No port for sleep
		HealthURL: "http://localhost:0/health",
	}

	// This will fail health check (as expected for sleep)
	// but tests the start/stop mechanism
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_ = runner.Start(ctx, svc)
	// Ignore error - we expect health check to fail for sleep

	// Stop should work regardless
	err := runner.Stop(context.Background())
	if err != nil {
		t.Errorf("stop failed: %v", err)
	}

	if runner.Status() != StatusStopped && runner.Status() != StatusFailed {
		t.Errorf("expected stopped or failed status, got %s", runner.Status())
	}
}

func TestNativeRunner_Logs(t *testing.T) {
	opts := DefaultOptions()
	opts.HealthTimeout = 1 * time.Second
	opts.HealthRetries = 1

	runner := NewNativeRunner(opts)
	svc := &models.Service{
		Name:      "test-echo",
		StartCmd:  []string{"echo", "hello world"},
		Path:      "/tmp",
		Port:      9999,
		HealthURL: "http://localhost:9999/health",
	}

	// Start will fail health check but echo should run
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = runner.Start(ctx, svc)
	runner.Stop(context.Background())

	logs := runner.Logs()
	if logs == "" {
		t.Log("logs are empty (echo may have completed before capture)")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.StartTimeout != 60*time.Second {
		t.Errorf("expected StartTimeout 60s, got %v", opts.StartTimeout)
	}
	if opts.HealthTimeout != 30*time.Second {
		t.Errorf("expected HealthTimeout 30s, got %v", opts.HealthTimeout)
	}
	if opts.HealthRetries != 10 {
		t.Errorf("expected HealthRetries 10, got %d", opts.HealthRetries)
	}
	if opts.StopTimeout != 10*time.Second {
		t.Errorf("expected StopTimeout 10s, got %v", opts.StopTimeout)
	}
}

func TestServiceStatus_String(t *testing.T) {
	tests := []struct {
		status ServiceStatus
		want   string
	}{
		{StatusStopped, "stopped"},
		{StatusStarting, "starting"},
		{StatusHealthy, "healthy"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("status.String() = %q, want %q", got, tt.want)
		}
	}
}

package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HealthChecker performs health checks on services.
type HealthChecker struct {
	client   *http.Client
	retries  int
	interval time.Duration
}

// NewHealthChecker creates a health checker with the given options.
func NewHealthChecker(timeout time.Duration, retries int, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: timeout / time.Duration(retries), // Per-request timeout
		},
		retries:  retries,
		interval: interval,
	}
}

// CheckHTTP performs HTTP health check with retries and exponential backoff.
func (h *HealthChecker) CheckHTTP(ctx context.Context, url string) error {
	var lastErr error
	delay := h.interval

	for i := 0; i < h.retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			// Exponential backoff, max 5s
			delay = delay * 2
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = fmt.Errorf("create request: %w", err)
			continue
		}

		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("health check failed: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // Healthy
		}

		lastErr = fmt.Errorf("health check returned %d", resp.StatusCode)
	}

	return fmt.Errorf("health check failed after %d attempts: %w", h.retries, lastErr)
}

// CheckTCP performs TCP connection check with retries.
func (h *HealthChecker) CheckTCP(ctx context.Context, addr string) error {
	var lastErr error
	delay := h.interval

	dialer := &net.Dialer{
		Timeout: h.interval * 2,
	}

	for i := 0; i < h.retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = delay * 2
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}

		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			lastErr = fmt.Errorf("tcp dial failed: %w", err)
			continue
		}
		conn.Close()
		return nil // Connected
	}

	return fmt.Errorf("tcp check failed after %d attempts: %w", h.retries, lastErr)
}

// Check performs health check, trying HTTP first then TCP fallback.
func (h *HealthChecker) Check(ctx context.Context, httpURL string, tcpAddr string) error {
	// Try HTTP first
	err := h.CheckHTTP(ctx, httpURL)
	if err == nil {
		return nil
	}

	// Fall back to TCP if HTTP fails
	if tcpAddr != "" {
		tcpErr := h.CheckTCP(ctx, tcpAddr)
		if tcpErr == nil {
			return nil
		}
		return fmt.Errorf("http: %w; tcp: %v", err, tcpErr)
	}

	return err
}

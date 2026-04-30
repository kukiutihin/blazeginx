package httpserver

import (
	"blazeginx/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ServerWithLimiter(cfg config.RateLimit, t *testing.T) *httptest.Server {
	buckets := NewTokenBucket(cfg)
	return getServer(
		[]func(http.Handler) http.Handler{
			buckets.BuildRateLimiter(),
		},
		[]handler{
			{
				"/",
				getHandler(t, "Hi"),
			},
		},
	)
}

func TestLimitExceeded(t *testing.T) {
	t.Parallel()
	cfg := config.RateLimit{
		Enabled:           true,
		MaxTokens:         2,
		DefaultExpiration: time.Hour * 1,
		RefillRate:        time.Hour * 1,
	}
	server := ServerWithLimiter(cfg, t)

	if status := getRequestCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := getRequestCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := getRequestCode(server.URL, t); status != http.StatusTooManyRequests {
		t.Errorf("Expected status: %d, but got: %d", http.StatusTooManyRequests, status)
	}
}

func TestRefillTokens(t *testing.T) {
	t.Parallel()
	cfg := config.RateLimit{
		Enabled:           true,
		MaxTokens:         2,
		DefaultExpiration: time.Hour * 1,
		RefillRate:        time.Second * 1,
	}
	server := ServerWithLimiter(cfg, t)

	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)

	time.Sleep(time.Second * 1)

	if status := getRequestCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}
}

func TestCrossRefill(t *testing.T) {
	t.Parallel()
	cfg := config.RateLimit{
		Enabled:           true,
		MaxTokens:         2,
		DefaultExpiration: time.Hour * 1,
		RefillRate:        time.Second * 1,
	}
	server := ServerWithLimiter(cfg, t)

	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)

	time.Sleep(time.Second * 1)

	getRequestCode(server.URL, t)
	if status := getRequestCode(server.URL, t); status != http.StatusTooManyRequests {
		t.Errorf("Expected status: %d, but got: %d", http.StatusTooManyRequests, status)
	}
}

func TestMaxTokens(t *testing.T) {
	t.Parallel()
	cfg := config.RateLimit{
		Enabled:           true,
		MaxTokens:         3,
		DefaultExpiration: time.Hour * 1,
		RefillRate:        time.Second * 1,
	}
	server := ServerWithLimiter(cfg, t)

	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)

	time.Sleep(time.Second * 5)

	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)
	getRequestCode(server.URL, t)
	if status := getRequestCode(server.URL, t); status != http.StatusTooManyRequests {
		t.Errorf("Expected status: %d, but got: %d", http.StatusTooManyRequests, status)
	}
}

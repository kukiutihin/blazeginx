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
	return getServer(t, buckets.BuildRateLimiter())
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

	if status := getRequest(server, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := getRequest(server, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := getRequest(server, t); status != http.StatusTooManyRequests {
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

	getRequest(server, t)
	getRequest(server, t)

	time.Sleep(time.Second * 1)

	if status := getRequest(server, t); status != http.StatusOK {
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

	getRequest(server, t)
	getRequest(server, t)

	time.Sleep(time.Second * 1)

	getRequest(server, t)
	if status := getRequest(server, t); status != http.StatusTooManyRequests {
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

	getRequest(server, t)
	getRequest(server, t)
	getRequest(server, t)

	time.Sleep(time.Second * 5)

	getRequest(server, t)
	getRequest(server, t)
	getRequest(server, t)
	getRequest(server, t)
	if status := getRequest(server, t); status != http.StatusTooManyRequests {
		t.Errorf("Expected status: %d, but got: %d", http.StatusTooManyRequests, status)
	}
}

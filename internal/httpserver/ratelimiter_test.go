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
	return NewServer(
		[]func(http.Handler) http.Handler{
			buckets.BuildRateLimiter(),
		},
		[]handler{
			{
				"/",
				NewHandler(t, "Hi"),
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

	if status := DoGetCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := DoGetCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	if status := DoGetCode(server.URL, t); status != http.StatusTooManyRequests {
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

	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)

	time.Sleep(time.Second * 1)

	if status := DoGetCode(server.URL, t); status != http.StatusOK {
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

	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)

	time.Sleep(time.Second * 1)

	DoGetCode(server.URL, t)
	if status := DoGetCode(server.URL, t); status != http.StatusTooManyRequests {
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

	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)

	time.Sleep(time.Second * 5)

	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)
	DoGetCode(server.URL, t)
	if status := DoGetCode(server.URL, t); status != http.StatusTooManyRequests {
		t.Errorf("Expected status: %d, but got: %d", http.StatusTooManyRequests, status)
	}
}

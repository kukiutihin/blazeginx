package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// sleepyMiddleware must respect the contract and check context.
func sleepyMiddleware(sleep time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(sleep)
			if IsDone(r.Context()) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func serverWithTimeout(
	t *testing.T,
	timeout time.Duration,
	sleep time.Duration,
) *httptest.Server {
	return getServer(
		[]func(http.Handler) http.Handler{
			emptyRequestLogger(logger.New(config.EnvLocal)),
			Timeout(timeout),
			sleepyMiddleware(sleep),
		},
		[]handler{
			{
				"/",
				getHandler(t, "Hi"),
			},
		},
	)
}

func TestTimeoutValid(t *testing.T) {
	server := serverWithTimeout(t, time.Second, 0)

	if status := getRequestCode(server.URL, t); status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}
}

func TestTimeoutExceeded(t *testing.T) {
	server := serverWithTimeout(t, time.Second, time.Second*2)

	if status := getRequestCode(server.URL, t); status != http.StatusGatewayTimeout {
		t.Errorf("Expected status: %d, but got: %d",
			http.StatusGatewayTimeout, status,
		)
	}
}

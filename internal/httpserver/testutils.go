package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func emptyRequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), Logger, log)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			r = r.WithContext(ctx)
			next.ServeHTTP(ww, r)
		})
	}
}

func getHandler(t *testing.T, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(message))
		if err != nil {
			t.Errorf("Failed to send response: %s", err)
		}
	})
}

type handler struct {
	path string
	h    http.Handler
}

func getServerUnstarted(
	middlewares []func(http.Handler) http.Handler,
	handlers []handler,
) *httptest.Server {
	router := chi.NewRouter()

	log := logger.New(config.EnvLocal)
	router.Use(emptyRequestLogger(log))
	for _, f := range middlewares {
		router.Use(f)
	}

	for _, h := range handlers {
		router.Handle(h.path, h.h)
	}

	return httptest.NewUnstartedServer(router)
}

func getServer(
	middlewares []func(http.Handler) http.Handler,
	handlers []handler,
) *httptest.Server {
	srv := getServerUnstarted(middlewares, handlers)
	srv.Start()
	return srv
}

func getRequest(url string, t *testing.T) (int, string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("Failed to read response: %s", err)
	}

	return resp.StatusCode, string(body)
}

func getRequestCode(url string, t *testing.T) int {
	status, _ := getRequest(url, t)
	return status
}

func serverWithAddr(t *testing.T, url string, hs []handler) *httptest.Server {
	listener, err := net.Listen("tcp", url)
	if err != nil {
		t.Errorf("Error while creating server: %s", err)
	}

	server := getServerUnstarted(
		[]func(http.Handler) http.Handler{
			emptyRequestLogger(logger.New(config.EnvLocal)),
		},
		hs,
	)
	t.Logf("listener: %s, new listener: %s", server.Listener.Addr(), listener.Addr())

	server.Listener.Close()
	server.Listener = listener
	server.Start()

	return server
}

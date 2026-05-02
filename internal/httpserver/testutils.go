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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewEmptyRequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), Logger, log)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			r = r.WithContext(ctx)
			next.ServeHTTP(ww, r)
		})
	}
}

func NewHandler(t *testing.T, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(message))
		if err != nil {
			t.Errorf("Failed to send response: %s", err)
		}
	})
}

func NewSleepyHandler(t *testing.T, message string, sleepTime time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepTime)
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

func NewServerUnstarted(
	middlewares []func(http.Handler) http.Handler,
	handlers []handler,
) *httptest.Server {
	router := chi.NewRouter()

	log := logger.New(config.EnvLocal)
	router.Use(NewEmptyRequestLogger(log))
	for _, f := range middlewares {
		router.Use(f)
	}

	for _, h := range handlers {
		router.Handle(h.path, h.h)
	}

	return httptest.NewUnstartedServer(router)
}

func NewServer(
	middlewares []func(http.Handler) http.Handler,
	handlers []handler,
) *httptest.Server {
	srv := NewServerUnstarted(middlewares, handlers)
	srv.Start()
	return srv
}

func DoGet(url string, t *testing.T) (int, string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("Failed to close body: %s", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response: %s", err)
	}

	return resp.StatusCode, string(body)
}

func DoGetCode(url string, t *testing.T) int {
	status, _ := DoGet(url, t)
	return status
}

func NewServerWithAddr(t *testing.T, url string, hs []handler) *httptest.Server {
	listener, err := net.Listen("tcp", url)
	if err != nil {
		t.Errorf("Error while creating server: %s", err)
	}

	server := NewServerUnstarted(
		[]func(http.Handler) http.Handler{
			NewEmptyRequestLogger(logger.New(config.EnvLocal)),
		},
		hs,
	)

	err = server.Listener.Close()
	if err != nil {
		t.Errorf("Failed to close listener: %s", err)
	}
	server.Listener = listener
	server.Start()

	return server
}

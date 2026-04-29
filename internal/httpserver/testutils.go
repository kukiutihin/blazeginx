package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"context"
	"log/slog"
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

func getServer(t *testing.T, fs ...func(http.Handler) http.Handler) *httptest.Server {
	router := chi.NewRouter()

	log := logger.New(config.EnvLocal)
	router.Use(emptyRequestLogger(log))
	for _, f := range fs {
		router.Use(f)
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Hi"))
		if err != nil {
			t.Errorf("Failed to send response: %s", err)
		}
	})
	return httptest.NewServer(router)
}

func getRequest(server *httptest.Server, t *testing.T) int {
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}
	return resp.StatusCode
}

package testutils

import (
	"blazeginx/internal/config"
	"blazeginx/internal/httpserver/requestctx"
	"blazeginx/internal/logger"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newEmptyRequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := requestctx.WithLogger(r.Context(), log)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			r = r.WithContext(ctx)
			next.ServeHTTP(ww, r)
		})
	}
}

type Handler struct {
	Path string
	H    http.Handler
}

func NewServerUnstarted(
	middlewares []func(http.Handler) http.Handler,
	handlers []Handler,
) *httptest.Server {
	router := chi.NewRouter()

	log := logger.New(config.EnvLocal)
	router.Use(newEmptyRequestLogger(log))
	for _, f := range middlewares {
		router.Use(f)
	}

	for _, h := range handlers {
		router.Handle(h.Path, h.H)
	}

	return httptest.NewUnstartedServer(router)
}

func NewServer(
	handlers []Handler,
) *httptest.Server {
	srv := NewServerUnstarted([]func(http.Handler) http.Handler{}, handlers)
	srv.Start()
	return srv
}

// func NewServerWithMiddlewares(
// 	middlewares []func(http.Handler) http.Handler,
// 	handlers []Handler,
// ) *httptest.Server {
// 	srv := NewServerUnstarted(middlewares, handlers)
// 	srv.Start()
// 	return srv
// }

func NewServerWithAddr(t *testing.T, url string, hs []Handler) *httptest.Server {
	listener, err := net.Listen("tcp", url)
	if err != nil {
		t.Errorf("Error while creating server: %s", err)
	}

	server := NewServerUnstarted(
		[]func(http.Handler) http.Handler{
			newEmptyRequestLogger(logger.New(config.EnvLocal)),
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

func NewHandler(t *testing.T, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(message))
		if err != nil {
			t.Errorf("Failed to send response: %s", err)
		}
	})
}

func handleResponse(resp *http.Response, t *testing.T) (int, string) {
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

func DoGet(url string, t *testing.T) (int, string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}
	return handleResponse(resp, t)
}

func DoPost(url string, t *testing.T) (int, string) {
	r := strings.NewReader("")
	resp, err := http.Post(url, "", r)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}
	return handleResponse(resp, t)
}

func DoPut(url string, t *testing.T) (int, string) {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPut, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}

	return handleResponse(resp, t)
}

func DoDelete(url string, t *testing.T) (int, string) {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected http response, but got an error: %s", err)
	}

	return handleResponse(resp, t)
}

// func NewSleepyHandler(t *testing.T, message string, sleepTime time.Duration) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		time.Sleep(sleepTime)
// 		w.WriteHeader(http.StatusOK)
// 		_, err := w.Write([]byte(message))
// 		if err != nil {
// 			t.Errorf("Failed to send response: %s", err)
// 		}
// 	})
// }
//
//
//
// func DoGetCode(url string, t *testing.T) int {
// 	status, _ := DoGet(url, t)
// 	return status
// }

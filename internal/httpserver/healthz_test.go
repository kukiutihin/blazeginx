package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func serverWithHealthz(routes []config.Route, timeout time.Duration) *httptest.Server {
	return NewServer(
		[]func(http.Handler) http.Handler{
			NewEmptyRequestLogger(logger.New(config.EnvLocal)),
		},
		[]handler{
			{
				"/healthz",
				NewHealthzHandler(routes, timeout),
			},
		},
	)
}

func decodeResponse(t *testing.T, resp string) healthResponse {
	var res healthResponse
	err := json.Unmarshal([]byte(resp), &res)
	if err != nil {
		t.Errorf("Failed to parse response: %s", err)
	}
	return res
}

func TestHealthValid(t *testing.T) {
	t.Parallel()
	upstream1 := NewServerWithAddr(t, "127.0.0.1:15233", []handler{
		{
			"/scarymse",
			NewHandler(t, "ok1"),
		},
	})
	defer upstream1.Close()

	upstream2 := NewServerWithAddr(t, "127.0.0.1:15677", []handler{
		{
			"/healthz",
			NewHandler(t, "ok2"),
		},
	})
	defer upstream2.Close()

	routes := []config.Route{
		{
			Name:       "1st",
			Path:       "/some",
			Url:        "http://127.0.0.1:15233",
			HealthPath: "/scarymse",
		},
		{
			Name:       "2nd",
			Path:       "/other",
			Url:        "http://127.0.0.1:15677",
			HealthPath: "/healthz",
		},
	}

	srv := serverWithHealthz(routes, time.Second*3)

	want := healthResponse{
		Message: "OK",
		UpstreamResponses: []healthUpstreamResponse{
			{
				Name:   "1st",
				Status: http.StatusOK,
				Body:   "ok1",
			},
			{
				Name:   "2nd",
				Status: http.StatusOK,
				Body:   "ok2",
			},
		},
	}

	status, body := DoGet(srv.URL+"/healthz", t)
	if status != http.StatusOK {
		t.Errorf("Expected status: %d, but got: %d", http.StatusOK, status)
	}

	decoded := decodeResponse(t, body)
	if diff := cmp.Diff(want, decoded,
		cmpopts.IgnoreFields(healthUpstreamResponse{}, "ResponseTimeMs"),
	); diff != "" {
		t.Fatalf("Mismatch: %s", diff)
	}
}

func TestHealthNotExists(t *testing.T) {
	t.Parallel()
	upstream := NewServerWithAddr(t, "127.0.0.1:15521", []handler{})
	defer upstream.Close()

	routes := []config.Route{
		{
			Name:       "serv",
			Path:       "/some",
			Url:        "http://127.0.0.1:15521",
			HealthPath: "/healthz",
		},
	}

	srv := serverWithHealthz(routes, time.Second*3)

	want := healthResponse{
		Message: "1 upstream requests failed",
		UpstreamResponses: []healthUpstreamResponse{
			{
				Name:   "serv",
				Status: http.StatusNotFound,
				Body:   "404 page not found\n",
			},
		},
	}

	status, body := DoGet(srv.URL+"/healthz", t)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status: %d, but got: %d", http.StatusInternalServerError, status)
	}

	decoded := decodeResponse(t, body)
	if diff := cmp.Diff(want, decoded,
		cmpopts.IgnoreFields(healthUpstreamResponse{}, "ResponseTimeMs"),
	); diff != "" {
		t.Fatalf("Mismatch: %s", diff)
	}
}

func TestHealthTimeout(t *testing.T) {
	t.Parallel()
	upstream1 := NewServerWithAddr(t, "127.0.0.1:15522", []handler{
		{
			"/healthz",
			NewSleepyHandler(t, "finalle", time.Second*2),
		},
	})
	defer upstream1.Close()

	upstream2 := NewServerWithAddr(t, "127.0.0.1:15523", []handler{
		{
			"/healthz",
			NewSleepyHandler(t, "finalle", time.Second*4),
		},
	})
	defer upstream2.Close()

	routes := []config.Route{
		{
			Name:       "blazeman",
			Path:       "/some",
			Url:        "http://127.0.0.1:15522",
			HealthPath: "/healthz",
		},
		{
			Name:       "slowman",
			Path:       "/other",
			Url:        "http://127.0.0.1:15523",
			HealthPath: "/healthz",
		},
	}

	srv := serverWithHealthz(routes, time.Second*3)

	want := healthResponse{
		Message: "1 upstream requests failed",
		UpstreamResponses: []healthUpstreamResponse{
			{
				Name:   "blazeman",
				Status: http.StatusOK,
				Body:   "finalle",
			},
			{
				Name:  "slowman",
				Error: "timeout",
			},
		},
	}

	status, body := DoGet(srv.URL+"/healthz", t)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status: %d, but got: %d", http.StatusInternalServerError, status)
	}

	decoded := decodeResponse(t, body)
	if diff := cmp.Diff(want, decoded,
		cmpopts.IgnoreFields(healthUpstreamResponse{}, "ResponseTimeMs"),
	); diff != "" {
		t.Fatalf("Mismatch: %s", diff)
	}
}

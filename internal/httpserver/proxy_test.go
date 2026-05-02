package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func serverWithProxys(upstreamRoutes []config.Route) *httptest.Server {
	proxys := make([]handler, 0)
	for _, r := range upstreamRoutes {
		proxys = append(proxys,
			handler{
				NormalizeProxyRoute(r.Path),
				BuildProxy(r, time.Second*1),
			},
		)
	}

	return NewServer(
		[]func(http.Handler) http.Handler{
			NewEmptyRequestLogger(logger.New(config.EnvLocal)),
		},
		proxys,
	)
}

func TestProxyValid(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1337", []handler{
		{
			"/hi/hello/", NewHandler(t, "keks"),
		},
		{
			"/chelik/", NewHandler(t, "kekas"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/",
		Url:         "http://127.0.0.1:1337",
		StripPrefix: false,
	}

	route2 := config.Route{
		Path:        "/chelik/",
		Url:         "http://127.0.0.1:1337",
		StripPrefix: false,
	}

	prox := serverWithProxys([]config.Route{route1, route2})

	status, body := DoGet(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, body = DoGet(prox.URL+"/chelik/", t)
	if body != "kekas" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "kekas", body, status)
	}
}

func TestProxyNonExistsHandlers(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1400", []handler{
		{
			"/chelik/", NewHandler(t, "kekas"),
		},
	})
	srv.Close()

	route2 := config.Route{
		Path:        "/chelik/",
		Url:         "http://127.0.0.1:1400",
		StripPrefix: false,
	}

	prox := serverWithProxys([]config.Route{route2})

	status, _ := DoGet(prox.URL+"/skibidu/", t)
	if status != http.StatusNotFound {
		t.Errorf("Expected: %d, but got: %d", http.StatusNotFound, status)
	}
}

func TestUglyPaths(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1388", []handler{
		{
			"/hi/hello/", NewHandler(t, "keks"),
		},
		{
			"/chelik/", NewHandler(t, "kekas"),
		},
		{
			"/haskelllord/", NewHandler(t, "dm"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "hi/",
		Url:         "http://127.0.0.1:1388",
		StripPrefix: false,
	}

	route2 := config.Route{
		Path:        "chelik",
		Url:         "http://127.0.0.1:1388",
		StripPrefix: false,
	}

	route3 := config.Route{
		Path:        "/haskelllord",
		Url:         "http://127.0.0.1:1388",
		StripPrefix: false,
	}

	prox := serverWithProxys([]config.Route{route1, route2, route3})

	status, body := DoGet(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, body = DoGet(prox.URL+"/chelik/", t)
	if body != "kekas" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "kekas", body, status)
	}

	status, body = DoGet(prox.URL+"/haskelllord/", t)
	if body != "dm" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "dm", body, status)
	}
}

func TestStripPrefix(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1355", []handler{
		{
			"/hello/", NewHandler(t, "keks"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/",
		Url:         "http://127.0.0.1:1355",
		StripPrefix: true,
	}

	prox := serverWithProxys([]config.Route{route1})

	status, body := DoGet(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}
}

func TestStripLongPrefix(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1356", []handler{
		{
			"/cool/", NewHandler(t, "keks"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/howareyou/imok/",
		Url:         "http://127.0.0.1:1356",
		StripPrefix: true,
	}

	prox := serverWithProxys([]config.Route{route1})

	status, body := DoGet(prox.URL+"/hi/howareyou/imok/cool/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}
}

func TestUpstreamTimeout(t *testing.T) {
	t.Parallel()
	srv := NewServerWithAddr(t, "127.0.0.1:1357", []handler{
		{
			"/fast/", NewHandler(t, "keks"),
		},
		{
			"/slow/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second)
			}),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/",
		Url:         "http://127.0.0.1:1357",
		StripPrefix: true,
	}

	prox := serverWithProxys([]config.Route{route1})

	status, body := DoGet(prox.URL+"/hi/fast/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, _ = DoGet(prox.URL+"/hi/slow/", t)
	if status != http.StatusGatewayTimeout {
		t.Errorf("Expected: %d, but got: %d", http.StatusGatewayTimeout, status)
	}
}

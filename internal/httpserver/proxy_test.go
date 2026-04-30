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
				CanonicalChiRoute(r.Path),
				BuildProxy(r, time.Second*1),
			},
		)
	}

	return getServer(
		[]func(http.Handler) http.Handler{
			emptyRequestLogger(logger.New(config.EnvLocal)),
		},
		proxys,
	)
}

func TestProxyValid(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1337", []handler{
		{
			"/hi/hello/", getHandler(t, "keks"),
		},
		{
			"/chelik/", getHandler(t, "kekas"),
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

	status, body := getRequest(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, body = getRequest(prox.URL+"/chelik/", t)
	if body != "kekas" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "kekas", body, status)
	}
}

func TestProxyNonExistsHandlers(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1400", []handler{
		{
			"/chelik/", getHandler(t, "kekas"),
		},
	})
	srv.Close()

	route2 := config.Route{
		Path:        "/chelik/",
		Url:         "http://127.0.0.1:1400",
		StripPrefix: false,
	}

	prox := serverWithProxys([]config.Route{route2})

	status, _ := getRequest(prox.URL+"/skibidu/", t)
	if status != http.StatusNotFound {
		t.Errorf("Expected: %d, but got: %d", http.StatusNotFound, status)
	}
}

func TestUglyPaths(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1388", []handler{
		{
			"/hi/hello/", getHandler(t, "keks"),
		},
		{
			"/chelik/", getHandler(t, "kekas"),
		},
		{
			"/haskelllord/", getHandler(t, "dm"),
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

	status, body := getRequest(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, body = getRequest(prox.URL+"/chelik/", t)
	if body != "kekas" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "kekas", body, status)
	}

	status, body = getRequest(prox.URL+"/haskelllord/", t)
	if body != "dm" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "dm", body, status)
	}
}

func TestStripPrefix(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1355", []handler{
		{
			"/hello/", getHandler(t, "keks"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/",
		Url:         "http://127.0.0.1:1355",
		StripPrefix: true,
	}

	prox := serverWithProxys([]config.Route{route1})

	status, body := getRequest(prox.URL+"/hi/hello/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}
}

func TestStripLongPrefix(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1356", []handler{
		{
			"/cool/", getHandler(t, "keks"),
		},
	})
	defer srv.Close()

	route1 := config.Route{
		Path:        "/hi/howareyou/imok/",
		Url:         "http://127.0.0.1:1356",
		StripPrefix: true,
	}

	prox := serverWithProxys([]config.Route{route1})

	status, body := getRequest(prox.URL+"/hi/howareyou/imok/cool/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}
}

func TestUpstreamTimeout(t *testing.T) {
	t.Parallel()
	srv := serverWithAddr(t, "127.0.0.1:1357", []handler{
		{
			"/fast/", getHandler(t, "keks"),
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

	status, body := getRequest(prox.URL+"/hi/fast/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, _ = getRequest(prox.URL+"/hi/slow/", t)
	if status != http.StatusGatewayTimeout {
		t.Errorf("Expected: %d, but got: %d", http.StatusGatewayTimeout, status)
	}
}

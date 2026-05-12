package proxy

import (
	"blazeginx/internal/config"
	"blazeginx/internal/routing"
	tt "blazeginx/internal/testutils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type fakeLoadBalancer struct {
	u *url.URL
}

func (lb fakeLoadBalancer) Next() *url.URL {
	return lb.u
}

var transportST http.Transport = http.Transport{
	ResponseHeaderTimeout: time.Hour,
	TLSHandshakeTimeout:   time.Hour,
	ExpectContinueTimeout: time.Hour,
	MaxConnsPerHost:       100,
	MaxIdleConnsPerHost:   100,
}

func newServerWithProxy(services []config.Service) *httptest.Server {
	return newServerWithProxyCustomTS(services, &transportST)
}

func newServerWithProxyCustomTS(
	services []config.Service,
	ts *http.Transport,
) *httptest.Server {
	handlers := make([]tt.Handler, 0)
	for _, s := range services {
		path, _ := routing.NormalizeRoutePath(s.Path)
		prox := ReverseProxy{
			LoadBalancer: fakeLoadBalancer{u: s.Urls[0]},
			BufferPool:   nil,
			Transport:    ts, Service: &s,
			Metrics: nil,
		}
		handlers = append(handlers, tt.Handler{
			Path: routing.Wildcard(path),
			H:    prox.New(),
		})
	}
	return tt.NewServer(handlers)
}

func TestProxyValid(t *testing.T) {
	t.Parallel()
	srv := tt.NewServerWithAddr(t, "127.0.0.1:1337", []tt.Handler{
		{
			Path: "/hi/hello/",
			H:    tt.NewHandler(t, "keks"),
		},
		{
			Path: "/chelik/",
			H:    tt.NewHandler(t, "kekas"),
		},
	})
	defer srv.Close()

	u, _ := url.Parse("http://127.0.0.1:1337")

	s1 := config.Service{
		Path:        "/hi/",
		Urls:        []*url.URL{u},
		StripPrefix: false,
	}

	s2 := config.Service{
		Path:        "/chelik/",
		Urls:        []*url.URL{u},
		StripPrefix: false,
	}

	prox := newServerWithProxy([]config.Service{s1, s2})

	status, body := tt.DoGet(prox.URL+"/hi/hello/", t)
	if body != "keks" || status != http.StatusOK {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, body = tt.DoGet(prox.URL+"/chelik/", t)
	if body != "kekas" || status != http.StatusOK {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "kekas", body, status)
	}
}

func TestProxyForwardsMethod(t *testing.T) {
	t.Parallel()
	srv := tt.NewServerWithAddr(t, "127.0.0.1:1338", []tt.Handler{
		{
			Path: "/",
			H: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(r.Method))
				if err != nil {
					t.Errorf("Failed to send response: %s", err)
				}
			}),
		},
	})
	defer srv.Close()

	u, _ := url.Parse("http://127.0.0.1:1338")
	s := config.Service{
		Path:        "/",
		Urls:        []*url.URL{u},
		StripPrefix: false,
	}

	prox := newServerWithProxy([]config.Service{s})

	_, body := tt.DoGet(prox.URL+"/", t)
	if body != "GET" {
		t.Errorf("Expected: GET, but got: %s", body)
	}

	_, body = tt.DoPost(prox.URL+"/", t)
	if body != "POST" {
		t.Errorf("Expected: POST, but got: %s", body)
	}

	_, body = tt.DoPut(prox.URL+"/", t)
	if body != "PUT" {
		t.Errorf("Expected: PUT, but got: %s", body)
	}

	_, body = tt.DoDelete(prox.URL+"/", t)
	if body != "DELETE" {
		t.Errorf("Expected: DELETE, but got: %s", body)
	}
}

func TestStripPrefix(t *testing.T) {
	t.Parallel()
	srv := tt.NewServerWithAddr(t, "127.0.0.1:1340", []tt.Handler{
		{
			Path: "/hello/",
			H:    tt.NewHandler(t, "keks1"),
		},
		{
			Path: "/cool/",
			H:    tt.NewHandler(t, "keks2"),
		},
		{
			Path: "/",
			H:    tt.NewHandler(t, "keks3"),
		},
	})
	defer srv.Close()

	u, _ := url.Parse("http://127.0.0.1:1340")
	s1 := config.Service{
		Path:        "/hi/",
		Urls:        []*url.URL{u},
		StripPrefix: true,
	}

	s2 := config.Service{
		Path:        "/hi/howareyou/imok/",
		Urls:        []*url.URL{u},
		StripPrefix: true,
	}

	s3 := config.Service{
		Path:        "/kekason",
		Urls:        []*url.URL{u},
		StripPrefix: true,
	}

	prox := newServerWithProxy([]config.Service{s1, s2, s3})

	status, body := tt.DoGet(prox.URL+"/hi/hello/", t)
	if body != "keks1" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks1", body, status)
	}

	status, body = tt.DoGet(prox.URL+"/hi/howareyou/imok/cool/", t)
	if body != "keks2" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks2", body, status)
	}

	status, body = tt.DoGet(prox.URL+"/kekason", t)
	if body != "keks3" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks3", body, status)
	}
}

func TestProxyNonExistsHandlers(t *testing.T) {
	t.Parallel()
	srv := tt.NewServerWithAddr(t, "127.0.0.1:1339", []tt.Handler{
		{
			Path: "/chelik/",
			H:    tt.NewHandler(t, "kekas"),
		},
	})
	srv.Close()

	u, _ := url.Parse("http://127.0.0.1:1339")
	s := config.Service{
		Path:        "/chelik/",
		Urls:        []*url.URL{u},
		StripPrefix: false,
	}

	prox := newServerWithProxy([]config.Service{s})

	status, _ := tt.DoGet(prox.URL+"/skibidu/", t)
	if status != http.StatusNotFound {
		t.Errorf("Expected: %d, but got: %d", http.StatusNotFound, status)
	}
}

func TestUpstreamTimeout(t *testing.T) {
	t.Parallel()
	srv := tt.NewServerWithAddr(t, "127.0.0.1:1357", []tt.Handler{
		{
			Path: "/fast/",
			H:    tt.NewHandler(t, "keks"),
		},
		{
			Path: "/slow/",
			H: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second * 2)
			}),
		},
	})
	defer srv.Close()

	u, _ := url.Parse("http://127.0.0.1:1357")
	s := config.Service{
		Path:        "/hi/",
		Urls:        []*url.URL{u},
		StripPrefix: true,
	}

	prox := newServerWithProxyCustomTS([]config.Service{s}, &http.Transport{
		ResponseHeaderTimeout: time.Second * 1,
	})

	status, body := tt.DoGet(prox.URL+"/hi/fast/", t)
	if body != "keks" {
		t.Errorf("Expected: %s, but got: %s, with status: %d", "keks", body, status)
	}

	status, _ = tt.DoGet(prox.URL+"/hi/slow/", t)
	if status != http.StatusGatewayTimeout {
		t.Errorf("Expected: %d, but got: %d", http.StatusGatewayTimeout, status)
	}
}

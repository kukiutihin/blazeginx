package proxy

import (
	"blazeginx/internal/config"
	"blazeginx/internal/httpserver/loadbalancer"
	"blazeginx/internal/httpserver/metrics"
	"blazeginx/internal/httpserver/requestctx"
	"blazeginx/internal/routing"
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
)

// http.Transport{
// 			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
// 			TLSHandshakeTimeout:   5 * time.Second,
// 			ExpectContinueTimeout: 1 * time.Second,
// 			ForceAttemptHTTP2:     true,
// 			MaxConnsPerHost:       int(cfg.MaxConnsPerHost),
// 			MaxIdleConnsPerHost:   int(cfg.MaxIdleConnsPerHost),
// 		},

type ReverseProxy struct {
	LoadBalancer loadbalancer.LoadBalancer
	BufferPool   httputil.BufferPool
	Transport    http.RoundTripper
	Service      *config.Service
	Metrics      *metrics.Metrics
}

func (rp *ReverseProxy) New() http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			log := requestctx.Logger(pr.In.Context()).With(
				"component", "proxy",
				"service", rp.Service.Name,
			)

			if rp.Service.StripPrefix {
				pr.Out.URL.Path = routing.StripPrefix(pr.In.URL.Path, rp.Service.Path)
			}

			endpoint := rp.LoadBalancer.Next()
			pr.Out.URL.Host = endpoint.Host
			pr.Out.URL.Scheme = endpoint.Scheme
			pr.SetXForwarded()

			log.Debug("request prepared",
				"upstream_url", pr.Out.URL.String(),
			)
		},

		ModifyResponse: func(r *http.Response) error {
			log := requestctx.Logger(r.Request.Context()).With(
				"component", "proxy",
				"service", rp.Service.Name,
				"upstream_url", r.Request.URL.String(),
			)

			// TODO: metrics
			// TODO: cache

			log.Debug("Received response from upstream",
				"status", r.Status,
			)
			return nil
		},

		Transport:  rp.Transport,
		BufferPool: rp.BufferPool,

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log := requestctx.Logger(r.Context()).With(
				"component", "proxy",
				"service", rp.Service.Name,
			)

			status := http.StatusBadGateway

			switch {
			case errors.Is(err, context.DeadlineExceeded):
				status = http.StatusGatewayTimeout
				log.Warn("Timeout exceeded",
					"error", err,
				)
			default:
				log.Warn("Undefined error occured",
					"error", err,
				)
			}
			http.Error(w, http.StatusText(status), status)
		},
	}
}

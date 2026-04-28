package httpserver

import (
	"blazeginx/internal/config"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func BuildProxy(
	route config.Route,
	addr string,
	upstreamTimeout time.Duration,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		url := CreateUrl(route, addr, r.URL.Path)

		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}

		log := ctx.Value(Logger).(*slog.Logger)
		log.Debug("Received in proxy", "url", url)

		// Context with upstream timeout
		upstreamCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(upstreamCtx, r.Method, url, r.Body)
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			log.Error(
				"Failed to build request for service",
				"url", url,
				"error", err,
			)
			return
		}

		// Timeout middleware will return GatewayTimeout response itself
		if IsDone(ctx) {
			return
		}

		req.Header = r.Header.Clone()

		resp, err := http.DefaultClient.Do(req)
		if upstreamCtx.Err() == context.DeadlineExceeded {
			log.Warn(
				"Service timeout exceeded",
				"url", url,
			)

			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}

		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			log.Error(
				"Failed to send request to service",
				"url", url,
				"error", err,
			)
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Error(
					"Failed to close connection",
					"error", err,
				)
			}
		}()

		log.Debug(
			"Received a response from service",
			"url", url,
			"status", resp.Status,
		)

		if IsDone(ctx) {
			return
		}

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		if IsDone(ctx) {
			return
		}

		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			log.Error(
				"Failed to send response",
				"error", err,
			)
		}
	}
}

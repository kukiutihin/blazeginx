package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type upstreamContext struct {
	ResponseWriter http.ResponseWriter
	Log            *slog.Logger
	Upstream       string
}

func handleUpstreamInternalErr(message string, err error, ctx upstreamContext) {
	http.Error(ctx.ResponseWriter,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
	ctx.Log.Error(message,
		"upstream_name", ctx.Upstream,
		"error", err,
	)
}

func BuildProxy(route config.Route, upstreamTimeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.GetServiceLogger(ctx.Value(Logger).(*slog.Logger), "proxy")
		uctx := upstreamContext{
			ResponseWriter: w,
			Log:            log,
			Upstream:       route.Name,
		}

		url, err := CreateFullUrl(route, r.URL.Path, r.URL.RawQuery)
		if err != nil {
			handleUpstreamInternalErr("Failed to create request to upstream", err, uctx)
			return
		}

		// Context with upstream timeout
		upstreamCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(upstreamCtx, r.Method, url, r.Body)
		if err != nil {
			handleUpstreamInternalErr("Failed to create request to upstream", err, uctx)
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
			handleUpstreamInternalErr("Failed to send request to upstream", err, uctx)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

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
	})
}

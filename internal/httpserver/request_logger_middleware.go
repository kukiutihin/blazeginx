package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger is middleware which passes the request info with other logs from this request handlers
func RequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("total_time", time.Since(t1).String()),
				)
			}()

			ctx := context.WithValue(r.Context(), Logger, entry)
			r = r.WithContext(ctx)

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

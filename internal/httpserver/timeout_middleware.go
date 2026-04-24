package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Timeout takes Duration and returns middleware which passes context.WithTimeout
// and checks it. If context.DeadlineExceeded writes 504
func Timeout(t time.Duration, log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), t)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
					log.Warn("Request declined, server timeout exceeded")
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

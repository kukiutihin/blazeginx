package httpserver

import (
	"blazeginx/internal/config"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func GetRouter(log *slog.Logger, cfg config.Config) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(RequestLogger(log))

	// TODO: /metrics handler

	// TODO: /healthz handler

	// group of handlers for proxying
	router.Group(func(r chi.Router) {
		tokens := NewTokenBucket(cfg.RateLimit)
		r.Use(tokens.BuildRateLimiter())

		r.Use(Timeout(cfg.Timeout.Server, log))

		// Config must be valid. config.MustRead() always valid
		for _, c := range cfg.Routes {
			r.HandleFunc(
				CanonicalRoute(c.Path),
				BuildProxy(c, cfg.ServiceMap[c.Service], cfg.Timeout.Upstream),
			)
		}
	})

	// TODO: static fallback

	return router
}

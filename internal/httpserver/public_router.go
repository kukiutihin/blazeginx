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

	// group of handlers for proxying
	router.Group(func(r chi.Router) {
		tokens := NewTokenBucket(cfg.RateLimit)
		if cfg.RateLimit.Enabled {
			r.Use(tokens.BuildRateLimiter())
		}

		r.Use(Timeout(cfg.Timeout.Server))

		// Config must be valid. config.MustRead() always valid
		for _, c := range cfg.Routes {
			r.Handle(
				NormalizeProxyRoute(c.Path),
				BuildProxy(c, cfg.Timeout.Upstream),
			)
		}
	})

	if cfg.Static.Enabled {
		router.NotFound(NewStaticHandler(cfg.Static).ServeHTTP)
	}

	return router
}

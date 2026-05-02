package httpserver

import (
	"blazeginx/internal/config"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func GetLocalRouter(log *slog.Logger, cfg config.Config) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(RequestLogger(log))

	router.Handle("/healthz", NewHealthzHandler(cfg.Routes, cfg.Timeout.Upstream))
	router.Handle("/healthz/", NewHealthzHandler(cfg.Routes, cfg.Timeout.Upstream))

	return router
}

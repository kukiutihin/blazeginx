package logger

import (
	"blazeginx/internal/config"
	"log/slog"
	"os"
)

func New(env config.Env) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		cfg := slog.HandlerOptions{Level: slog.LevelDebug}
		log = slog.New(slog.NewTextHandler(os.Stdout, &cfg))

	case config.EnvDev:
		cfg := slog.HandlerOptions{Level: slog.LevelDebug}
		log = slog.New(slog.NewJSONHandler(os.Stdout, &cfg))

	case config.EnvProd:
		cfg := slog.HandlerOptions{Level: slog.LevelInfo}
		log = slog.New(slog.NewJSONHandler(os.Stdout, &cfg))
	}

	return log
}

package application

import (
	"blazeginx/internal/config"
	"blazeginx/internal/httpserver"
	"blazeginx/internal/logger"

	"net/http"
	"os"
)

func Start() {
	config := config.MustRead()
	log := logger.New(config.Env)
	localLog := logger.New(config.Env)

	log.Info(
		"server started",
		"env", config.Env,
		"addr", config.Addr,
	)

	local := &http.Server{
		Addr:        config.AdminAddr,
		Handler:     httpserver.GetLocalRouter(localLog, config),
		IdleTimeout: config.Timeout.Idle,
	}

	go func() {
		if err := local.ListenAndServe(); err != nil {
			log.Error("Failed to start server", "error", err.Error())
			os.Exit(1)
		}
	}()

	public := &http.Server{
		Addr:        config.Addr,
		Handler:     httpserver.GetRouter(log, config),
		IdleTimeout: config.Timeout.Idle,
	}

	if err := public.ListenAndServe(); err != nil {
		log.Error("Failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

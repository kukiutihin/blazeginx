package application

import (
	"blazeginx/internal/config"
	"blazeginx/internal/httpserver"
	"blazeginx/internal/logger"

	"net/http"
	"os"
	"strconv"
)

func Start() {
	config := config.MustRead()
	log := logger.New(config.Env)
	log.Info(
		"server started",
		"env", config.Env,
		"port", config.Port,
	)

	srv := &http.Server{
		Addr:        strconv.FormatUint(uint64(config.Port), 10),
		Handler:     httpserver.GetRouter(log, config),
		IdleTimeout: config.Timeout.Idle,
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Error("Failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

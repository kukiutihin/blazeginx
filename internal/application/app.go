package application

import (
	"blazeginx/internal/config"
	mymiddleware "blazeginx/internal/httpserver/middleware"
	"blazeginx/internal/logger"
	"blazeginx/internal/proxy"
	"blazeginx/internal/httpserver"

	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start() {
	config := config.MustRead()
	log := logger.New(config.Env)
	log.Info(
		"server started",
		"env", config.Env,
		"port", config.Port,
	)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mymiddleware.RequestLogger(log))

    // TODO: /metrics handler
    
    // TODO: /healthz handler

    // group of handlers for proxying 
    router.Group(func(r chi.Router) {

        // TODO: ratelimit middleware 
        // TODO: timeout middleware

        // config must be valid. config.MustRead() always valid
        for _, c := range config.Routes {
            r.HandleFunc(
                httpserver.CanonicalRoute(c.Path), 
                proxy.BuildHandler(config.ServiceMap[c.Service]), 
            )
        }
    }) 

    // TODO: static fallback

	err := http.ListenAndServe(":"+
		strconv.FormatUint(uint64(config.Port), 10), 
        router,
	)
	if err != nil {
		log.Error("Failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

package main 

import (
	"blazeginx/internal/config"
    "blazeginx/internal/logger"
)

func main() {
    config := config.ReadConfig()
    log := logger.GetLogger(config.Env)
    log.Info(
        "server started",
        "env", config.Env,
        "port", config.Port,
    )
}

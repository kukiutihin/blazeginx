package main 

import (
	"blazeginx/internal/config"
	"fmt"
)

func main() {
    config := config.ReadConfig()
    fmt.Println(config)
}

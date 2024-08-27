package main

import (
	"log"
	"my_project/client/config"
	"my_project/client/service"
)

func main() {
    config.InitConfig()
    config.InitLog()
    
    ch, err := config.InitRabbitMQ()
    if err != nil {
        log.Fatalf("Failed to initialize RabbitMQ: %v", err)
        return
    }
    defer ch.Close()

    service.StartConsuming(ch)
    
    select {}
}


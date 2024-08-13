package main

import (
    "log"
    "my_project/server/config"
    "my_project/server/internal/common"
    "my_project/server/internal/router"
    "my_project/server/internal/service"
    "net/http"
    "github.com/cloudwego/hertz/pkg/app/client"
)

func main() {
    config.Init()

    hertzClient, err := client.NewClient()
    if err != nil {
        log.Fatalf("Failed to create Hertz client: %v", err)
    }

    r := router.InitializeRoutes()

    go service.RunService(common.MySQLDB, hertzClient)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

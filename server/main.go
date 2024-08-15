package main

import (
    "log"
    "my_project/server/config"
    "my_project/server/internal/common"
    "my_project/server/internal/service"
    "my_project/server/internal/router"
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

    if common.ServerConfig.Mode == "simulation" {
        log.Println("Running in simulation mode")
        go service.RunService(common.MySQLDB, hertzClient)
    } else if common.ServerConfig.Mode == "real" {
        log.Println("Running in real probe mode")
        // 启动服务以接受真实探测任务的请求
    } else {
        log.Fatalf("Unknown mode: %s", common.ServerConfig.Mode)
    }

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
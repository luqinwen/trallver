package main

import (
    "log"
    "net/http"
    "my_project/config"
    "my_project/internal/common"
    "my_project/internal/router"
    "my_project/internal/service"

    "github.com/cloudwego/hertz/pkg/app/client"
)

func main() {
    config.Init()

    // 初始化 Hertz 客户端
    hertzClient, err := client.NewClient()
    if err != nil {
        log.Fatalf("Failed to create Hertz client: %v", err)
    }

    // 初始化路由
    r := router.InitializeRoutes()

    // 启动服务
    go service.RunService(common.MySQLDB, hertzClient)

    // 启动HTTP服务器
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

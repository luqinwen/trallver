package main

import (
    "log"
    "my_project/config"
    "my_project/internal/common"
    "my_project/internal/service"

    "github.com/cloudwego/hertz/pkg/app/client"
)

func main() {
    config.InitLog()

    // 初始化 Hertz 客户端
    hertzClient, err := client.NewClient()
    if err != nil {
        log.Fatalf("Failed to create Hertz client: %v", err)
    }

    service.RunService(common.DB, hertzClient)
}


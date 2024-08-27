package main

import (
	"log"
	"my_project/server/config"
	"my_project/server/internal/router"
	"my_project/server/internal/service"

	"github.com/cloudwego/hertz/pkg/app/server"
    "github.com/robfig/cron/v3"
)

func main() {
    // 初始化配置
    config.Init()

    // 初始化 RabbitMQ
    conn, ch, err := config.InitRabbitMQ() // 修改为接收三个返回值
    if err != nil {
        log.Fatalf("Failed to initialize RabbitMQ: %v", err)
        return
    }
    defer conn.Close() // 确保在程序退出时关闭 RabbitMQ 连接
    defer ch.Close()   // 确保在程序退出时关闭 RabbitMQ 通道


    // 初始化 Hertz 服务器，指定监听端口为 8080
    h := server.Default(server.WithHostPorts(":8080"))

    // 初始化路由并传递 RabbitMQ 通道
    router.InitializeRoutes(h, ch)

    c := cron.New(cron.WithSeconds())
    c.AddFunc("@every 1m", func() { // 每分钟发送一次任务
        log.Println("Automatically sending task to client")
        task := service.CreateProbeTask() // 使用统一的初始化函数

        if err := service.ProcessProbeTask(ch, task); err != nil {
            log.Printf("Failed to process task: %v", err)
        }
    })

    c.Start() // 启动 cron 调度

    go service.HandleProbeResults(ch)
    
    // 启动 Hertz 服务器
    log.Println("Starting Hertz server on :8080")
    h.Spin()

    // 阻止主协程退出
    select {}
}

package router

import (
    "my_project/server/internal/service"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/streadway/amqp"
)

func InitializeRoutes(h *server.Hertz, ch *amqp.Channel) {
    h.POST("/probe_task", service.HandleProbeTask(ch))       // 任务下发
}

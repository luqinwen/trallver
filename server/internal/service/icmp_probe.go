package service

import (
    "context"
    "log"
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/streadway/amqp"
)

// 全局自增ID，用于生成唯一的 taskID
var globalTaskID uint32 = 1

// 生成下一个任务ID
func getNextTaskID() uint32 {
    globalTaskID++
    return globalTaskID
}

// HandleProbeTask 用于通过 HTTP 请求触发探测任务
func HandleProbeTask(ch *amqp.Channel) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 生成唯一的 taskID
        taskID := getNextTaskID()

        log.Printf("Generated taskID: %d", taskID)

        // 调用 GenerateAndProcessTask 生成并处理任务
        err := GenerateAndProcessTask(ch, taskID)

        // 判断任务是否成功下发，并记录日志
        if err != nil {
            log.Printf("Failed to assign task with ID: %d, error: %v", taskID, err)
            c.String(500, "Failed to assign task")
        } else {
            log.Printf("Successfully assigned task with ID: %d", taskID)
            c.String(200, "Task assigned successfully")
        }
    }
}

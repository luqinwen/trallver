package service

import (
    "log"
    "my_project/server/internal/dao"
    "my_project/server/internal/model"
    "github.com/robfig/cron/v3"
    "github.com/streadway/amqp"
)

var fixedTask *model.ProbeTaskFixed  // 缓存首次任务的固定字段

// 启动 cron 任务
func StartCronJob(ch *amqp.Channel) {
    c := cron.New(cron.WithSeconds())
    c.AddFunc("@every 1m", func() {
        log.Println("Automatically sending task to client")

        if fixedTask == nil {
            initializeFixedTask()  // 初始化首次任务
        }

        // 生成并处理任务
        taskID, err := dao.GetNextTaskID()
        if err != nil {
            log.Fatalf("Failed to get next task ID: %v", err)
            return
        }

        if err := GenerateAndProcessTask(ch, taskID); err != nil {
            log.Printf("Failed to process task: %v", err)
        }
    })
    c.Start()
}

// 初始化首次固定任务
func initializeFixedTask() {
    log.Println("Generating first task with fixed fields")
    fixedTask = CreateFixedTask()

    fixedTaskID, err := dao.StoreFixedTask(fixedTask)
    if err != nil {
        log.Fatalf("Failed to store fixed task: %v", err)
        return
    }

    fixedTask.ID = fixedTaskID  // 设置固定字段的 ID
}

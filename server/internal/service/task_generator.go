package service

import (
    "log"
    "my_project/server/internal/dao"
    "my_project/server/internal/model"
    "github.com/streadway/amqp"
)

// 全局变量，用于缓存首次任务的固定字段
var cachedFixedTask *model.ProbeTaskFixed
var firstTaskSent bool = false

// GenerateAndProcessTask 生成任务并调用 ProcessProbeTask，返回 error 以便调用者处理错误
func GenerateAndProcessTask(ch *amqp.Channel, taskID uint32) error {
    var dynamicTask model.ProbeTaskDynamic
    var fixedTaskID uint32
    var err error

    if !firstTaskSent {
        log.Println("Generating first task with fixed fields")
    
        // 首次任务，创建固定字段并缓存
        fixedTask := CreateFixedTask()
        cachedFixedTask = fixedTask  // 缓存固定字段
    
        // 存储固定字段并获取固定字段的ID
        fixedTaskID, err = dao.StoreFixedTask(cachedFixedTask)
        if err != nil {
            log.Fatalf("Failed to store fixed task: %v", err)
            return err
        }
        cachedFixedTask.ID = fixedTaskID  // 将数据库生成的 ID 赋值给缓存的 fixedTask
    
        // 标记已生成首次任务
        firstTaskSent = true
    } else {
        // 如果不是首次任务，直接使用缓存的 fixedTaskID
        fixedTaskID = cachedFixedTask.ID
    }
    
    // 每次生成动态字段
    dynamicTask = CreateDynamicTask(taskID)

    // 调用 ProcessProbeTask 处理任务，使用缓存的固定字段
    err = ProcessProbeTask(ch, *cachedFixedTask, dynamicTask)
    if err != nil {
        log.Printf("Failed to process probe task: %v", err)
        return err
    }

    // 存储动态字段并传递 fixedTaskID
    err = dao.StoreProbeTask(&dynamicTask, fixedTaskID)
    if err != nil {
        log.Printf("Failed to store dynamic task: %v", err)
        return err
    }

    return nil  // 成功返回 nil
}
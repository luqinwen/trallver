package service

import (
    "log"
    "my_project/server/internal/model"
    "github.com/streadway/amqp"
    "time"
)

// 处理探测任务，分别处理固定字段和动态字段
func ProcessProbeTask(ch *amqp.Channel, fixed model.ProbeTaskFixed, dynamic model.ProbeTaskDynamic) error {

    // 更新动态字段的状态和分发时间
    dynamic.Status = model.StatusInProgress
    dynamic.DispatchTime = uint32(time.Now().Unix())  // 转换为 uint32

    log.Printf("Assigned probe task to client: %+v", dynamic)

    var binaryData []byte
    var err error

    // 判断是否是首次任务
    if !firstTaskSent {
        // 首次任务发送固定字段和动态字段
        log.Println("First task, sending fixed and dynamic fields.")

        taskWithFixedFields := struct {
            DynamicTask model.ProbeTaskDynamic
            FixedFields model.ProbeTaskFixed
        }{
            DynamicTask: dynamic,
            FixedFields: fixed,
        }

        // 将固定字段和动态字段序列化
        binaryData, err = ToBinaryTask(taskWithFixedFields)
        if err != nil {
            log.Printf("Failed to serialize task with fixed fields: %v", err)
            dynamic.Status = model.StatusFailed
            return err
        }

        // 标记首次任务已发送
        firstTaskSent = true

    } else {
        // 后续任务只发送动态字段
        log.Println("Subsequent task, sending dynamic fields only.")

        // 将动态字段序列化
        binaryData, err = ToBinaryTask(dynamic)
        if err != nil {
            log.Printf("Failed to serialize dynamic task: %v", err)
            dynamic.Status = model.StatusFailed
            return err
        }
    }

    // 压缩二进制数据
    compressedData, err := Compress(binaryData)
    if err != nil {
        log.Printf("Failed to compress probe task: %v", err)
        dynamic.Status = model.StatusFailed
        return err
    }

    // 声明 fanout 交换机，确保交换机存在
    err = ch.ExchangeDeclare(
        "task_exchange",  // 交换机名称
        "fanout",         // 交换机类型
        true,             // durable
        false,            // auto-deleted
        false,            // internal
        false,            // no-wait
        nil,              // arguments
    )
    if err != nil {
        log.Printf("Failed to declare exchange: %v", err)
        dynamic.Status = model.StatusFailed
        return err
    }

    // 发布消息到 fanout 交换机
    err = ch.Publish(
        "task_exchange",  // 交换机名称
        "",               // 对于 fanout 交换机，routing key 可以留空
        false,
        false,
        amqp.Publishing{
            ContentType: "application/octet-stream", // 使用二进制内容类型
            Body:        compressedData,             // 使用压缩后的数据
        })
    if err != nil {
        log.Printf("Failed to publish task to exchange: %v", err)
        dynamic.Status = model.StatusFailed
        return err
    }

    dynamic.Status = model.StatusCompleted // 如果发布成功，更新状态为 Completed
    log.Println("Task published successfully to RabbitMQ")

    return nil
}

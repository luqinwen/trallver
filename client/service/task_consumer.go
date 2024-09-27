package service

import (
    "encoding/json"
    "log"
    "my_project/client/model"
    "sync"
    "github.com/spf13/viper"
    "github.com/streadway/amqp"
)

var (
    fixedFieldsCache model.ProbeTaskFixed // 用于缓存固定字段
    once sync.Once                        // 确保固定字段缓存只执行一次
)

// 缓存固定字段，只在首次接收到任务时进行缓存
func CacheFixedFields(fixedFields model.ProbeTaskFixed) {
    once.Do(func() {
        fixedFieldsCache = fixedFields
        log.Println("Fixed fields cached:", fixedFieldsCache)
    })
}

func StartConsuming(ch *amqp.Channel) {
    log.Println("Starting to consume messages from RabbitMQ...")

    queueName := viper.GetString("rabbitmq.task_queue")
    _, err := ch.QueueDeclare(
        queueName,
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to declare task queue: %v", err)
        return
    }

    err = ch.QueueBind(
        queueName,
        "",
        viper.GetString("rabbitmq.task_exchange"),
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to bind task queue to exchange: %v", err)
        return
    }

    msgs, err := ch.Consume(
        queueName,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to register a consumer: %v", err)
        return
    }

    go func() {
        log.Println("Consumer is now listening for messages...")
        for msg := range msgs {
            log.Printf("Received a message from RabbitMQ: %s", msg.Body)

            // 尝试接收包含固定字段和动态字段的组合
            var taskWithFixedFields struct {
                DynamicTask model.ProbeTaskDynamic
                FixedTask   model.ProbeTaskFixed
            }
            err := json.Unmarshal(msg.Body, &taskWithFixedFields)
            if err == nil {
                // 如果解析成功，说明是首次任务
                CacheFixedFields(taskWithFixedFields.FixedTask) // 缓存固定字段
                ExecuteProbeTask(&taskWithFixedFields.DynamicTask) // 执行任务
                continue
            }

            // 如果不是首次任务，尝试仅接收动态字段
            var dynamicTask model.ProbeTaskDynamic
            err = json.Unmarshal(msg.Body, &dynamicTask)
            if err != nil {
                log.Printf("Failed to unmarshal task: %v", err)
                continue
            }

            log.Printf("Received dynamic task from server: %+v", dynamicTask)
            ExecuteProbeTask(&dynamicTask) // 执行任务
        }

        log.Println("Consumer loop has exited, no more messages are being processed.")
    }()
}

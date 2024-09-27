package service

import (
    "encoding/json"
    "log"
    "my_project/client/model"
    "github.com/spf13/viper"
    "github.com/streadway/amqp"
)

func WaitForConfirmation(ch *amqp.Channel, taskID uint32) bool {
    confirmationQueue := viper.GetString("rabbitmq.confirmation_queue")

    // 注册消费者，监听 confirmation_queue
    msgs, err := ch.Consume(
        confirmationQueue, // 确认消息队列
        "",                // 消费者标签
        true,              // 自动确认
        false,             // 独占
        false,             // 不等待
        false,             // 本地消息
        nil,
    )
    if err != nil {
        log.Printf("Failed to consume from confirmation queue: %v", err)
        return false
    }

    log.Printf("Waiting for confirmation messages in queue: %s", confirmationQueue)

    for msg := range msgs {
        var confirmation model.Confirmation
        err := json.Unmarshal(msg.Body, &confirmation)
        if err != nil {
            log.Printf("Failed to unmarshal confirmation: %v", err)
            continue
        }

        // 匹配任务ID
        if confirmation.TaskID == taskID && confirmation.Status == model.StatusCompleted {
            log.Printf("Confirmation received for task ID: %d with status: %d", taskID, confirmation.Status)
            return true
        }
    }

    return false
}

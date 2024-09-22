package service

import (
    "encoding/json"
    "log"
    "my_project/client/config"
    "my_project/client/model"
    "github.com/spf13/viper"
    "github.com/streadway/amqp"
)

func ReportResultsToMQ(ch *amqp.Channel, result *model.ProbeResult) {
    log.Printf("Attempting to publish probe results: %+v", result)

    body, err := json.Marshal(result)
    if err != nil {
        log.Printf("Failed to encode report data to JSON: %v", err)
        return
    }

    for attempts := 0; attempts < 3; attempts++ {
        err = ch.Publish(
            viper.GetString("rabbitmq.result_exchange"),
            viper.GetString("rabbitmq.result_routing_key"),
            false,
            false,
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            })

        if err == nil {
            log.Println("Probe results reported successfully to server via RabbitMQ")

            // 等待服务器确认消息
            if WaitForConfirmation(ch, result.TaskID) {
                log.Println("Result confirmed by server")
                return
            } else {
                log.Println("Failed to receive confirmation from server")
            }
        }

        log.Printf("Failed to publish result to queue: %v", err)
        if err == amqp.ErrClosed {
            log.Println("Channel is closed. Attempting to reconnect...")

            ch.Close()

            var conn *amqp.Connection
            conn, ch, err = config.InitRabbitMQ()
            if err != nil {
                log.Printf("Failed to reconnect to RabbitMQ: %v", err)
                return
            }
            defer conn.Close()
        }
    }

    log.Println("Failed to report probe results after multiple attempts")
}

func WaitForConfirmation(ch *amqp.Channel, taskID uint32) bool {
    // 读取确认交换机和队列的配置信息
    confirmationQueue := viper.GetString("rabbitmq.confirmation_queue")
    confirmationExchange := viper.GetString("rabbitmq.confirmation_exchange")
    confirmationRoutingKey := viper.GetString("rabbitmq.confirmation_routing_key")

    // 确保绑定确认队列到确认交换机
    err := ch.QueueBind(
        confirmationQueue,
        confirmationRoutingKey,
        confirmationExchange,
        false,
        nil,
    )
    if err != nil {
        log.Printf("Failed to bind confirmation queue: %v", err)
        return false
    }

    log.Printf("Waiting for confirmation messages in queue: %s", confirmationQueue)

    msgs, err := ch.Consume(
        confirmationQueue,
        "",
        true,  // 自动确认
        false, // 非独占
        false, // 不阻塞
        false, // 本地队列
        nil,
    )
    if err != nil {
        log.Printf("Failed to register a consumer for confirmations on queue %s: %v", confirmationQueue, err)
        return false
    }

    // 等待确认消息
    for msg := range msgs {
        var confirmation model.Confirmation
        err := json.Unmarshal(msg.Body, &confirmation)
        if err != nil {
            log.Printf("Failed to unmarshal confirmation: %v", err)
            continue
        }

        // 检查确认消息是否对应当前的任务
        if confirmation.TaskID == taskID && confirmation.Status == model.StatusCompleted {
            log.Printf("Received confirmation for task ID: %d", taskID)
            return true
        }
    }

    return false
}

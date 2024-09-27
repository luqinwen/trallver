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
    // PacketLoss 和 Threshold 已经被打包到 Packed 中，不再直接使用 PacketLoss 和 Threshold 字段
    // result.Packed 不需要重新打包

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

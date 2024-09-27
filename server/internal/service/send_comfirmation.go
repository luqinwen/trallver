package service

import (
	"encoding/json"
	"fmt"
	"log"
	"my_project/server/internal/model"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)


func SendConfirmation(ch *amqp.Channel, taskID uint32, status model.TaskStatus, routingKey string) error {
    confirmation := model.Confirmation{
        TaskID: taskID,
        Status: status,
    }

    body, err := json.Marshal(confirmation)
    if err != nil {
        return fmt.Errorf("failed to marshal confirmation: %w", err)
    }

    // 发送确认消息到相应客户端的 confirmation_queue
    err = ch.Publish(
        viper.GetString("rabbitmq.confirmation_exchange"), // 确认消息的交换机
        routingKey,                                        // 根据客户端 routing key 发送消息
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        return fmt.Errorf("failed to publish confirmation: %w", err)
    }

    log.Printf("Confirmation sent for task ID: %d with status: %d using routing key: %s", taskID, status, routingKey)
    return nil
}

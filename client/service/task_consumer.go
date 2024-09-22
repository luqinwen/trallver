package service

import (
    "encoding/json"
    "log"
    "my_project/client/model"
    "github.com/spf13/viper"
    "github.com/streadway/amqp"
)

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
            var task model.ProbeTask
            err := json.Unmarshal(msg.Body, &task)
            if err != nil {
                log.Printf("Failed to unmarshal task: %v", err)
                continue
            }

            task.Status = model.StatusInProgress
            log.Printf("Received probe task from server: %+v", task)

            result := ExecuteProbeTask(&task)
            if result == nil {
                log.Printf("Task execution failed for task: %+v", task)
                task.Status = model.StatusFailed
                continue
            }

            task.Status = model.StatusCompleted
            log.Printf("Task executed successfully, reporting result: %+v", result)
            ReportResultsToMQ(ch, result)
        }

        log.Println("Consumer loop has exited, no more messages are being processed.")
    }()
}

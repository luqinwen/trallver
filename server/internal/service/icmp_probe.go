package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"my_project/server/internal/dao"
	"my_project/server/internal/model"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func HandleProbeTask(ch *amqp.Channel) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        task := model.ProbeTask{
            IP:        "8.8.8.8", // 示例IP，实际可从请求或数据库中获取
            Count:     4,
            Threshold: 10, // 丢包率阈值
            Timeout:   5,  // 超时时间
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        err := dao.StoreProbeTask(&task)
        if err != nil {
            log.Printf("Failed to insert task into MySQL: %v", err)
            c.String(500, "Failed to insert task into MySQL")
            return
        }
        log.Printf("Assigned probe task to client: %+v", task)

        body, err := json.Marshal(task)
        if err != nil {
            log.Printf("Failed to marshal probe task: %v", err)
            c.String(500, "Failed to marshal probe task")
            return
        }
        
        // 绑定队列到交换机，确保消息能够正确路由
        err = ch.QueueBind(
            viper.GetString("rabbitmq.task_queue"),
            viper.GetString("rabbitmq.task_routing_key"),
            viper.GetString("rabbitmq.exchange"),
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to bind queue to exchange: %v", err)
            c.String(500, "Failed to bind queue to exchange")
            return
        }

        log.Printf("Publishing task to RabbitMQ: exchange=%s, routing_key=%s, body=%s", viper.GetString("rabbitmq.exchange"), viper.GetString("rabbitmq.routing_key"), string(body))


        err = ch.Publish(
            viper.GetString("rabbitmq.exchange"),
            viper.GetString("rabbitmq.task_routing_key"),
            false,
            false,
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            })
        if err != nil {
            log.Printf("Failed to publish task to queue: %v", err)
            c.String(500, "Failed to publish task to queue")
            return
        }

        log.Println("Task published successfully to RabbitMQ")
        c.String(200, "Task assigned successfully")
    }
}


func HandleProbeResults(ch *amqp.Channel) {
    log.Printf("Registering consumer for queue: %s", viper.GetString("rabbitmq.result_queue"))

    // 绑定队列到交换机
    err := ch.QueueBind(
        viper.GetString("rabbitmq.result_queue"),
        viper.GetString("rabbitmq.result_routing_key"),
        viper.GetString("rabbitmq.exchange"),
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to bind queue to exchange: %v", err)
        return
    }

    msgs, err := ch.Consume(
        viper.GetString("rabbitmq.result_queue"),
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

    log.Println("Consumer registered successfully")

    for msg := range msgs {
        log.Printf("Received message from RabbitMQ: %s", string(msg.Body))

        var result model.ProbeResult
        err := json.Unmarshal(msg.Body, &result)
        if err != nil {
            log.Printf("Failed to unmarshal probe result: %v", err)
            continue
        }

        log.Printf("Processed probe result: %+v", result)

        timestamp := result.Timestamp.Unix()
        err = dao.StoreClickHouse(timestamp, result.IP, result.PacketLoss, float64(result.MinRTT.Microseconds()), float64(result.MaxRTT.Microseconds()), float64(result.AvgRTT.Microseconds()))
        if err != nil {
            log.Printf("Failed to store probe result to ClickHouse: %v", err)
            continue
        }

        if result.PacketLoss > float64(result.Threshold) {
            log.Printf("Packet loss exceeds threshold, reporting to Prometheus: %f > %f", result.PacketLoss, result.Threshold)
            ReportToPrometheus(&result, timestamp)
        }
    }
}


// ReportToPrometheus 上报探测结果到 Prometheus
func ReportToPrometheus(result *model.ProbeResult, timestamp int64) {
    prometheusHost := viper.GetString("prometheus.host")
    prometheusPort := viper.GetInt("prometheus.port")
    prometheusJob := viper.GetString("prometheus.job")

    uri := fmt.Sprintf("http://%s:%d/metrics/job/%s", prometheusHost, prometheusPort, prometheusJob)
    
    metrics := fmt.Sprintf("packet_loss{ip=\"%s\", timestamp=\"%d\"} %f\n", result.IP, timestamp, result.PacketLoss)
    log.Printf("Sending data to Prometheus: %s", metrics)

    hertzClient, err := client.NewClient()  // 确保正确实例化 client
    if err != nil {
        log.Fatalf("Failed to create Hertz client: %v", err)
    }

    req := &protocol.Request{}
    req.SetMethod("POST")
    req.SetRequestURI(uri)
    req.Header.Set("Content-Type", "text/plain")
    req.SetBodyString(metrics)

    resp := &protocol.Response{}
    err = hertzClient.Do(context.Background(), req, resp)
    if err != nil {
        log.Printf("Error sending data to Prometheus: %v", err)
    } else {
        log.Printf("Successfully sent data to Prometheus: %s, Response: %s", metrics, resp.Body())
    }
}

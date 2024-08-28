package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"my_project/server/internal/dao"
	"my_project/server/internal/model"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)


func CreateProbeTask() *model.ProbeTask {
    return &model.ProbeTask{
        IP:        "8.8.8.8", // 示例IP，实际可从请求或数据库中获取
        Count:     4,
        Threshold: 10, // 丢包率阈值
        Timeout:   5,  // 超时时间
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func ProcessProbeTask(ch *amqp.Channel, task *model.ProbeTask) error {
    
    task.DispatchTime = time.Now()

    err := dao.StoreProbeTask(task)
    if err != nil {
        log.Printf("Failed to insert task into MySQL: %v", err)
        return err
    }
    log.Printf("Assigned probe task to client: %+v", task)

    body, err := json.Marshal(task)
    if err != nil {
        log.Printf("Failed to marshal probe task: %v", err)
        return err
    }

    // 声明fanout交换机，确保交换机存在
    err = ch.ExchangeDeclare(
        viper.GetString("rabbitmq.task_exchange"),
        "fanout",  // 交换机类型
        true,      // durable
        false,     // auto-deleted
        false,     // internal
        false,     // no-wait
        nil,       // arguments
    )
    if err != nil {
        log.Printf("Failed to declare exchange: %v", err)
        return err
    }

    // 发布消息到fanout交换机，不需要指定routing key
    err = ch.Publish(
        viper.GetString("rabbitmq.task_exchange"),
        "",  // 对于fanout交换机，routing key可以留空
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        log.Printf("Failed to publish task to exchange: %v", err)
        return err
    }

    log.Println("Task published successfully to RabbitMQ")
    return nil
}

func HandleProbeTask(ch *amqp.Channel) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        task := CreateProbeTask()

        err := ProcessProbeTask(ch, task)
        if err != nil {
            c.String(500, "Failed to assign task")
        } else {
            c.String(200, "Task assigned successfully")
        }
    }
}

var (
	roundCounter   int
	roundMutex     sync.Mutex
	resultsStorage map[int][]model.ProbeResult
	storageMutex   sync.Mutex
)

func init() {
	roundCounter = 1
	resultsStorage = make(map[int][]model.ProbeResult)
}

func consumeResultQueue(ch *amqp.Channel, queueName string, queueID int, resultsChan chan model.ProbeResult) {
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
        log.Fatalf("Failed to register a consumer for queue %s: %v", queueName, err)
    }

    log.Printf("Started consumer for queue: %s", queueName)

    for msg := range msgs {
        log.Printf("Received message from %s: %s", queueName, string(msg.Body))
    
        var result model.ProbeResult
        err := json.Unmarshal(msg.Body, &result)
        if err != nil {
            log.Printf("Failed to unmarshal probe result: %v", err)
            continue
        }

        log.Printf("Processed probe result from %s: %+v", queueName, result)

        // 计算总时延
        receiveTime := time.Now()
        totalLatency := receiveTime.Sub(result.DispatchTime).Milliseconds()

        // 存储独立结果到 queue_results 表
        timestamp := result.Timestamp.Unix()
        err = dao.StoreQueueResults(timestamp, queueID, result.IP, result.PacketLoss, 
            float64(result.MinRTT.Microseconds()), float64(result.MaxRTT.Microseconds()), 
            float64(result.AvgRTT.Microseconds()), uint32(totalLatency))
        if err != nil {
            log.Printf("Failed to store queue result to ClickHouse: %v", err)
            continue
        }

        // 将结果存储到当前回合的累积器
        storageMutex.Lock()
        resultsStorage[roundCounter] = append(resultsStorage[roundCounter], result)
        storageMutex.Unlock()

        // 检查当前回合是否已经收集完所有队列的结果
        storageMutex.Lock()
        if len(resultsStorage[roundCounter]) == 3 {
            // 聚合处理并存储结果
            aggregateAndStoreResults(resultsStorage[roundCounter])

            // 清空累积器并开始新一回合
            delete(resultsStorage, roundCounter)
            roundCounter++
        }
        storageMutex.Unlock()
    }
}

func aggregateAndStoreResults(results []model.ProbeResult) {
    var totalPacketLoss float64
    var totalLatencyMs uint32
    var messageCount int

    for _, result := range results {
        totalPacketLoss += result.PacketLoss
        receiveTime := time.Now()
        totalLatency := receiveTime.Sub(result.DispatchTime).Milliseconds()
        totalLatencyMs += uint32(totalLatency)
        messageCount++
    }

    if messageCount > 0 {
        avgPacketLoss := totalPacketLoss / float64(messageCount)
        avgLatencyMs := totalLatencyMs / uint32(messageCount)

        log.Printf("Aggregated Average Packet Loss: %f, Average Latency: %d ms", avgPacketLoss, avgLatencyMs)

        // 存储聚合后的结果
        timestamp := time.Now().Unix()
        err := dao.StoreAggregatedResults(timestamp, avgPacketLoss, avgLatencyMs)
        if err != nil {
            log.Printf("Failed to store aggregated probe result to ClickHouse: %v", err)
        }
    }
}

func StartConsumingResults(ch *amqp.Channel) {
    resultQueues := []string{"result_queue1", "result_queue2", "result_queue3"}
    resultsChan := make(chan model.ProbeResult, len(resultQueues))

    for i, queueName := range resultQueues {
        go consumeResultQueue(ch, queueName, i+1, resultsChan)
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

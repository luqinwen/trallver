package service

import (
	"encoding/json"
	"log"
	"my_project/server/internal/dao"
	"my_project/server/internal/model"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

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
        // 解压缩消息体
        decompressedData, err := Decompress(msg.Body)
        if err != nil {
            log.Printf("Failed to decompress probe result: %v", err)
            continue
        }

        var result model.ProbeResult
        err = json.Unmarshal(decompressedData, &result)
        if err != nil {
            log.Printf("Failed to unmarshal probe result: %v", err)
            continue
        }

        log.Printf("Processed probe result from %s: %+v", queueName, result)

        // 计算总时延
        receiveTime := time.Now()
        dispatchTime := time.Unix(result.DispatchTime, 0)
        totalLatency := receiveTime.Sub(dispatchTime).Milliseconds()

        // 存储独立结果到 queue_results 表
        timestamp := result.Timestamp
        taskID := result.TaskID  // 从 result 中获取 TaskID

        err = dao.StoreQueueResults(timestamp, taskID, queueID, result.IP, result.PacketLoss, 
            result.MinRTT, result.MaxRTT, result.AvgRTT, uint32(totalLatency))

        if err != nil {
            log.Printf("Failed to store queue result to ClickHouse: %v", err)
            continue
        }

        // 确定发送到客户端的 routing key
        var routingKey string
        switch queueName {
        case viper.GetString("rabbitmq.result_queue1"):
            routingKey = viper.GetString("rabbitmq.confirmation_routing_key1")
        case viper.GetString("rabbitmq.result_queue2"):
            routingKey = viper.GetString("rabbitmq.confirmation_routing_key2")
        case viper.GetString("rabbitmq.result_queue3"):
            routingKey = viper.GetString("rabbitmq.confirmation_routing_key3")
        default:
            log.Printf("Unknown result queue: %s", queueName)
            continue
        }

        // 发送确认消息
        err = SendConfirmation(ch, result.TaskID, model.StatusCompleted, routingKey)
        if err != nil {
            log.Printf("Failed to send confirmation for task ID: %d, error: %v", result.TaskID, err)
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
        totalPacketLoss += float64(result.PacketLoss)
        receiveTime := time.Now()
        dispatchTime := time.Unix(result.DispatchTime, 0)
        totalLatency := receiveTime.Sub(dispatchTime).Milliseconds()
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
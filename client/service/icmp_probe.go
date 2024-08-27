package service

import (
	"encoding/json"
	"log"
	"my_project/client/model"
    "my_project/client/config"
	"time"

	"github.com/prometheus-community/pro-bing"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// StartConsuming 启动消息消费
func StartConsuming(ch *amqp.Channel) {
    log.Println("Starting to consume messages from RabbitMQ...")

    msgs, err := ch.Consume(
        viper.GetString("rabbitmq.task_queue"),
        "",
        true,  // 自动确认
        false, // 非独占
        false, // 不阻塞
        false, // 本地
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to register a consumer: %v", err)
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
            log.Printf("Received probe task from server: %+v", task)

            result := ExecuteProbeTask(&task)
            if result == nil {
                log.Printf("Task execution failed for task: %+v", task)
                continue
            }

            log.Printf("Task executed successfully, reporting result: %+v", result)
            ReportResultsToMQ(ch, result)
        }

        log.Println("Consumer loop has exited, no more messages are being processed.")
    }()
}


// ExecuteProbeTask 执行探测任务
func ExecuteProbeTask(task *model.ProbeTask) *model.ProbeResult {

    startTime := time.Now() // 记录任务开始处理时间

    pinger, err := probing.NewPinger(task.IP)
    if err != nil {
        log.Printf("Failed to create pinger: %v", err)
        return nil
    }

    pinger.Count = task.Count
    pinger.Timeout = time.Duration(task.Timeout) * time.Second

    var packetLoss float64
    var minRTT, maxRTT, avgRTT time.Duration

    pinger.OnRecv = func(pkt *probing.Packet) {
        log.Printf("Received packet from %s: time=%v", pkt.IPAddr, pkt.Rtt)
    }

    pinger.OnFinish = func(stats *probing.Statistics) {
        log.Printf("Probe finished. Packet loss: %v%%, Min RTT: %v, Max RTT: %v, Avg RTT: %v",
            stats.PacketLoss, stats.MinRtt, stats.MaxRtt, stats.AvgRtt)

        packetLoss = stats.PacketLoss
        minRTT = stats.MinRtt
        maxRTT = stats.MaxRtt
        avgRTT = stats.AvgRtt
    }

    log.Printf("Starting probe to %s", task.IP)
    pinger.Run()

    endTime := time.Now() // 记录任务处理结束时间
    processingLatency := endTime.Sub(startTime).Milliseconds() // 计算处理时延
    log.Printf("Processing latency: %v ms", processingLatency)


    // 创建 ProbeResult 结构体并返回
    result := &model.ProbeResult{
        IP:         task.IP,
        Timestamp:  time.Now(),
        PacketLoss: packetLoss,
        MinRTT:     minRTT,
        MaxRTT:     maxRTT,
        AvgRTT:     avgRTT,
        Threshold:  task.Threshold,
        Success:    packetLoss <= float64(task.Threshold),
        DispatchTime: task.DispatchTime,
    }

    return result
}


func ReportResultsToMQ(ch *amqp.Channel, result *model.ProbeResult) {
    log.Printf("Attempting to publish probe results: %+v", result)

    // 绑定队列到交换机，确保队列绑定到正确的路由键
    err := ch.QueueBind(
        viper.GetString("rabbitmq.result_queue"),       // 队列名称
        viper.GetString("rabbitmq.result_routing_key"), // 路由键
        viper.GetString("rabbitmq.exchange"),           // 交换机名称
        false,                                          // 是否阻塞
        nil,                                            // 其他属性
    )
    if err != nil {
        log.Printf("Failed to bind queue to exchange: %v", err)
        return
    }

    body, err := json.Marshal(result)
    if err != nil {
        log.Printf("Failed to encode report data to JSON: %v", err)
        return
    }

    err = ch.Publish(
        viper.GetString("rabbitmq.exchange"),
        viper.GetString("rabbitmq.result_routing_key"),  
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        log.Printf("Failed to publish result to queue: %v", err)
        if err == amqp.ErrClosed {
            log.Println("Channel is closed. Attempting to reconnect...")
            ch, err = config.InitRabbitMQ()
            if err != nil {
                log.Printf("Failed to reconnect to RabbitMQ: %v", err)
                return
            }
            ReportResultsToMQ(ch, result)
        }
    } else {
        log.Println("Probe results reported successfully to server via RabbitMQ")
    }
}


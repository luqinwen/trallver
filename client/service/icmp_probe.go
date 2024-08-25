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

// ExecuteProbeTask 执行探测任务
func ExecuteProbeTask(task *model.ProbeTask) *model.ProbeResult {
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
    }

    return result
}

func ReceiveProbeTaskFromMQ(ch *amqp.Channel) *model.ProbeTask {
    if ch == nil {
        log.Println("RabbitMQ channel is nil. Attempting to reconnect...")
        var err error
        ch, err = config.InitRabbitMQ()
        if err != nil {
            log.Printf("Failed to reconnect to RabbitMQ: %v", err)
            return nil
        }
    }
    
    log.Println("Waiting for tasks from RabbitMQ...")

    err := ch.QueueBind(
        viper.GetString("rabbitmq.task_queue"),      // 队列名称
        viper.GetString("rabbitmq.task_routing_key"), // 路由键
        viper.GetString("rabbitmq.exchange"),        // 交换机名称
        false,                                       // 是否阻塞
        nil,                                         // 其他属性
    )
    if err != nil {
        log.Printf("Failed to bind queue to exchange: %v", err)
        return nil
    }

    msgs, err := ch.Consume(
        viper.GetString("rabbitmq.task_queue"),
        "",
        true,
        false,
        false,
        false,
        nil,
    )

    if err != nil {
        log.Printf("Failed to register a consumer: %v", err)
        if err == amqp.ErrClosed {
            log.Println("Channel is closed. Attempting to reconnect...")
            ch, err = config.InitRabbitMQ()
            if err != nil {
                log.Printf("Failed to reconnect to RabbitMQ: %v", err)
                return nil
            }
            return ReceiveProbeTaskFromMQ(ch)
        }
        return nil
    }

    for msg := range msgs {
        log.Printf("Received a message from RabbitMQ: %s", msg.Body)
        var task model.ProbeTask
        err := json.Unmarshal(msg.Body, &task)
        if err != nil {
            log.Printf("Failed to unmarshal task: %v", err)
            continue
        }
        log.Printf("Received probe task from server: %+v", task)
        return &task
    }

    log.Println("No more messages in the queue.")
    return nil
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


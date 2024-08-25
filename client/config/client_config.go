package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// Config 结构体保存配置文件中解析的值
type Config struct {
    Probe struct {
        IP        string
        Count     int
        Threshold int
        Timeout   int
    }
    LogFile string
}

// 全局变量，用于存储加载的配置
var ClientConfig Config

// InitConfig 初始化配置文件
func InitConfig() {
    viper.SetConfigName("client_config") // 配置文件名称 (不包含扩展名)
    viper.AddConfigPath("/root/config")      // 配置文件所在的路径
    viper.SetConfigType("yaml")          // 设置配置文件类型

    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    // 解析配置文件中的值到 Config 结构体中
    err = viper.Unmarshal(&ClientConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }

    log.Println("Config file loaded successfully")
}

func InitRabbitMQ() (*amqp.Channel, error) {
    var conn *amqp.Connection
    var ch *amqp.Channel
    var err error

    rabbitmqURL := viper.GetString("rabbitmq.url")
    log.Printf("Initializing RabbitMQ connection with URL: %s", rabbitmqURL)

    for i := 0; i < 10; i++ {  // 尝试多次连接
        log.Printf("Attempt %d: Connecting to RabbitMQ...", i+1)

        conn, err = amqp.Dial(rabbitmqURL)
        if err != nil {
            log.Printf("Failed to connect to RabbitMQ: %v, retrying in 5 seconds...", err)
            time.Sleep(5 * time.Second)  // 等待5秒后重试
            continue
        }

        ch, err = conn.Channel()
        if err != nil {
            log.Printf("Failed to open a channel: %v, retrying in 5 seconds...", err)
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Successfully connected to RabbitMQ and created channel")


        // 声明交换机
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.exchange"),
            "direct",  // 交换机类型
            true,      // 持久化
            false,     // 自动删除
            false,     // 内部使用
            false,     // 是否阻塞
            nil,       // 额外属性
        )
        if err != nil {
            log.Printf("Failed to declare exchange: %v", err)
            ch.Close()
            continue
        }

        // 声明队列，确保在初始化后队列存在
        _, err = ch.QueueDeclare(
            viper.GetString("rabbitmq.task_queue"),
            true,
            false,
            false,
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to declare task_queue: %v", err)
            ch.Close()
            continue
        }

        _, err = ch.QueueDeclare(
            viper.GetString("rabbitmq.result_queue"),
            true,
            false,
            false,
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to declare result_queue: %v", err)
            ch.Close()
            continue
        }

        return ch, nil
    }

    return nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts: %v", err)
}

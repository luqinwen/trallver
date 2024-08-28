package config

import (
	"fmt"
	"log"
	"os"
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
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        log.Fatalf("CONFIG_PATH environment variable is not set")
    }

    viper.SetConfigFile(configPath)  // 使用环境变量指定的配置文件路径

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


func InitRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
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
            time.Sleep(5 * time.Second)
            continue
        }

        ch, err = conn.Channel()
        if err != nil {
            log.Printf("Failed to open a channel: %v, retrying in 5 seconds...", err)
            conn.Close()  // 关闭连接
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Successfully connected to RabbitMQ and created channel")

        // 声明fanout交换机用于任务接收
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.task_exchange"),
            "fanout",  // 交换机类型
            true,      // 持久化
            false,     // 自动删除
            false,     // 内部使用
            false,     // 是否阻塞
            nil,       // 额外属性
        )
        if err != nil {
            log.Printf("Failed to declare task exchange: %v", err)
            ch.Close()
            conn.Close()  // 关闭连接
            continue
        }

        // 声明direct交换机用于结果上报
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.result_exchange"),
            "direct",  // 交换机类型
            true,      // 持久化
            false,     // 自动删除
            false,     // 内部使用
            false,     // 是否阻塞
            nil,       // 额外属性
        )
        if err != nil {
            log.Printf("Failed to declare result exchange: %v", err)
            ch.Close()
            conn.Close()  // 关闭连接
            continue
        }

        return conn, ch, nil  // 返回连接和通道
    }

    return nil, nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts: %v", err)
}
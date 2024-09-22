package config

import (
	"fmt"
	"log"
	"my_project/server/internal/common" // 确保导入了 common 包
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// Config 结构体保存配置文件中解析的值
type Config struct {
    MySQL struct {
        User     string
        Password string
        Host     string
        Port     int
        Database string
    }
    ClickHouse struct {
        Host string
        Port int
    }
    LogFile string
}

// 全局变量，用于存储加载的配置
var ServerConfig Config

// InitConfig 初始化配置文件
func InitConfig() {
    viper.SetConfigName("server_config") // 配置文件名称（不包含扩展名）
    viper.AddConfigPath("/root/config")  // 配置文件所在路径
    viper.SetConfigType("yaml")          // 设置配置文件类型

    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    // 解析配置文件中的值到 Config 结构体中
    err = viper.Unmarshal(&ServerConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }

    log.Println("Config file loaded successfully")
}

func InitRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
    var ch *amqp.Channel
    var conn *amqp.Connection
    var err error

    rabbitmqURL := viper.GetString("rabbitmq.url")
    log.Printf("Initializing RabbitMQ connection with URL: %s", rabbitmqURL)

    for i := 0; i < 10; i++ {  // 重试10次
        log.Printf("Attempt %d: Connecting to RabbitMQ...", i+1)

        conn, err = amqp.Dial(rabbitmqURL)
        if err != nil {
            log.Printf("Failed to connect to RabbitMQ: %v, retrying in 10 seconds...", err)
            time.Sleep(10 * time.Second)  // 每次失败后等待10秒
            continue
        }

        log.Println("Successfully connected to RabbitMQ, creating channel...")

        ch, err = conn.Channel()
        if err != nil {
            log.Printf("Failed to create RabbitMQ channel: %v, retrying in 5 seconds...", err)
            conn.Close()  // 通道创建失败时，关闭连接
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Successfully created RabbitMQ channel")

        // 声明 fanout 交换机，用于任务下发
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.task_exchange"), // 从配置文件获取任务交换机名称
            "fanout",                                 // 交换机类型为 fanout
            true,                                     // 是否持久化
            false,                                    // 是否自动删除
            false,                                    // 是否排他
            false,                                    // 是否阻塞
            nil,                                      // 其他属性
        )

        if err != nil {
            log.Printf("Failed to declare task exchange '%s': %v", viper.GetString("rabbitmq.task_exchange"), err)
            ch.Close()  // 如果交换机声明失败，关闭通道
            time.Sleep(5 * time.Second)
            continue
        }

        // 声明 direct 交换机，用于结果上传
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.result_exchange"), // 从配置文件获取结果交换机名称
            "direct",                                   // 交换机类型为 direct
            true,                                       // 是否持久化
            false,                                      // 是否自动删除
            false,                                      // 是否排他
            false,                                      // 是否阻塞
            nil,                                        // 其他属性
        )

        if err != nil {
            log.Printf("Failed to declare result exchange '%s': %v", viper.GetString("rabbitmq.result_exchange"), err)
            ch.Close()  // 如果交换机声明失败，关闭通道
            time.Sleep(5 * time.Second)
            continue
        }

        // 声明确认消息的 direct 交换机
        err = ch.ExchangeDeclare(
            viper.GetString("rabbitmq.confirmation_exchange"), // 确认交换机名称
            "direct",                                          // 交换机类型为 direct
            true,                                              // 是否持久化
            false,                                             // 是否自动删除
            false,                                             // 是否排他
            false,                                             // 是否阻塞
            nil,                                               // 其他属性
        )
        if err != nil {
            log.Printf("Failed to declare confirmation exchange '%s': %v", viper.GetString("rabbitmq.confirmation_exchange"), err)
            ch.Close()
            time.Sleep(5 * time.Second)
            continue
        }

        // 声明 task_queue 队列，确保在初始化后队列存在
        _, err = ch.QueueDeclare(
            viper.GetString("rabbitmq.task_queue"),
            true,
            false,
            false,
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to declare task_queue: %v, retrying in 5 seconds...", err)
            ch.Close()    // 如果队列声明失败，关闭通道
            conn.Close()  // 并关闭连接
            time.Sleep(5 * time.Second)
            continue
        }

        log.Println("Successfully declared task_queue")

        // 声明 result_queue1 队列，确保在初始化后队列存在
        for i := 1; i <= 3; i++ {
            queueName := fmt.Sprintf("result_queue%d", i)
            _, err = ch.QueueDeclare(
                queueName,
                true,
                false,
                false,
                false,
                nil,
            )
            if err != nil {
                log.Printf("Failed to declare %s: %v, retrying in 5 seconds...", queueName, err)
                ch.Close()    // 如果队列声明失败，关闭通道
                conn.Close()  // 并关闭连接
                time.Sleep(5 * time.Second)
                continue
            }

            // 将队列绑定到结果交换机上
            err = ch.QueueBind(
                queueName,                             // 队列名称
                fmt.Sprintf("result_routing_key%d", i),       // routing key
                viper.GetString("rabbitmq.result_exchange"), // 交换机名称
                false,
                nil,
            )
            if err != nil {
                log.Printf("Failed to bind %s to exchange: %v, retrying in 5 seconds...", queueName, err)
                ch.Close()    // 如果绑定失败，关闭通道
                conn.Close()  // 并关闭连接
                time.Sleep(5 * time.Second)
                continue
            }

            log.Printf("Successfully declared and bound %s", queueName)
        }

        return conn, ch, nil  // 返回连接和通道
    }

        // 声明 confirmation_queue 队列，用于接收服务器发送的确认消息
        _, err = ch.QueueDeclare(
            viper.GetString("rabbitmq.confirmation_queue"),
            true,
            false,
            false,
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to declare confirmation_queue: %v", err)
            ch.Close()
            conn.Close()
            time.Sleep(5 * time.Second)
            return nil, nil, fmt.Errorf("failed to declare confirmation_queue: %v", err)  // 替换 continue 为 return
        }

        // 将 confirmation_queue 绑定到确认交换机
        err = ch.QueueBind(
            viper.GetString("rabbitmq.confirmation_queue"),       // 队列名称
            viper.GetString("rabbitmq.confirmation_routing_key"), // routing key
            viper.GetString("rabbitmq.confirmation_exchange"),    // 交换机名称
            false,
            nil,
        )
        if err != nil {
            log.Printf("Failed to bind confirmation_queue to confirmation exchange: %v", err)
            ch.Close()
            conn.Close()
            time.Sleep(5 * time.Second)
            return nil, nil, fmt.Errorf("failed to bind confirmation_queue: %v", err)  // 替换 continue 为 return
        }

        log.Println("Successfully declared and bound confirmation_queue")

    
    if conn != nil {
        log.Println("Closing RabbitMQ connection due to repeated failures")
        conn.Close()  // 最终失败时关闭连接
    }

    return nil, nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts: %v", err)
}


// Init 初始化配置和日志
func Init() {
    InitConfig()
    InitLog() // 使用从 ServerConfig 结构体中获取的日志文件路径初始化日志
    InitRabbitMQ()
    common.InitClickHouse()
    common.InitMySQL()
}



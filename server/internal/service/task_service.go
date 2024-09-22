package service

import (
	"log"
	"my_project/server/internal/dao"
	"my_project/server/internal/model"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func CreateProbeTask() *model.ProbeTask {
    return &model.ProbeTask{
        IP:          [16]byte{8, 8, 8, 8},  // 这里假设是一个IPv4地址，实际情况下应根据需求填充
        Count:       4,
        Timeout:     5,  // 超时时间
        DispatchTime: time.Now().Unix(),  // Unix时间戳
        Status:       model.StatusPending, // 设置初始状态为Pending
    }
}


func ProcessProbeTask(ch *amqp.Channel, task *model.ProbeTask) error {

    task.Status = model.StatusInProgress // 更新状态为In Progress

    task.DispatchTime = time.Now().Unix()

    err := dao.StoreProbeTask(task)
    if err != nil {
        log.Printf("Failed to insert task into MySQL: %v", err)
        task.Status = model.StatusFailed
        return err
    }
    log.Printf("Assigned probe task to client: %+v", task)


    // 将 ProbeTask 序列化为二进制数据
    binaryData, err := ToBinaryTask(task)

    if err != nil {
        log.Printf("Failed to serialize probe task to binary: %v", err)
        task.Status = model.StatusFailed
        return err
    }

	// 压缩二进制数据
    compressedData, err := Compress(binaryData)
    if err != nil {
        log.Printf("Failed to compress probe task: %v", err)
        task.Status = model.StatusFailed
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
        task.Status = model.StatusFailed
        return err
    }

    // 发布消息到fanout交换机，不需要指定routing key
    err = ch.Publish(
        viper.GetString("rabbitmq.task_exchange"),
        "",  // 对于fanout交换机，routing key可以留空
        false,
        false,
        amqp.Publishing{
            ContentType: "application/octet-stream", // 使用二进制内容类型
            Body:        compressedData,             // 使用压缩后的数据
        })
    if err != nil {
        log.Printf("Failed to publish task to exchange: %v", err)
        task.Status = model.StatusFailed // 如果发布失败，更新状态为Failed
        return err
    }

    task.Status = model.StatusCompleted // 如果发布成功，更新状态为Completed
    log.Println("Task published successfully to RabbitMQ")

    return nil
}
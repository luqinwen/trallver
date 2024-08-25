package main

import (
	"log"
	"my_project/client/config"
	"my_project/client/service"
	"time"
)

func main() {
    config.InitConfig()
    config.InitLog()
    
    ch, err := config.InitRabbitMQ()
    if err != nil {
        log.Fatalf("Failed to initialize RabbitMQ: %v", err)
        return
    }
    defer ch.Close()

    for {
        log.Println("Waiting to receive probe task from RabbitMQ...")
        task := service.ReceiveProbeTaskFromMQ(ch)
        if task == nil {
            log.Println("No task received, retrying in 5 seconds...")
            time.Sleep(5 * time.Second)
            continue
        }

        log.Printf("Received task: %+v", task)
        result := service.ExecuteProbeTask(task)
        if result == nil {
            log.Println("Failed to execute probe task, retrying...")
            continue
        }

        log.Println("Reporting results to RabbitMQ...")
        service.ReportResultsToMQ(ch, result)
    }
}


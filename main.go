package main

import (
    "database/sql"
    "fmt"
    "log"
    "math/rand"
    "time"

    _ "github.com/ClickHouse/clickhouse-go"
    "github.com/go-resty/resty/v2"
)

const threshold = 90 // 阈值设置为90

func main() {
    // 设置随机数种子
    rand.Seed(time.Now().UnixNano())

    // 初始化Resty客户端
    client := resty.New()

    // 连接ClickHouse数据库
    conn, err := sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
    if err != nil {
        log.Fatalf("Error connecting to ClickHouse: %v", err)
    }
    defer conn.Close()

    // 创建表（如果不存在）
    _, err = conn.Exec(`
        CREATE TABLE IF NOT EXISTS my_table (
            timestamp UInt64,
            value     Int32
        ) ENGINE = Log
    `)
    if err != nil {
        log.Fatalf("Error creating table: %v", err)
    }

    for {
        timestamp := uint64(time.Now().Unix())
        randomNumber := rand.Intn(100)

        // 写入Clickhouse
        writeToClickhouse(conn, timestamp, randomNumber)

        if randomNumber > threshold {
            // 发送到Prometheus
            sendToPrometheus(client, timestamp, randomNumber)
        }

        time.Sleep(1 * time.Second)
    }
}

func writeToClickhouse(conn *sql.DB, timestamp uint64, value int) {
    log.Printf("Attempting to insert into ClickHouse: timestamp=%d, value=%d", timestamp, value)
    tx, err := conn.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        return
    }

    stmt, err := tx.Prepare("INSERT INTO my_table (timestamp, value) VALUES (?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        tx.Rollback()
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(timestamp, value)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        tx.Rollback()
        return
    }

    if err := tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
    } else {
        log.Printf("Successfully inserted into ClickHouse: timestamp=%d, value=%d", timestamp, value)
    }
}

func sendToPrometheus(client *resty.Client, timestamp uint64, value int) {
    metrics := fmt.Sprintf("random_value{value=\"%d\", timestamp=\"%d\"} %d\n", value, timestamp, timestamp)
    log.Printf("Sending data to Prometheus: %s", metrics)
    resp, err := client.R().
        SetHeader("Content-Type", "text/plain").
        SetBody(metrics).
        Post("http://192.168.188.130:9091/metrics/job/random")
    if err != nil {
        log.Printf("Error sending data to Prometheus: %v", err)
    } else {
        log.Printf("Successfully sent data to Prometheus: %s, Response: %s", metrics, resp)
    }
}


package service

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "math/rand"
    "time"

    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/protocol"
)

const (
    Threshold = 90 // 阈值设置为90
)

func RunService(conn *sql.DB, hertzClient *client.Client) {
    rand.Seed(time.Now().UnixNano())

    for {
        timestamp := uint64(time.Now().Unix())
        randomNumber := rand.Intn(100)

        // 写入ClickHouse
        writeToClickHouse(conn, timestamp, randomNumber)

        if randomNumber > Threshold {
            sendToPrometheus(hertzClient, timestamp, randomNumber)
        }

        time.Sleep(1 * time.Second)
    }
}

func writeToClickHouse(conn *sql.DB, timestamp uint64, value int) {
    log.Printf("Attempting to insert into ClickHouse: timestamp=%d, value=%d", timestamp, value)

    // 检查表是否存在
    var tableExists bool
    retries := 3
    for retries > 0 {
        err := conn.QueryRow("EXISTS TABLE my_database.my_table").Scan(&tableExists)
        if err == nil && tableExists {
            break
        }
        retries--
        log.Printf("Retrying... attempts left: %d", retries)
        time.Sleep(1 * time.Second)
    }
    if retries == 0 {
        log.Fatalf("Failed to verify table existence after multiple attempts")
        return
    }

    tx, err := conn.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        return
    }

    ip := "127.0.0.1"
    packetLoss := 0.0
    minRtt := 0.0
    maxRtt := 0.0
    avgRtt := 0.0

    stmt, err := tx.Prepare("INSERT INTO my_database.my_table (timestamp, ip, packet_loss, min_rtt, max_rtt, avg_rtt) VALUES (?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        tx.Rollback()
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(time.Unix(int64(timestamp), 0), ip, packetLoss, minRtt, maxRtt, avgRtt)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        tx.Rollback()
        return
    }

    if err := tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
    } else {
        log.Printf("Successfully inserted into ClickHouse: timestamp=%d, ip=%s, packet_loss=%f, min_rtt=%f, max_rtt=%f, avg_rtt=%f", timestamp, ip, packetLoss, minRtt, maxRtt, avgRtt)
    }
}

func sendToPrometheus(hertzClient *client.Client, timestamp uint64, value int) {
    metrics := fmt.Sprintf("random_value{value=\"%d\", timestamp=\"%d\"} %d\n", value, timestamp, timestamp)
    log.Printf("Sending data to Prometheus: %s", metrics)

    req := &protocol.Request{}
    req.SetMethod("POST")
    req.SetRequestURI("http://172.20.10.4:9091/metrics/job/random")
    req.Header.Set("Content-Type", "text/plain")
    req.SetBodyString(metrics)

    resp := &protocol.Response{}
    err := hertzClient.Do(context.Background(), req, resp)
    if err != nil {
        log.Printf("Error sending data to Prometheus: %v", err)
    } else {
        log.Printf("Successfully sent data to Prometheus: %s, Response: %s", metrics, resp.Body())
    }
}
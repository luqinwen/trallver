package common

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/ClickHouse/clickhouse-go"
    "github.com/spf13/viper"
)

var ClickHouseDB *sql.DB

func InitClickHouse() {
    dsn := fmt.Sprintf("tcp://%s:%d?debug=true",
        viper.GetString("clickhouse.host"),
        viper.GetInt("clickhouse.port"))

    var err error
    for i := 0; i < 5; i++ { // 重试 5 次
        ClickHouseDB, err = sql.Open("clickhouse", dsn)
        if err != nil {
            log.Printf("Error connecting to ClickHouse: %v", err)
        } else {
            err = ClickHouseDB.Ping()
            if err == nil {
                log.Println("Successfully connected to ClickHouse")
                break
            }
            log.Printf("Failed to ping ClickHouse, attempt (%d/5): %v", i+1, err)
        }
        time.Sleep(2 * time.Second) // 等待 2 秒后重试
    }

    if err != nil {
        log.Fatalf("Failed to connect to ClickHouse after retries: %v", err)
    }

    // 确保数据库存在
    _, err = ClickHouseDB.Exec("CREATE DATABASE IF NOT EXISTS my_database")
    if err != nil {
        log.Fatalf("Error creating database: %v", err)
    }
    log.Println("ClickHouse database created successfully or already exists")

    // 创建 aggregated_results 表
    _, err = ClickHouseDB.Exec(`
        CREATE TABLE IF NOT EXISTS my_database.aggregated_results (
            timestamp DateTime,
            avg_packet_loss Float64,
            avg_latency_ms UInt32
        ) ENGINE = MergeTree()
        ORDER BY timestamp
    `)
    if err != nil {
        log.Fatalf("Error creating table aggregated_results: %v", err)
    }
    log.Println("ClickHouse table aggregated_results created successfully")

    // 创建 queue_results 表
    _, err = ClickHouseDB.Exec(`
        CREATE TABLE IF NOT EXISTS my_database.queue_results (
            timestamp DateTime,
            task_id UInt32,          -- 添加 task_id 字段以关联任务
            queue_id Int32,
            ip UInt32,               -- 使用 UInt32 存储 IPv4 地址
            packet_loss UInt8,       -- 丢包率百分比，使用 UInt8 类型
            min_rtt UInt16,          -- 最小往返时间（毫秒），使用 UInt16 类型
            max_rtt UInt16,          -- 最大往返时间（毫秒），使用 UInt16 类型
            avg_rtt UInt16,          -- 平均往返时间（毫秒），使用 UInt16 类型
            latency_ms UInt32        -- 总时延（毫秒），使用 UInt32 类型
        ) ENGINE = MergeTree()
        ORDER BY timestamp;
    `)
    if err != nil {
        log.Fatalf("Error creating table queue_results: %v", err)
    }

    log.Println("ClickHouse table queue_results created successfully")
}

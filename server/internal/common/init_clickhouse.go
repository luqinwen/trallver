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

    // 创建表
    _, err = ClickHouseDB.Exec(`
        CREATE TABLE IF NOT EXISTS my_database.my_table (
            timestamp DateTime,
            ip String,
            packet_loss Float64,
            min_rtt Float64,
            max_rtt Float64,
            avg_rtt Float64,
            threshold   Int32,
            success     UInt8 
        ) ENGINE = MergeTree()
        ORDER BY timestamp
    `)
    if err != nil {
        log.Fatalf("Error creating table: %v", err)
    }
    log.Println("ClickHouse table created successfully")
}

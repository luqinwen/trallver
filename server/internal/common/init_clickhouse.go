package common

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/ClickHouse/clickhouse-go"
    "github.com/spf13/viper"
)

var ClickHouseDB *sql.DB

func InitClickHouse() {
    dsn := fmt.Sprintf("tcp://%s:%d?debug=true",
        viper.GetString("database.clickhouse.host"),
        viper.GetInt("database.clickhouse.port"))

    var err error
    ClickHouseDB, err = sql.Open("clickhouse", dsn)
    if err != nil {
        log.Fatalf("Error connecting to ClickHouse: %v", err)
    }

    if err = ClickHouseDB.Ping(); err != nil {
        log.Fatalf("Failed to ping ClickHouse: %v", err)
    }
    log.Println("Successfully connected to ClickHouse")

    _, err = ClickHouseDB.Exec(`
        CREATE TABLE IF NOT EXISTS my_database.my_table (
            timestamp DateTime,
            ip String,
            packet_loss Float64,
            min_rtt Float64,
            max_rtt Float64,
            avg_rtt Float64
        ) ENGINE = MergeTree()
        ORDER BY timestamp
    `)
    if err != nil {
        log.Fatalf("Error creating table: %v", err)
    }
    log.Println("ClickHouse table created successfully")
}

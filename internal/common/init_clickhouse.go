package common

import (
    "database/sql"
    "log"

    _ "github.com/ClickHouse/clickhouse-go"
)

var ClickHouseDB *sql.DB

func InitClickHouse() {
    var err error
    ClickHouseDB, err = sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
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

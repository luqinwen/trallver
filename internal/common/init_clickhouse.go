package common

import (
	"database/sql"
	"log"

	_ "github.com/ClickHouse/clickhouse-go"
)

var DB *sql.DB

func InitClickHouse() {
	var err error
	DB, err = sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
	if err != nil {
		log.Fatalf("Error connecting to ClickHouse: %v", err)
	}

	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS my_table (
            timestamp UInt64,
            value     Int32
        ) ENGINE = Log
    `)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

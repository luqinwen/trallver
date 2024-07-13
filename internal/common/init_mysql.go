package common

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "log"
)

var MySQLDB *sql.DB

func InitMySQL() {
    var err error
    MySQLDB, err = sql.Open("mysql", "user:password@tcp(localhost:3306)/my_database")
    if err != nil {
        log.Fatalf("Failed to connect to MySQL: %v", err)
    }
    if err = MySQLDB.Ping(); err != nil {
        log.Fatalf("Failed to ping MySQL: %v", err)
    }
    log.Println("MySQL initialized...")
}

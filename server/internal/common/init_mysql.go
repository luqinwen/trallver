package common

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "github.com/spf13/viper"
)

var MySQLDB *sql.DB

func InitMySQL() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        viper.GetString("database.mysql.user"),
        viper.GetString("database.mysql.password"),
        viper.GetString("database.mysql.host"),
        viper.GetInt("database.mysql.port"),
        viper.GetString("database.mysql.dbname"))

    var err error
    MySQLDB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to MySQL: %v", err)
    }
    if err = MySQLDB.Ping(); err != nil {
        log.Fatalf("Failed to ping MySQL: %v", err)
    }
    log.Println("MySQL initialized...")
}

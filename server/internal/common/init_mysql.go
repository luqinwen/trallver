package common

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/spf13/viper"
)

var MySQLDB *sql.DB

func InitMySQL() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        viper.GetString("mysql.user"),
        viper.GetString("mysql.password"),
        viper.GetString("mysql.host"),
        viper.GetInt("mysql.port"),
        viper.GetString("mysql.database"))

    var err error

    // 增加重试逻辑，最多重试5次，每次间隔2秒
    for i := 0; i < 5; i++ {
        MySQLDB, err = sql.Open("mysql", dsn)
        if err != nil {
            log.Printf("Failed to connect to MySQL (attempt %d): %v", i+1, err)
            time.Sleep(2 * time.Second)
            continue
        }

        err = MySQLDB.Ping()
        if err != nil {
            log.Printf("Failed to ping MySQL (attempt %d): %v", i+1, err)
            time.Sleep(2 * time.Second)
            continue
        }

        log.Println("MySQL initialized successfully")
        break
    }

    if err != nil {
        log.Fatalf("Could not establish connection to MySQL after 5 attempts: %v", err)
    }

    // 创建表结构
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS probe_tasks (
        id INT AUTO_INCREMENT PRIMARY KEY,
        ip VARCHAR(255) NOT NULL,
        count INT NOT NULL,
        port INT DEFAULT 0,
        threshold INT NOT NULL,
        timeout INT NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL
    );`
    
    _, err = MySQLDB.Exec(createTableQuery)
    if err != nil {
        log.Fatalf("Failed to create table probe_tasks: %v", err)
    }

    log.Println("Table probe_tasks ensured.")
}


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
        id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
        ip VARBINARY(16) NOT NULL,     -- 使用 VARBINARY(16) 存储 IPv4 或 IPv6 地址
        count TINYINT NOT NULL,        -- 使用 TINYINT 存储执行次数
        timeout SMALLINT NOT NULL,     -- 使用 SMALLINT 存储超时时间（秒）
        dispatch_time BIGINT NOT NULL, -- 使用 BIGINT 存储 Unix 时间戳（秒）
        status TINYINT NOT NULL        -- 使用 TINYINT 存储任务状态
    );`
    
    _, err = MySQLDB.Exec(createTableQuery)
    if err != nil {
        log.Fatalf("Failed to create table probe_tasks: %v", err)
    }

    log.Println("Table probe_tasks ensured.")
}


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

    // 创建表结构 - 固定字段表
    createFixedTableQuery := `
    CREATE TABLE IF NOT EXISTS probe_task_fixed (
        id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,   -- 自增主键ID
        ip VARBINARY(4) NOT NULL,                     -- 存储 IPv4 地址
        packed INT UNSIGNED NOT NULL                  -- 打包的字段，存储 Timeout, Count, Threshold
    );`

    _, err = MySQLDB.Exec(createFixedTableQuery)
    if err != nil {
        log.Fatalf("Failed to create table probe_task_fixed: %v", err)
    }
    log.Println("Table probe_task_fixed ensured.")

    // 创建表结构 - 动态字段表
    createDynamicTableQuery := `
    CREATE TABLE IF NOT EXISTS probe_tasks (
        id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,   -- 自增任务ID
        fixed_task_id INT UNSIGNED NOT NULL,          -- 引用固定字段表的 ID
        dispatch_time INT UNSIGNED NOT NULL,          -- 任务下发时间（秒级 Unix 时间戳，使用 uint32）
        status TINYINT UNSIGNED NOT NULL,             -- 任务状态
        FOREIGN KEY (fixed_task_id) REFERENCES probe_task_fixed(id) ON DELETE CASCADE
    );`

    _, err = MySQLDB.Exec(createDynamicTableQuery)
    if err != nil {
        log.Fatalf("Failed to create table probe_tasks: %v", err)
    }
    log.Println("Table probe_tasks ensured.")
}

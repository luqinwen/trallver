package config

import (
    "log"
    "os"
)

// InitLog 初始化日志记录
func InitLog(logFile string) {
    // 检查日志目录是否存在，不存在则创建
    logDir := "logs"
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        err := os.Mkdir(logDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create log directory: %v", err)
        }
    }

    // 打开或创建日志文件
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }

    // 设置日志输出到文件
    log.SetOutput(file)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.Println("Log initialized successfully")
}

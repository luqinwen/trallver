package config

import (
    "log"
    "os"
    "github.com/spf13/viper"
)

// InitLog 初始化日志记录
func InitLog() {
    logFile := viper.GetString("log_file") // 从配置文件中读取日志文件路径

    if logFile == "" {
        log.Println("Log file path is empty, logging to stdout")
        return
    }

    logDir := "/root/logs" // 使用绝对路径，确保与 Dockerfile 中的路径一致
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        err := os.MkdirAll(logDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create log directory: %v", err)
        }
    }

    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }

    log.SetOutput(file)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.Println("Log initialized successfully")
}

package config

import (
    "log"
    "os"
)

func InitLog() {
    logDir := "logs"
    logFile := "logs/my_project.log"

    // 创建日志目录
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        err := os.Mkdir(logDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create log directory: %v", err)
        }
    }

    // 打开日志文件
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }

    // 设置日志输出
    log.SetOutput(file)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}


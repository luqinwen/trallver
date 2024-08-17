package config

import (
    "log"
    "github.com/spf13/viper"
    "my_project/server/internal/common" // 确保导入了 common 包
)

// Config 结构体保存配置文件中解析的值
type Config struct {
    MySQL struct {
        User     string
        Password string
        Host     string
        Port     int
        Database string
    }
    ClickHouse struct {
        Host string
        Port int
    }
    LogFile string
}

// 全局变量，用于存储加载的配置
var ServerConfig Config

// InitConfig 初始化配置文件
func InitConfig() {
    viper.SetConfigName("server_config") // 配置文件名称（不包含扩展名）
    viper.AddConfigPath("./config")  // 配置文件所在路径
    viper.SetConfigType("yaml")          // 设置配置文件类型

    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    // 解析配置文件中的值到 Config 结构体中
    err = viper.Unmarshal(&ServerConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }

    log.Println("Config file loaded successfully")
}

// Init 初始化配置和日志
func Init() {
    InitConfig()
    InitLog() // 使用从 ServerConfig 结构体中获取的日志文件路径初始化日志
    common.InitClickHouse()
    common.InitMySQL()
}



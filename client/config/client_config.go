package config

import (
    "log"
    "github.com/spf13/viper"
)

// Config 结构体保存配置文件中解析的值
type Config struct {
    Probe struct {
        IP        string
        Count     int
        Threshold int
        Timeout   int
    }
    LogFile string
}

// 全局变量，用于存储加载的配置
var ClientConfig Config

// InitConfig 初始化配置文件
func InitConfig() {
    viper.SetConfigName("client_config") // 配置文件名称 (不包含扩展名)
    viper.AddConfigPath("./config")      // 配置文件所在的路径
    viper.SetConfigType("yaml")          // 设置配置文件类型

    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    // 解析配置文件中的值到 Config 结构体中
    err = viper.Unmarshal(&ClientConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }

    log.Println("Config file loaded successfully")
}

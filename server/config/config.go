package config

import (
    "log"
    "github.com/spf13/viper"
    "my_project/server/internal/common"
)

// InitConfig 初始化配置文件
func InitConfig() {
    // 设置配置文件的名称和路径
    viper.SetConfigName("server_config") // 配置文件名称（不包含扩展名）
    viper.AddConfigPath("./config")      // 配置文件所在路径
    viper.SetConfigType("yaml")          // 配置文件类型

    // 读取配置文件内容
    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    // 将配置文件解析到结构体或全局变量中
    err = viper.Unmarshal(&common.ServerConfig)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }

    log.Println("Config file loaded successfully")
}

// Init 初始化配置和日志
func Init() {
    InitConfig()
    InitLog(common.ServerConfig.LogFile)
    common.InitClickHouse()
    common.InitMySQL()
}



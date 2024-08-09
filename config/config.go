package config

import (
    "log"
    "my_project/internal/common"
    "github.com/spf13/viper"
)

func InitConfig() {
    // 设置配置文件的名称和路径
    viper.SetConfigName("config") // 配置文件名称 (不包含扩展名)
    viper.AddConfigPath(".")      // 配置文件所在的路径 (当前路径)
    
    // 读取配置文件内容
    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    log.Println("Config file loaded successfully")
}

func Init() {
    InitLog()
    common.InitClickHouse()
    common.InitMySQL()
}


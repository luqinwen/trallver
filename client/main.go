package main

import (
    "log"
    "my_project/client/config"
    "my_project/client/service"
    "my_project/client/model" // 正确导入 model 包
    "github.com/spf13/viper"
)

func main() {
    // 读取配置文件
    viper.SetConfigName("client_config")
    viper.AddConfigPath("/root/config")
    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    log.Printf("Config file loaded successfully, log file path: %s", viper.GetString("log_file"))

    // 初始化日志
    config.InitLog()

    log.Println("Starting ICMP Probe Client")
    // 这里继续实现其他逻辑

    // 从配置文件加载探测任务
    task := &model.ProbeTask{ // 使用 model.ProbeTask 而不是 service.ProbeTask
        IP:        config.ClientConfig.Probe.IP,
        Count:     config.ClientConfig.Probe.Count,
        Threshold: config.ClientConfig.Probe.Threshold,
        Timeout:   config.ClientConfig.Probe.Timeout,
    }

    // 执行探测任务
    service.ExecuteProbeTask(task)

    // 上报探测结果
    service.ReportResultsToServer(task)
}

package main

import (
    "log"
    "my_project/client/config"
    "my_project/client/service"
    "my_project/client/model" // 正确导入 model 包
)

func main() {
    // 初始化配置文件
    config.InitConfig()

    // 初始化日志
    config.InitLog(config.ClientConfig.LogFile)

    log.Println("Starting ICMP Probe Client")

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

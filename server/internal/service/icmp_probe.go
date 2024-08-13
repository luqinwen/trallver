package service

import (
    "log"
    "my_project/server/internal/dao"
    "my_project/server/internal/model"
    "github.com/prometheus-community/pro-bing"
    "time"
)

// ExecuteProbeTask 执行探测任务
func ExecuteProbeTask(task model.ProbeTask) {
    pinger, err := probing.NewPinger(task.IP)
    if err != nil {
        log.Printf("Failed to create pinger: %v", err)
        return
    }

    pinger.Count = task.Count
    pinger.Timeout = time.Duration(task.Count) * time.Second

    pinger.OnRecv = func(pkt *probing.Packet) {
        log.Printf("Received packet from %s: time=%v", pkt.IPAddr, pkt.Rtt)
    }

    pinger.OnFinish = func(stats *probing.Statistics) {
        log.Printf("Probe finished. Packet loss: %v%%, Min RTT: %v, Max RTT: %v, Avg RTT: %v",
            stats.PacketLoss, stats.MinRtt, stats.MaxRtt, stats.AvgRtt)
        
        timestamp := time.Now().Unix()
        err := dao.StoreClickHouse(timestamp, task.IP, stats.PacketLoss, float64(stats.MinRtt.Microseconds()), float64(stats.MaxRtt.Microseconds()), float64(stats.AvgRtt.Microseconds()))
        if err != nil {
            log.Printf("Failed to store probe result to ClickHouse: %v", err)
        }

        if stats.PacketLoss > float64(task.Threshold) {
            ReportToPrometheus(stats)
        }
    }

    log.Printf("Starting probe to %s", task.IP)
    pinger.Run()
}

// ReportToPrometheus 上报探测结果到 Prometheus
func ReportToPrometheus(stats *probing.Statistics) {
    // 实现上报 Prometheus 的逻辑
    log.Printf("Reporting to Prometheus: %+v", stats)
}

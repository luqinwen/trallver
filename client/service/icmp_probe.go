package service

import (
    "log"
    "my_project/client/model"
    "github.com/prometheus-community/pro-bing"
    "time"
)

// ExecuteProbeTask 执行探测任务
func ExecuteProbeTask(task *model.ProbeTask) {
    pinger, err := probing.NewPinger(task.IP)
    if err != nil {
        log.Printf("Failed to create pinger: %v", err)
        return
    }

    pinger.Count = task.Count
    pinger.Timeout = time.Duration(task.Timeout) * time.Second

    pinger.OnRecv = func(pkt *probing.Packet) {
        log.Printf("Received packet from %s: time=%v", pkt.IPAddr, pkt.Rtt)
    }

    pinger.OnFinish = func(stats *probing.Statistics) {
        log.Printf("Probe finished. Packet loss: %v%%, Min RTT: %v, Max RTT: %v, Avg RTT: %v",
            stats.PacketLoss, stats.MinRtt, stats.MaxRtt, stats.AvgRtt)

        task.PacketLoss = stats.PacketLoss
        task.MinRTT = stats.MinRtt
        task.MaxRTT = stats.MaxRtt
        task.AvgRTT = stats.AvgRtt
    }

    log.Printf("Starting probe to %s", task.IP)
    pinger.Run()
}

// ReportResultsToServer 上报探测结果到Server端
func ReportResultsToServer(task *model.ProbeTask) {
    // 实现上报逻辑，比如使用HTTP POST请求发送数据到Server端
    log.Printf("Reporting results for IP: %s, PacketLoss: %v%% to server", task.IP, task.PacketLoss)
    // 可以使用类似于http.Post的方式发送探测结果到Server端
}

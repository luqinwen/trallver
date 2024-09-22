package service

import (
    "log"
    "my_project/client/config"
    "my_project/client/model"
    "net"
    "time"
    "github.com/prometheus-community/pro-bing"
)

func ExecuteProbeTask(task *model.ProbeTask) *model.ProbeResult {
    task.Status = model.StatusInProgress

    startTime := time.Now()
    ipString := net.IP(task.IP[:]).String()

    pinger, err := probing.NewPinger(ipString)
    if err != nil {
        log.Printf("Failed to create pinger: %v", err)
        task.Status = model.StatusFailed
        return nil
    }

    pinger.Count = int(task.Count)
    pinger.Timeout = time.Duration(task.Timeout) * time.Second

    var packetLoss float64
    var minRTT, maxRTT, avgRTT time.Duration

    pinger.OnRecv = func(pkt *probing.Packet) {
        log.Printf("Received packet from %s: time=%v", pkt.IPAddr, pkt.Rtt)
    }

    pinger.OnFinish = func(stats *probing.Statistics) {
        log.Printf("Probe finished. Packet loss: %v%%, Min RTT: %v, Max RTT: %v, Avg RTT: %v",
            stats.PacketLoss, stats.MinRtt, stats.MaxRtt, stats.AvgRtt)

        packetLoss = stats.PacketLoss
        minRTT = stats.MinRtt
        maxRTT = stats.MaxRtt
        avgRTT = stats.AvgRtt
    }

    log.Printf("Starting probe to %s", task.IP)
    pinger.Run()

    endTime := time.Now()
    processingLatency := endTime.Sub(startTime).Milliseconds()
    log.Printf("Processing latency: %v ms", processingLatency)

    task.Status = model.StatusCompleted

    result := &model.ProbeResult{
        TaskID:       task.ID,  // 从任务中获取并传递 TaskID
        IP:           task.IP,
        Timestamp:    time.Now().Unix(),
        PacketLoss:   uint8(packetLoss),
        MinRTT:       uint16(minRTT.Milliseconds()),
        MaxRTT:       uint16(maxRTT.Milliseconds()),
        AvgRTT:       uint16(avgRTT.Milliseconds()),
        DispatchTime: task.DispatchTime,
        Threshold:    config.ClientConfig.Probe.Threshold,
    }

    return result
}

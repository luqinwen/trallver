package service

import (
    "log"
    "my_project/client/model"
    "net"
    "time"
    "github.com/prometheus-community/pro-bing"
)

func ExecuteProbeTask(task *model.ProbeTaskDynamic) *model.ProbeResult {
    task.Status = model.StatusInProgress

    startTime := time.Now()
    
    // 从缓存中获取固定字段
    fixedFields := GetFixedFields()

    // 将 uint32 转换为 IPv4 地址字符串
    ipString := net.IPv4(byte(fixedFields.IP>>24), byte(fixedFields.IP>>16), byte(fixedFields.IP>>8), byte(fixedFields.IP)).String()

    pinger, err := probing.NewPinger(ipString)
    if err != nil {
        log.Printf("Failed to create pinger: %v", err)
        task.Status = model.StatusFailed
        return nil
    }

    // 解包固定字段中的 Timeout, Count, Threshold
    timeout, count, threshold := model.UnpackFixedFields(fixedFields.Packed)

    pinger.Count = int(count)                         // 使用解包后的Count
    pinger.Timeout = time.Duration(timeout) * time.Second // 使用解包后的Timeout

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

    log.Printf("Starting probe to %s", ipString)
    pinger.Run()

    endTime := time.Now()
    processingLatency := endTime.Sub(startTime).Milliseconds()
    log.Printf("Processing latency: %v ms", processingLatency)

    task.Status = model.StatusCompleted

    // 打包结果中的 PacketLoss 和 Threshold 字段
    packedFields := model.PackResultFields(uint8(packetLoss), threshold)

    result := &model.ProbeResult{
        TaskID:       task.ID,  // 从任务中获取并传递 TaskID
        IP:           fixedFields.IP, // 从缓存获取IP，保持为uint32
        Timestamp:    uint32(time.Now().Unix()),
        MinRTT:       uint16(minRTT.Milliseconds()),
        MaxRTT:       uint16(maxRTT.Milliseconds()),
        AvgRTT:       uint16(avgRTT.Milliseconds()),
        DispatchTime: task.DispatchTime,
        Packed:       packedFields, // 使用打包后的 PacketLoss 和 Threshold (16位)
    }

    return result
}

// 获取缓存中的固定字段
func GetFixedFields() model.ProbeTaskFixed {
    return fixedFieldsCache
}

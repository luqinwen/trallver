package service

import (
    "log"
    "my_project/client/model"
    "github.com/prometheus-community/pro-bing"
    "time"
    "bytes"
    "encoding/json"
    "net/http"
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

// ReportResultsToServer 上报探测结果到 Server 端
func ReportResultsToServer(task *model.ProbeTask) {
    // 定义服务器端的 API 端点
    serverURL := "http://your_server_ip:8080/probe_results" // 替换为实际的服务器地址

    // 构建上报的数据结构
    reportData := struct {
        IP         string        `json:"ip"`
        PacketLoss float64       `json:"packet_loss"`
        MinRTT     time.Duration `json:"min_rtt"`
        MaxRTT     time.Duration `json:"max_rtt"`
        AvgRTT     time.Duration `json:"avg_rtt"`
        Timestamp  time.Time     `json:"timestamp"`
    }{
        IP:         task.IP,
        PacketLoss: task.PacketLoss,
        MinRTT:     task.MinRTT,
        MaxRTT:     task.MaxRTT,
        AvgRTT:     task.AvgRTT,
        Timestamp:  time.Now(),
    }

    // 将数据结构编码为 JSON
    jsonData, err := json.Marshal(reportData)
    if err != nil {
        log.Printf("Failed to encode report data to JSON: %v", err)
        return
    }

    // 创建 HTTP POST 请求
    req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Failed to create HTTP request: %v", err)
        return
    }

    // 设置请求头
    req.Header.Set("Content-Type", "application/json")

    // 发送 HTTP 请求
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Failed to send probe result to server: %v", err)
        return
    }
    defer resp.Body.Close()

    // 处理服务器的响应
    if resp.StatusCode != http.StatusOK {
        log.Printf("Server returned non-OK status: %v", resp.Status)
    } else {
        log.Println("Probe results reported successfully to server")
    }
}
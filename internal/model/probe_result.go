package model

import "time"

// ProbeResult 定义探测结果的结构体
type ProbeResult struct {
    Timestamp   time.Time
    IP          string
    PacketLoss  float64
    MinRTT      time.Duration
    MaxRTT      time.Duration
    AvgRTT      time.Duration
}

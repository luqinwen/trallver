package model

import (
    "time"
)

// ProbeTask 定义探测任务的结构体
type ProbeTask struct {
    IP         string        // 探测目标的IP地址
    Count      int           // 探测的次数
    Port       int           // 探测目标的端口（ICMP可不设置）
    Threshold  int           // 丢包率阈值
    Timeout    int           // 探测超时时间（秒）
    CreatedAt  time.Time     // 任务创建时间
    UpdatedAt  time.Time     // 任务更新时间
    PacketLoss float64       // 丢包率
    MinRTT     time.Duration // 最小往返时间
    MaxRTT     time.Duration // 最大往返时间
    AvgRTT     time.Duration // 平均往返时间
}


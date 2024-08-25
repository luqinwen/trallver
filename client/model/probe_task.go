package model

import (
    "time"
)

type ProbeTask struct {
    IP        string    `json:"ip"`          // 探测目标的IP地址
    Count     int       `json:"count"`       // 探测的次数
    Port      int       `json:"port"`        // 探测目标的端口（ICMP可不设置）
    Threshold int       `json:"threshold"`   // 丢包率阈值
    Timeout   int       `json:"timeout"`     // 探测超时时间（秒）
    CreatedAt time.Time `json:"created_at"`  // 任务创建时间
    UpdatedAt time.Time `json:"updated_at"`  // 任务更新时间
}



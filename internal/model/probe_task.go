package model

import "time"

// ProbeTask 定义探测任务的结构体
type ProbeTask struct {
    IP        string    `json:"ip"`
    Count     int       `json:"count"`
    Port      int       `json:"port"`
    Threshold int       `json:"threshold"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

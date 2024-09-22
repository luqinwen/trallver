package model

type TaskStatus uint8

const (
    StatusPending    TaskStatus = iota // 0 - 任务待处理
    StatusInProgress                    // 1 - 任务进行中
    StatusCompleted                     // 2 - 任务已完成
    StatusFailed                        // 3 - 任务失败
)

type ProbeTask struct {
    ID           uint32   `json:"id"`            // 添加ID字段，类型为uint32
    IP           [16]byte `json:"ip"`           // 固定长度数组，保存IPv4或IPv6地址
    Count        uint8    `json:"count"`        // 任务的执行次数
    Timeout      uint16   `json:"timeout"`      // 超时时间（秒）
    DispatchTime int64    `json:"dispatch_time"`// Unix 时间戳 (秒)
    Status       TaskStatus `json:"status"`      // 任务状态
}



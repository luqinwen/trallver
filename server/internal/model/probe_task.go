package model

type TaskStatus uint8

const (
    StatusPending    TaskStatus = iota // 0 - 任务待处理
    StatusInProgress                    // 1 - 任务进行中
    StatusCompleted                     // 2 - 任务已完成
    StatusFailed                        // 3 - 任务失败
)


type ProbeTaskFixed struct {
    ID       uint32  `json:"id"`       // 唯一的任务ID
    IP       uint32  `json:"ip"`       // 优化为IPv4存储
    Packed   uint32  `json:"packed"`   // 打包Timeout, Count, Threshold
}


type ProbeTaskDynamic struct {
    DispatchTime uint32     `json:"dispatch_time"`// 动态字段，任务下发时间（使用uint32代替int64）
    ID           uint32     `json:"id"`           // 动态字段，任务ID
    Status       TaskStatus `json:"status"`       // 动态字段，任务状态
}

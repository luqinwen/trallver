package model


type ProbeResult struct {
    TaskID       uint32   `json:"task_id"`      // 添加TaskID字段
    IP           [16]byte `json:"ip"`           // 固定长度数组，保存IPv4地址
    Timestamp    int64    `json:"timestamp"`    // Unix 时间戳 (秒)
    PacketLoss   uint8    `json:"packet_loss"`  // 丢包率百分比
    MinRTT       uint16   `json:"min_rtt"`      // 最小往返时间（毫秒）
    MaxRTT       uint16   `json:"max_rtt"`      // 最大往返时间（毫秒）
    AvgRTT       uint16   `json:"avg_rtt"`      // 平均往返时间（毫秒）
    Threshold    uint8    `json:"threshold"`    // 丢包率阈值（百分比）
    DispatchTime int64    `json:"dispatch_time"`// Unix 时间戳 (秒)
}

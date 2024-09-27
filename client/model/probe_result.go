package model

type ProbeResult struct {
    TaskID       uint32   // 任务ID，4字节
    IP           uint32   // IPv4地址，4字节
    Timestamp    uint32   // 时间戳，4字节
    DispatchTime uint32   // 下发时间，4字节
    Packed       uint16   // 打包后的字段：PacketLoss、Threshold，2字节
    MinRTT       uint16   // 最小往返时间（毫秒），2字节
    MaxRTT       uint16   // 最大往返时间（毫秒），2字节
    AvgRTT       uint16   // 平均往返时间（毫秒），2字节
}

package service

import (
    "time"
    "my_project/server/internal/model"
)

// 创建固定任务字段，首次任务时调用
func CreateFixedTask() *model.ProbeTaskFixed {
    return &model.ProbeTaskFixed{
        IP:        134744072,           // 示例 IP 地址 8.8.8.8 对应的 uint32
        Packed:    model.PackFixedFields(5, 4, 10), // 打包 Timeout, Count, Threshold
    }
}


// 创建动态任务字段，每次任务都会调用
func CreateDynamicTask(taskID uint32) model.ProbeTaskDynamic {
    return model.ProbeTaskDynamic{
        ID:           taskID,             // 唯一任务ID
        DispatchTime: uint32(time.Now().Unix()),  // 任务下发时间
        Status:       model.StatusPending, // 初始状态为待处理
    }
}

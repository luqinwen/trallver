package model

type Confirmation struct {
    TaskID uint32     `json:"task_id"` // 任务的唯一ID，改为uint32以减少占用空间
    Status TaskStatus `json:"status"`  // 使用现有的TaskStatus枚举类型
}

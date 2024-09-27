package dao

import (
    "log"
    "my_project/server/internal/common"
    "my_project/server/internal/model"
)

// GetNextTaskID 获取下一个自增的 task_id
func GetNextTaskID() (uint32, error) {
    var taskID uint32
    // 查询 MySQL 中的下一个自增值
    err := common.MySQLDB.QueryRow("SELECT AUTO_INCREMENT FROM information_schema.tables WHERE table_name = 'probe_tasks' AND table_schema = DATABASE()").Scan(&taskID)
    if err != nil {
        log.Printf("Failed to get next task_id: %v", err)
        return 0, err
    }
    return taskID, nil
}

// StoreFixedTask 存储首次任务的固定字段到 MySQL（首次任务）
func StoreFixedTask(fixedTask *model.ProbeTaskFixed) (uint32, error) {
    stmt, err := common.MySQLDB.Prepare("INSERT INTO probe_task_fixed (ip, packed) VALUES (?, ?)")
    if err != nil {
        log.Printf("Error preparing statement for fixed task: %v", err)
        return 0, err
    }
    defer stmt.Close()

    // 将 IP 地址从 uint32 转换为 []byte
    ipBytes := []byte{
        byte(fixedTask.IP >> 24),
        byte(fixedTask.IP >> 16),
        byte(fixedTask.IP >> 8),
        byte(fixedTask.IP),
    }

    // 执行插入操作
    res, err := stmt.Exec(ipBytes, fixedTask.Packed)
    if err != nil {
        log.Printf("Error executing statement for fixed task: %v", err)
        return 0, err
    }

    // 获取插入的固定字段的 ID
    lastInsertID, err := res.LastInsertId()
    if err != nil {
        log.Printf("Error fetching last insert ID for fixed task: %v", err)
        return 0, err
    }

    log.Printf("Successfully inserted fixed task with ID: %d", lastInsertID)
    return uint32(lastInsertID), nil
}

// StoreProbeTask 存储动态字段的任务到 MySQL
func StoreProbeTask(dynamicTask *model.ProbeTaskDynamic, fixedTaskID uint32) error {
    // 准备 SQL 语句，插入动态字段（fixed_task_id、dispatch_time、status）
    stmt, err := common.MySQLDB.Prepare("INSERT INTO probe_tasks (fixed_task_id, dispatch_time, status) VALUES (?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement for dynamic task: %v", err)
        return err
    }
    defer stmt.Close()

    // 执行插入操作，并获取插入后的结果
    res, err := stmt.Exec(fixedTaskID, dynamicTask.DispatchTime, dynamicTask.Status)
    if err != nil {
        log.Printf("Error executing statement for dynamic task: %v", err)
        return err
    }

    // 获取自动生成的 ID 并赋值给 dynamicTask.ID
    lastInsertID, err := res.LastInsertId()
    if err != nil {
        log.Printf("Error fetching last insert ID for dynamic task: %v", err)
        return err
    }

    // 将 lastInsertID 转换为 uint32 并赋值给 dynamicTask.ID
    dynamicTask.ID = uint32(lastInsertID)
    log.Printf("Successfully inserted dynamic task with ID: %d", dynamicTask.ID)
    return nil
}

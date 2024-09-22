package dao

import (
    "my_project/server/internal/common"
    "my_project/server/internal/model"
    "log"
)

// StoreProbeTask 存储探测任务的元数据到 MySQL
func StoreProbeTask(task *model.ProbeTask) error {
    // 准备 SQL 语句，注意我们不需要插入 ID 字段，因为它是自动生成的
    stmt, err := common.MySQLDB.Prepare("INSERT INTO my_database.probe_tasks (ip, count, timeout, dispatch_time, status) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    // 将 IP 地址从 [16]byte 转换为 []byte
    ipBytes := task.IP[:]

    // 执行插入操作，并获取结果
    res, err := stmt.Exec(ipBytes, task.Count, task.Timeout, task.DispatchTime, task.Status)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        return err
    }

    // 获取自动生成的 ID 并赋值给 task.ID
    lastInsertID, err := res.LastInsertId()
    if err != nil {
        log.Printf("Error fetching last insert ID: %v", err)
        return err
    }

    // 将 lastInsertID 转换为 uint32 并赋值给 task.ID
    task.ID = uint32(lastInsertID)
    log.Printf("Successfully inserted probe task with ID: %d", task.ID)
    return nil
}

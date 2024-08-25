package dao

import (
    "my_project/server/internal/common"
    "my_project/server/internal/model"
    "log"
)

// StoreProbeTask 存储探测任务的元数据到 MySQL
func StoreProbeTask(task *model.ProbeTask) error {
    stmt, err := common.MySQLDB.Prepare("INSERT INTO my_database.probe_tasks (ip, count, port, threshold, timeout, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(task.IP, task.Count, task.Port, task.Threshold, task.Timeout, "pending", task.CreatedAt, task.UpdatedAt)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        return err
    }
    log.Printf("Successfully inserted probe task: %+v", task)
    return nil
}

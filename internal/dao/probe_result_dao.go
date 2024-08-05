package dao

import (
    "my_project/internal/common"
    "my_project/internal/model"
    "log"
)

// StoreProbeTask 存储探测任务的元数据到 MySQL
func StoreProbeTask(task *model.ProbeTask) error {
    stmt, err := common.MySQLDB.Prepare("INSERT INTO my_database.probe_tasks (ip, count, port, threshold, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(task.IP, task.Count, task.Port, task.Threshold, "pending", task.CreatedAt, task.UpdatedAt)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        return err
    }
    log.Printf("Successfully inserted probe task: %+v", task)
    return nil
}

// StoreClickHouse 存储探测结果到 ClickHouse
func StoreClickHouse(timestamp int64, ip string, packetLoss, minRtt, maxRtt, avgRtt float64) error {
    stmt, err := common.ClickHouseDB.Prepare("INSERT INTO my_database.my_table (timestamp, ip, packet_loss, min_rtt, max_rtt, avg_rtt) VALUES (?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(timestamp, ip, packetLoss, minRtt, maxRtt, avgRtt)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        return err
    }
    log.Printf("Successfully inserted into ClickHouse: timestamp=%d, ip=%s, packet_loss=%f, min_rtt=%f, max_rtt=%f, avg_rtt=%f", timestamp, ip, packetLoss, minRtt, maxRtt, avgRtt)
    return nil
}

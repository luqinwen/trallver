package dao

import (
    "my_project/server/internal/common"
    "log"
)

func StoreClickHouse(timestamp int64, ip string, packetLoss, minRtt, maxRtt, avgRtt float64) error {
    // 开始事务
    tx, err := common.ClickHouseDB.Begin()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return err
    }

    // 准备插入语句
    stmt, err := tx.Prepare("INSERT INTO my_database.my_table (timestamp, ip, packet_loss, min_rtt, max_rtt, avg_rtt) VALUES (?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    // 执行插入
    _, err = stmt.Exec(timestamp, ip, packetLoss, minRtt, maxRtt, avgRtt)
    if err != nil {
        log.Printf("Error executing statement: %v", err)
        // 如果插入失败，回滚事务
        tx.Rollback()
        return err
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        log.Printf("Failed to commit transaction: %v", err)
        return err
    }

    log.Printf("Successfully inserted into ClickHouse: timestamp=%d, ip=%s, packet_loss=%f, min_rtt=%f, max_rtt=%f, avg_rtt=%f", timestamp, ip, packetLoss, minRtt, maxRtt, avgRtt)
    return nil
}
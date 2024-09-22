package dao

import (
	"log"
	"my_project/server/internal/common"
	"net"
)

// 存储聚合结果到 aggregated_results 表
func StoreAggregatedResults(timestamp int64, avgPacketLoss float64, avgLatencyMs uint32) error {
    // 开始事务
    tx, err := common.ClickHouseDB.Begin()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return err
    }

    // 准备插入语句
    stmt, err := tx.Prepare("INSERT INTO my_database.aggregated_results (timestamp, avg_packet_loss, avg_latency_ms) VALUES (?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    // 执行插入
    _, err = stmt.Exec(timestamp, avgPacketLoss, avgLatencyMs)
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

    log.Printf("Successfully inserted into aggregated_results: timestamp=%d, avg_packet_loss=%f, avg_latency_ms=%d", timestamp, avgPacketLoss, avgLatencyMs)
    return nil
}

func StoreQueueResults(timestamp int64, taskID uint32, queueID int, ip [16]byte, packetLoss uint8, minRtt, maxRtt, avgRtt uint16, latencyMs uint32) error {
    // 开始事务
    tx, err := common.ClickHouseDB.Begin()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return err
    }

    // 准备插入语句
    stmt, err := tx.Prepare("INSERT INTO my_database.queue_results (timestamp, task_id, queue_id, ip, packet_loss, min_rtt, max_rtt, avg_rtt, latency_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()

    // 执行插入
    _, err = stmt.Exec(timestamp, taskID, queueID, ip[:], packetLoss, minRtt, maxRtt, avgRtt, latencyMs)
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

    log.Printf("Successfully inserted into queue_results: timestamp=%d, task_id=%d, queue_id=%d, ip=%s, packet_loss=%d, min_rtt=%d, max_rtt=%d, avg_rtt=%d, latency_ms=%d",
        timestamp, taskID, queueID, net.IP(ip[:]).String(), packetLoss, minRtt, maxRtt, avgRtt, latencyMs)
    return nil
}



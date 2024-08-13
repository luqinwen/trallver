package service

import (
    "encoding/json"
    "net/http"
    "my_project/server/internal/dao"
    "my_project/server/internal/model"
    "log"
    "time"
)

// HandleProbeTask 处理探测任务的HTTP请求
func HandleProbeTask(w http.ResponseWriter, r *http.Request) {
    var task model.ProbeTask
    err := json.NewDecoder(r.Body).Decode(&task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    task.CreatedAt = time.Now()
    task.UpdatedAt = time.Now()
    log.Printf("Received probe task: %+v", task)

    // 将探测任务存储到MySQL数据库
    err = dao.StoreProbeTask(&task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 异步执行探测任务
    go ExecuteProbeTask(task)
    w.WriteHeader(http.StatusAccepted)
}

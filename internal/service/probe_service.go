package service

import (
    "encoding/json"
    "net/http"
    "my_project/internal/dao"
    "my_project/internal/model"
    "log"
    "time"
)

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

    // 存储任务元数据到 MySQL
    err = dao.StoreProbeTask(&task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 异步执行探测任务
    go ExecuteProbeTask(task)

    w.WriteHeader(http.StatusAccepted)
}

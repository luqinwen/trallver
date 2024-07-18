package router

import (
    
    "github.com/gorilla/mux"
    "my_project/internal/service"
)

// InitializeRoutes 初始化路由
func InitializeRoutes() *mux.Router {
    router := mux.NewRouter()
    router.HandleFunc("/probe", service.HandleProbeTask).Methods("POST")
    return router
}

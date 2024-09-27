package common

// ServerConfig 保存服务器的配置
var ServerConfig struct {
    MySQL struct {
        User     string
        Password string
        Host     string
        Port     int
        Database string
    }
    ClickHouse struct {
        Host string
        Port int
    }
    LogFile string
}
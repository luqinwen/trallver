package config

import (
    "my_project/internal/common"
)

func Init() {
    InitLog()
    common.InitClickHouse()
    common.InitMySQL()
}

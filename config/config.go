package config

import (
    "my_project/internal/common"
)

func init() {
    common.InitClickHouse()
    InitLog()
}

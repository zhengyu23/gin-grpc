package global

import (
	"xrUncle/srvs/inventory_srv/config"

	"gorm.io/gorm"
)

// 全局变量
var (
	DB           *gorm.DB
	ServerConfig config.ServerConfig
	NacosConfig  config.NacosConfig
)

package global

import (
	"xrUncle/srvs/goods_srv/config"

	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

// 全局变量
var (
	DB           *gorm.DB // gorm DB
	ServerConfig config.ServerConfig
	NacosConfig  config.NacosConfig
	EsClient     *elastic.Client
)

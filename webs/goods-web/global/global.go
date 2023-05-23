package global

import (
	ut "github.com/go-playground/universal-translator"
	"xrUncle/webs/goods-web/config"
	"xrUncle/webs/goods-web/proto"
)

var (
	Trans        ut.Translator
	ServerConfig = &config.ServerConfig{}
	NacosConfig  = &config.NacosConfig{}
	GoodsSrvClient proto.GoodsClient
)

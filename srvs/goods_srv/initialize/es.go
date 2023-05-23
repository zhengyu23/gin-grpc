package initialize

import (
	"context"
	"fmt"
	"log"
	"os"
	"xrUncle/srvs/goods_srv/global"
	"xrUncle/srvs/goods_srv/model"

	"github.com/olivere/elastic/v7"
)

// InitEs 初始化 Elasticsearch
func InitEs() {
	// 初始化连接
	host := fmt.Sprintf("http://%s:%d", global.ServerConfig.EsInfo.Host, global.ServerConfig.EsInfo.Port)
	logger := log.New(os.Stdout, "chenbing", log.LstdFlags) // es Logger
	var err error
	// 新建es客户端 elastic
	global.EsClient, err = elastic.NewClient(elastic.SetURL(host), elastic.SetSniff(false),
		elastic.SetTraceLog(logger))
	if err != nil {
		panic(err)
	}

	// 通过es搜索"indices索引"

	exists, err := global.EsClient.IndexExists(model.EsGoods{}.GetIndexName()).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !exists {
		// 如果不存在则通过 Mapping 新建 Index
		_, err = global.EsClient.CreateIndex(model.EsGoods{}.GetIndexName()).BodyString(model.EsGoods{}.GetMapping()).Do(context.Background())
		if err != nil {
			panic(err)
		}
	}
}

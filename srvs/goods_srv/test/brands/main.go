package main

import (
	"context"
	"fmt"
	"xrUncle/srvs/goods_srv/proto"

	"google.golang.org/grpc"
)

var brandClient proto.GoodsClient
var conn *grpc.ClientConn

// Init 初始化服务
func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:8088", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	brandClient = proto.NewGoodsClient(conn)
}

// TestGetBrandList 测试获取商品列表
func TestGetBrandList() {
	rsq, err := brandClient.BrandList(context.Background(), &proto.BrandFilterRequest{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsq.Total)
	for _, brand := range rsq.Data {
		fmt.Println(brand.Name)
	}
}

func main() {
	Init()
	TestGetBrandList()
	conn.Close()
}

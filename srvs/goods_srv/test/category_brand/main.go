package main

import (
	"context"
	"fmt"
	"xrUncle/srvs/goods_srv/proto"
	"xrUncle/srvs/goods_srv/test/reMarshal"

	"google.golang.org/grpc"
)

var brandClient proto.GoodsClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	brandClient = proto.NewGoodsClient(conn)
}

// TestCategoryBrandList 测试获取商品分类与品牌列表
func TestCategoryBrandList() {
	rsq, err := brandClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsq.Total)
	fmt.Println(reMarshal.ReMarshal(rsq.Data))
}

// TestGetCategoryBrandList 测试获取商品分类与品牌列表详情
func TestGetCategoryBrandList() {
	rsq, err := brandClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 130366,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(reMarshal.ReMarshal(rsq.Data))
}

func main() {
	Init()
	//TestCategoryBrandList()
	TestGetCategoryBrandList()
	conn.Close()
}

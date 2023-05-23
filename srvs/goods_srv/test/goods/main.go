package main

import (
	"context"
	"fmt"
	"xrUncle/srvs/goods_srv/proto"
	"xrUncle/srvs/goods_srv/test/reMarshal"

	"google.golang.org/grpc"
)

var goodsClient proto.GoodsClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	goodsClient = proto.NewGoodsClient(conn)

}

// TestBatchGetGoods 测试批量获取商品信息
func TestBatchGetGoods() {
	rsp, err := goodsClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: []int32{421, 422, 423},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, good := range rsp.Data {
		fmt.Println(good.Name, good.ShopPrice)
	}
}

// TestGetGoodsDetail 测试获取商品详情
func TestGetGoodsDetail() {
	rsp, err := goodsClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: 421,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(reMarshal.ReMarshal(rsp))
}

// TestGoodsList 测试获取商品列表
func TestGoodsList() {
	//  ① 130361  ② 130370  ③ 130368
	rsp, err := goodsClient.GoodsList(context.Background(), &proto.GoodsFilterRequest{
		TopCategory: 130370,
		KeyWords:    "深海",
		PriceMin:    90,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, good := range rsp.Data {
		fmt.Println(good.Name, good.ShopPrice)
	}
}

func main() {
	Init()
	//TestGoodsList()
	//TestGetGoodsDetail()
	TestBatchGetGoods()
	//TestGetGoodsDetail()
	conn.Close()
}

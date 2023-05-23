package main

import (
	"context"
	"fmt"
	"xrUncle/srvs/goods_srv/proto"
	"xrUncle/srvs/goods_srv/test/reMarshal"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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

// TestGetCategoryList 测试获取商品分类列表
func TestGetCategoryList() {
	rsq, err := brandClient.GetAllCategorysList(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsq)
	fmt.Println(rsq.Total)
	//for _, category := range rsq.Data {
	//	fmt.Println(category.Name)
	//}
	fmt.Println(rsq.JsonData)
}

// TestGetSubCategoryList 测试获取商品分类子列表
func TestGetSubCategoryList() {
	rsq, err := brandClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: 130364,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(reMarshal.ReMarshal(rsq.SubCategorys))
	//fmt.Println(rsq.SubCategorys)
}

func main() {
	Init()
	//TestGetCategoryList()
	TestGetSubCategoryList()
	conn.Close()
}

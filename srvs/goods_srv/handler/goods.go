package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"xrUncle/srvs/goods_srv/global"
	"xrUncle/srvs/goods_srv/model"
	"xrUncle/srvs/goods_srv/proto"

	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GoodsServer 快速启动 grpc服务，proto内置实现接口的空方法
type GoodsServer struct {
	proto.UnimplementedGoodsServer
}

// ModelToResponse 商品详情
func ModelToResponse(goods model.Goods) proto.GoodsInfoResponse {
	return proto.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.CategoryID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SoldNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		Images:          goods.Images,
		DescImages:      goods.DescImages,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		Category: &proto.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
}

// 商品接口

// GoodsList 商品列表
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	//使用es的目的是搜索出商品的id来，通过id拿到具体的字段信息是通过mysql来完成
	//我们使用es是用来做搜索的， 是否应该将所有的mysql字段全部在es中保存一份
	//es用来做搜索，这个时候我们一般只把搜索和过滤的字段信息保存到es中
	//es可以用来当做mysql使用， 但是实际上mysql和es之间是互补的关系， 一般mysql用来做存储使用，es用来做搜索使用
	//es想要提高性能， 就要将es的内存设置的够大， 1k 2k

	// 功能调用：关键词搜索、查询新品、查询热门商品、通过价格区间筛选、通过商品分类筛选 -> 对过滤有要求
	goodsListResponse := &proto.GoodsListResponse{}

	q := elastic.NewBoolQuery()
	localDB := global.DB.Model(model.Goods{}) // 局部 DB
	if req.KeyWords != "" {
		q = q.Must(elastic.NewMultiMatchQuery(req.KeyWords, "name", "goods_brief"))
	}
	if req.IsHot {
		localDB = localDB.Where(model.Goods{IsHot: true})
		q = q.Filter(elastic.NewTermQuery("is_hot", req.IsHot))
	}
	if req.IsNew {
		q = q.Filter(elastic.NewTermQuery("is_new", req.IsNew))
	}
	if req.PriceMin > 0 { // 最低价格
		q = q.Filter(elastic.NewRangeQuery("shop_price").Gte(req.PriceMin))
	}
	if req.PriceMax > 0 { // 最高价格
		q = q.Filter(elastic.NewRangeQuery("shop_price").Lte(req.PriceMax))
	}
	if req.Brand > 0 { // 品牌筛选
		q = q.Filter(elastic.NewTermQuery("brands_id", req.Brand))
	}

	// 通过category查询商品 -> 子查询
	var subQuery string
	categoryIds := make([]interface{}, 0)
	if req.TopCategory > 0 {
		var category model.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}

		// 子查询 -> 级别1的查询数量最多
		if category.Level == 1 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE parent_category_id IN (SELECT id FROM category WHERE parent_category_id=%d)", req.TopCategory)
		} else if category.Level == 2 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE parent_category_id=%d", req.TopCategory)
		} else if category.Level == 3 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE id=%d", req.TopCategory)
		}

		type Result struct {
			ID int32
		}
		var results []Result
		global.DB.Model(model.Category{}).Raw(subQuery).Scan(&results)
		for _, re := range results {
			categoryIds = append(categoryIds, re.ID)
		}

		//生成terms查询
		q = q.Filter(elastic.NewTermsQuery("category_id", categoryIds...))
	}
	//分页
	//if req.Pages == 0 {
	//	req.Pages = 1
	//} // 因为从第0页开始
	switch {
	case req.PagePerNums > 100:
		req.PagePerNums = 100
	case req.PagePerNums <= 0:
		req.PagePerNums = 10
	}
	//result, err := global.EsClient.Search().Index(model.EsGoods{}.GetIndexName()).Query(q).From(int(req.Pages)).Size(int(req.PagePerNums)).Do(context.Background())
	result, err := global.EsClient.Search().Index(model.EsGoods{}.GetIndexName()).Query(q).From(int(req.Pages)).Do(context.Background())
	if err != nil {
		return nil, err
	}

	goodsIds := make([]int32, 0)
	goodsListResponse.Total = int32(result.Hits.TotalHits.Value)
	for _, value := range result.Hits.Hits {
		goods := model.EsGoods{}
		_ = json.Unmarshal(value.Source, &goods)
		goodsIds = append(goodsIds, goods.ID)
	}

	// 查询id在某个数组中的值 Find(&goods, goodsIds)
	var goods []model.Goods
	if len(goodsIds) != 0 {
		re := localDB.Preload("Category").Preload("Brands").Find(&goods, goodsIds)
		if re.Error != nil {
			return nil, re.Error
		}
	}

	for _, good := range goods {
		goodsInfoResponse := ModelToResponse(good)
		goodsListResponse.Data = append(goodsListResponse.Data, &goodsInfoResponse)
	}

	return goodsListResponse, nil
}

// BatchGetGoods 批量获取商品信息,通过req.Id (现在用户提交订单有多个商品，得批量查询商品的信息)
func (s *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	// Id                   []int32
	goodsListResponse := &proto.GoodsListResponse{}
	fmt.Println("访问了 goods_srv-handler-BatchGetGoods()")
	var goods []model.Goods
	//result := global.DB.Where(&goods,req.Id) // 没有执行？？
	//result := global.DB.Where(&goods,req.Id).Find(&goods) // where是加sql语句的，真正执行可以用Find()
	result := global.DB.Where(req.Id).Find(&goods) //
	for _, good := range goods {
		goodsInfoResponse := ModelToResponse(good)
		goodsListResponse.Data = append(goodsListResponse.Data, &goodsInfoResponse)
	}
	goodsListResponse.Total = int32(result.RowsAffected)
	return goodsListResponse, nil
}

// GetGoodsDetail 获取商品详情,通过req.Id
func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var goods model.Goods

	if result := global.DB.Preload("Category").Preload("Brands").First(&goods, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	goodsInforesponse := ModelToResponse(goods)
	return &goodsInforesponse, nil
}

// CreateGoods 添加商品
func (s *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	// 1. 查询商品分类是否存在
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	// 2. 查询品牌是否存在
	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	//先检查redis中是否有这个token
	//防止同一个token的数据重复插入到数据库中，如果redis中没有这个token则放入redis
	// 这里没有看见图片文件是如何上传的，在微服务中，普通的文件上传已经不再适用
	// 需要使用第三方的存储

	// 3. 将已有数据整合到goods结构体中
	goods := model.Goods{
		Brands:          brand,
		BrandsID:        brand.ID,
		Category:        category,
		CategoryID:      category.ID,
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		MarketPrice:     req.MarketPrice,
		ShopPrice:       req.ShopPrice,
		GoodsBrief:      req.GoodsBrief,
		ShipFree:        req.ShipFree,
		Images:          req.Images,
		DescImages:      req.DescImages,
		GoodsFrontImage: req.GoodsFrontImage,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		OnSale:          req.OnSale,
	}

	// 4.1 开始数据库事务
	tx := global.DB.Begin()
	result := tx.Save(&goods) // 会调用添加商品的钩子
	if result.Error != nil {
		// 如果事务执行失败,则回滚
		tx.Rollback()
		return nil, result.Error
	}
	// 4.2 数据库事务提交(事务完成,数据库数据已修改)
	tx.Commit()

	return &proto.GoodsInfoResponse{
		Id: goods.ID,
	}, nil
}

// DeleteGoods 删除商品
func (s *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Goods{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	return &emptypb.Empty{}, nil
}

// UpdateGoods 更新商品
func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	var goods model.Goods

	// **首先检查即将存储的信息是否存在**

	// 1. 通过商品ID检查商品是否存在
	if result := global.DB.First(&goods, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	// 2. 通过商品分类ID检查商品分类是否存在
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在") // 参数错误
	}
	// 3. 通过品牌ID检查品牌是否存在
	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	// 4. 将已有信息整合到goods结构体
	{
		goods.Brands = brand
		goods.BrandsID = brand.ID
		goods.Category = category
		goods.CategoryID = category.ID
		goods.Name = req.Name
		goods.GoodsSn = req.GoodsSn
		goods.MarketPrice = req.MarketPrice
		goods.ShopPrice = req.ShopPrice
		goods.Images = req.Images
		goods.DescImages = req.DescImages
		goods.GoodsFrontImage = req.GoodsFrontImage
		goods.IsNew = req.IsNew
		goods.IsHot = req.IsHot
		goods.OnSale = req.OnSale
	}

	// 5. 通过goods结构体将数据保存到数据库
	global.DB.Save(&goods)
	return &emptypb.Empty{}, nil
}

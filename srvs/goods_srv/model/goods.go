package model

import (
	"context"
	"strconv"
	"xrUncle/srvs/goods_srv/global"

	"gorm.io/gorm"
)

// First : 分类信息表、品牌表
// Second: 多对多关系

// gorm: 类型； 能否为null；设置为 null还是设置为空？
// 实际开发过程中，尽量设置为 not null -> 技术指路 https://zhuanlan.zhihu.com/p/73997266

// Category 商品分类模板
type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategoryID int32       `json:"parent"`
	ParentCategory   *Category   `json:"-"` // 外键，自己指向自己得使用指针
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	// 尽量将 int 设置成 int32， 因为proto中没有int类型，可减少转换操作
	Level int32 `gorm:"type:int;not null;default:1" json:"level"` // 商品分类的高度,即层级关系
	IsTab bool  `gorm:"default:false;not null" json:"is_tab"`
}

// Brands 品牌模板
type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);default:'';not null"`
}

// GoodsCategoryBrand 商品分类与品牌模板
// 商品分类结构体和品牌结构体之间的多对多关系表
// 联合唯一索引 - gorm
type GoodsCategoryBrand struct {
	BaseModel
	// 外键
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category
	// 外键
	BrandsID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Brands   Brands
}

// TableName 重载表名 - 钩子 - 生成表时调用 - gorm
func (GoodsCategoryBrand) TableName() string {
	// 数据库默认 -> goods_category_brand
	// 改成
	return "goodsCategoryBrand"
}

// Banner 轮播图模板 -> 可有商业行为
type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32  `gorm:"type:int;default:1;not null"`
}

// Goods 商品模板
type Goods struct {
	BaseModel

	// 商品分类外键
	CategoryID int32 `gorm:"type:int;not null"`
	Category   Category
	// 品牌外键
	BrandsID int32 `gorm:"type:int;not null"`
	Brands   Brands

	OnSale   bool `gorm:"default:false;not null"` // 是否已经上架
	ShipFree bool `gorm:"default:false;not null"` // 是否免运费
	IsNew    bool `gorm:"default:false;not null"` // 是否是新品
	IsHot    bool `gorm:"default:false;not null"` // 是否是热门商品 -> 广告商品

	Name     string `gorm:"type:varchar(50);not null"`   // 商品名称
	GoodsSn  string `gorm:"type:varchar(50);not null"`   // 商品编号
	ClickNum int32  `gorm:"type:int;default:0;not null"` // 点击数
	SoldNum  int32  `gorm:"type:int;default:0;not null"` // 销量
	FavNum   int32  `gorm:"type:int;default:0;not null"` // 收藏量
	//Stock // 库存暂不设置
	// float32在数据库内会自动对应成 float类型， 不像 int在数据库内会自动对应成 bigint类型
	MarketPrice     float32  `gorm:"not null"`                    // 商品价格
	ShopPrice       float32  `gorm:"not null"`                    // 销售价格
	GoodsBrief      string   `gorm:"type:varchar(100);not null"`  // 商品简介
	Images          GormList `gorm:"type:varchar(1000);not null"` // 商品浏览页图片
	DescImages      GormList `gorm:"type:varchar(1000);not null"` //商品详情页图片
	GoodsFrontImage string   `gorm:"type:varchar(200);not null"`  // 封面
}

// GoodsImages 商品图片模板
// 为什么不直接放在商品信息表里？，不合适，表joint会降低性能
type GoodsImages struct {
	GoodsID int
	Image   string
}

// AfterCreate 添加商品的钩子 - gorm
// 添加商品同时添加商品信息到es服务器内
func (g *Goods) AfterCreate(tx *gorm.DB) (err error) {
	esModel := EsGoods{
		ID:          g.ID,
		CategoryID:  g.CategoryID,
		BrandsID:    g.BrandsID,
		OnSale:      g.OnSale,
		ShipFree:    g.ShipFree,
		IsNew:       g.IsNew,
		IsHot:       g.IsHot,
		Name:        g.Name,
		ClickNum:    g.ClickNum,
		SoldNum:     g.SoldNum,
		FavNum:      g.FavNum,
		MarketPrice: g.MarketPrice,
		GoodsBrief:  g.GoodsBrief,
		ShopPrice:   g.ShopPrice,
	}

	_, err = global.EsClient.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// AfterUpdate 更新商品的钩子 - gorm
// 更新商品同时更新es服务器内商品信息
func (g *Goods) AfterUpdate(tx *gorm.DB) (err error) {
	esModel := EsGoods{
		ID:          g.ID,
		CategoryID:  g.CategoryID,
		BrandsID:    g.BrandsID,
		OnSale:      g.OnSale,
		ShipFree:    g.ShipFree,
		IsNew:       g.IsNew,
		IsHot:       g.IsHot,
		Name:        g.Name,
		ClickNum:    g.ClickNum,
		SoldNum:     g.SoldNum,
		FavNum:      g.FavNum,
		MarketPrice: g.MarketPrice,
		GoodsBrief:  g.GoodsBrief,
		ShopPrice:   g.ShopPrice,
	}

	_, err = global.EsClient.Update().Index(esModel.GetIndexName()).
		Doc(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// AfterDelete 删除商品的钩子
// 删除商品同时删除es服务器内商品
func (g *Goods) AfterDelete(tx *gorm.DB) (err error) {
	_, err = global.EsClient.Delete().Index(EsGoods{}.GetIndexName()).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

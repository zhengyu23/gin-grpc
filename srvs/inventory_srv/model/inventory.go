package model

import (
	"database/sql/driver"
	"encoding/json"
)

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index"`
	Stocks  int32 `gorm:"type:int"`
	Version int32 `gorm:"type:int"` //分布式锁的乐观锁
}

type StockSellDetail struct {
	OrderSn string `gorm:"type:varchar(200);index:idx_order_sn,unique"`
	Status  int32  `gorm:"type:varchar(200)"` //1 表示已扣减 2. 表示已归还
	// GormList(string: struct > json) 指明每件商品扣了多少件
	Detail GoodsDetailList `gorm:"type:varchar(200)"`
}

type GoodsDetail struct {
	Goods int32
	Num   int32
}

type GoodsDetailList []GoodsDetail

func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

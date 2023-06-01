package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
	"xrUncle/srvs/inventory_srv/model"
)

func main() {
	dsn := "root:root@tcp(192.168.31.138:3306)/chenbing_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold // 慢 SQL 阈值
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: false,       // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color // 禁用彩色打印
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	//_ = db.AutoMigrate(&model.Inventory{})
	//_ = db.AutoMigrate(&model.StockSellDetail{})

	//orderDetail := model.StockSellDetail{
	//	OrderSn: "imooc-bobby",
	//	Status:  1,
	//	Detail:  []model.GoodsDetail{{1,2},{2,3}},
	//}
	//db.Create(&orderDetail)

	var sellDetail model.StockSellDetail
	db.Where(model.StockSellDetail{OrderSn: "imooc-bobby"}).First(&sellDetail)
	fmt.Println(sellDetail.Detail)
}

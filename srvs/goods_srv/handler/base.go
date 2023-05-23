package handler

import (
	"fmt"

	"gorm.io/gorm"
)

// Paginate 数据分页 - GORM
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		// page: 页数
		// pageSize: 每页数量

		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		// offset: 从第几个数据开始

		offset := (page - 1) * pageSize
		fmt.Println(offset)
		return db.Offset(offset).Limit(pageSize)
		//db.Offset(5)
		//db.Limit(10)
		//return db
	}
}

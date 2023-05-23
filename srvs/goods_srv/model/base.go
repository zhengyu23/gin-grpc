package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type GormList []string

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan 实现 sql.Scanner 接口， Scan 将 value 扫描至 Jsonb
func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

// BaseModel 用户-公共模板
type BaseModel struct {
	ID        int32     `gorm:"primarykey;type:bigint" json:"id"` // 为什么使用int32？ bigint(数据库) 减少外键创建失败的可能性
	CreatedAt time.Time `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"-"`
	//DeletedAt	gorm.DeletedAt // 删除时间
	IsDeleted bool `json:"-"` // `gorm:"column:is_deleted"`
}

package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类模型
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(50);not null"`
	ParentID  uint           `json:"parent_id" gorm:"default:0"`
	Sort      int            `json:"sort" gorm:"default:0"`
	Status    int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-启用 0-禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (Category) TableName() string {
	return "categories"
}

package model

import (
	"time"

	"gorm.io/gorm"
)

// Ticket 门票模型
type Ticket struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"type:varchar(200);not null"`
	Description string         `json:"description" gorm:"type:text"`
	Image       string         `json:"image" gorm:"type:varchar(255)"`
	CategoryID  uint           `json:"category_id" gorm:"not null;index"`
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	OriginalPrice float64      `json:"original_price" gorm:"type:decimal(10,2)"`
	Stock       int            `json:"stock" gorm:"not null;default:0"`
	Sold        int            `json:"sold" gorm:"not null;default:0"`
	Status      int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-上架 0-下架
	StartTime   *time.Time     `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	Location    string         `json:"location" gorm:"type:varchar(255)"`
	Sort        int            `json:"sort" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Category Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// TableName 设置表名
func (Ticket) TableName() string {
	return "tickets"
}

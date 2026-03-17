package model

import (
	"time"

	"gorm.io/gorm"
)

// SeckillActivityStatus 秒杀活动状态
const (
	SeckillActivityStatusActive   = 1 // 进行中
	SeckillActivityStatusInactive = 0 // 未开始或已结束
)

// SeckillActivity 秒杀活动模型
type SeckillActivity struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null"`
	TicketID    uint           `json:"ticket_id" gorm:"not null;index"`
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	TotalStock  int            `json:"total_stock" gorm:"not null;default:0"`
	AvailableStock int         `json:"available_stock" gorm:"not null;default:0"`
	StartTime   time.Time      `json:"start_time" gorm:"not null"`
	EndTime     time.Time      `json:"end_time" gorm:"not null"`
	Status      int8           `json:"status" gorm:"type:tinyint;default:0"` // 0-未开始/已结束 1-进行中
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Ticket Ticket `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
}

// TableName 设置表名
func (SeckillActivity) TableName() string {
	return "seckill_activities"
}

// SeckillStock 秒杀库存模型
type SeckillStock struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ActivityID uint           `json:"activity_id" gorm:"not null;uniqueIndex"`
	TotalStock int            `json:"total_stock" gorm:"not null;default:0"`
	UsedStock  int            `json:"used_stock" gorm:"not null;default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Activity SeckillActivity `json:"activity,omitempty" gorm:"foreignKey:ActivityID"`
}

// TableName 设置表名
func (SeckillStock) TableName() string {
	return "seckill_stock"
}

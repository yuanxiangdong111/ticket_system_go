package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null"`
	Email     string         `json:"email" gorm:"type:varchar(100);uniqueIndex"`
	Phone     string         `json:"phone" gorm:"type:varchar(20)"`
	Nickname  string         `json:"nickname" gorm:"type:varchar(50)"`
	Avatar    string         `json:"avatar" gorm:"type:varchar(255)"`
	Status    int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-正常 0-禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

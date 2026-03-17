package model

import (
	"time"

	"gorm.io/gorm"
)

// CouponType 优惠券类型
const (
	CouponTypeCash     = 1 // 满减券
	CouponTypeDiscount = 2 // 折扣券
	CouponTypeSeckill  = 3 // 秒杀券
)

// CouponStatus 优惠券状态
const (
	CouponStatusAvailable = 1 // 可用
	CouponStatusDisabled  = 0 // 不可用
)

// UserCouponStatus 用户优惠券状态
const (
	UserCouponStatusUnused    = 1 // 未使用
	UserCouponStatusUsed      = 2 // 已使用
	UserCouponStatusExpired   = 3 // 已过期
)

// Coupon 优惠券模型
type Coupon struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null"`
	Type        int8           `json:"type" gorm:"type:tinyint;not null"` // 1-满减券 2-折扣券 3-秒杀券
	Discount    float64        `json:"discount" gorm:"type:decimal(10,2);not null"`
	MinAmount   float64        `json:"min_amount" gorm:"type:decimal(10,2)"`   // 满减券最低消费金额
	MaxDiscount float64        `json:"max_discount" gorm:"type:decimal(10,2)"` // 折扣券最大折扣金额
	TotalCount  int            `json:"total_count" gorm:"not null;default:0"`
	UsedCount   int            `json:"used_count" gorm:"not null;default:0"`
	StartTime   time.Time      `json:"start_time" gorm:"not null"`
	EndTime     time.Time      `json:"end_time" gorm:"not null"`
	Status      int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-可用 0-不可用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (Coupon) TableName() string {
	return "coupons"
}

// UserCoupon 用户优惠券关联模型
type UserCoupon struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	CouponID      uint           `json:"coupon_id" gorm:"not null;index"`
	Status        int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-未使用 2-已使用 3-已过期
	ObtainTime    time.Time      `json:"obtain_time" gorm:"default:current_timestamp"`
	UsedTime      *time.Time     `json:"used_time"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Coupon Coupon  `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
}

// TableName 设置表名
func (UserCoupon) TableName() string {
	return "user_coupons"
}

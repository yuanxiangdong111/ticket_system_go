package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
const (
	OrderStatusPending   = 1 // 待支付
	OrderStatusPaid      = 2 // 已支付
	OrderStatusCancelled = 3 // 已取消
	OrderStatusRefunded  = 4 // 已退款
)

// Order 订单模型
type Order struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	OrderNo        string         `json:"order_no" gorm:"type:varchar(32);uniqueIndex;not null"`
	UserID         uint           `json:"user_id" gorm:"not null;index"`
	TotalAmount    float64        `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	DiscountAmount float64        `json:"discount_amount" gorm:"type:decimal(10,2);default:0.00"`
	PayAmount      float64        `json:"pay_amount" gorm:"type:decimal(10,2);not null"`
	Status         int8           `json:"status" gorm:"type:tinyint;default:1"` // 1-待支付 2-已支付 3-已取消 4-已退款
	PayTime        *time.Time     `json:"pay_time"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User        User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Items       []OrderItem   `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	OrderCoupons []OrderCoupon `json:"order_coupons,omitempty" gorm:"foreignKey:OrderID"`
}

// TableName 设置表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单详情模型
type OrderItem struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	OrderID    uint           `json:"order_id" gorm:"not null;index"`
	TicketID   uint           `json:"ticket_id" gorm:"not null;index"`
	Quantity   int            `json:"quantity" gorm:"not null;default:1"`
	Price      float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	TotalPrice float64        `json:"total_price" gorm:"type:decimal(10,2);not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Order  Order  `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Ticket Ticket `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
}

// TableName 设置表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderCoupon 订单优惠券关联模型
type OrderCoupon struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	OrderID    uint           `json:"order_id" gorm:"not null;index"`
	CouponID   uint           `json:"coupon_id" gorm:"not null;index"`
	UserCouponID uint         `json:"user_coupon_id" gorm:"not null;index"`
	Discount   float64        `json:"discount" gorm:"type:decimal(10,2);not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Order     Order     `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Coupon    Coupon    `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
	UserCoupon UserCoupon `json:"user_coupon,omitempty" gorm:"foreignKey:UserCouponID"`
}

// TableName 设置表名
func (OrderCoupon) TableName() string {
	return "order_coupons"
}

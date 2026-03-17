package dao

import (
	"ticket_system/internal/model"
	"ticket_system/pkg/database"

	"gorm.io/gorm"
)

// OrderDAO 订单数据访问对象
type OrderDAO struct {
	db *gorm.DB
}

// NewOrderDAO 创建订单DAO实例
func NewOrderDAO() *OrderDAO {
	return &OrderDAO{db: database.DB}
}

// Create 创建订单
func (d *OrderDAO) Create(order *model.Order) error {
	return d.db.Create(order).Error
}

// GetByID 根据ID获取订单
func (d *OrderDAO) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := d.db.Preload("Items.Ticket").Preload("OrderCoupons").First(&order, id).Error
	return &order, err
}

// GetByOrderNo 根据订单号获取订单
func (d *OrderDAO) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := d.db.Preload("Items.Ticket").Preload("OrderCoupons").Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// GetByUserID 根据用户ID获取订单列表
func (d *OrderDAO) GetByUserID(userID uint, status int8, offset, limit int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	query := d.db.Preload("Items.Ticket").Preload("OrderCoupons").Where("user_id = ?", userID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total).Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders)
	return orders, total, nil
}

// Update 更新订单
func (d *OrderDAO) Update(order *model.Order) error {
	return d.db.Save(order).Error
}

// UpdateStatus 更新订单状态
func (d *OrderDAO) UpdateStatus(orderID uint, status int8) error {
	return d.db.Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// UpdatePayment 更新支付信息
func (d *OrderDAO) UpdatePayment(orderID uint, payTime string) error {
	return d.db.Model(&model.Order{}).Where("id = ?", orderID).Updates(
		map[string]interface{}{"status": model.OrderStatusPaid, "pay_time": payTime}).Error
}

// CreateOrderItem 创建订单详情
func (d *OrderDAO) CreateOrderItem(item *model.OrderItem) error {
	return d.db.Create(item).Error
}

// GetOrderItems 获取订单详情
func (d *OrderDAO) GetOrderItems(orderID uint) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	err := d.db.Preload("Ticket").Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// CreateOrderCoupon 创建订单优惠券关联
func (d *OrderDAO) CreateOrderCoupon(orderCoupon *model.OrderCoupon) error {
	return d.db.Create(orderCoupon).Error
}

// GetOrderCoupons 获取订单使用的优惠券
func (d *OrderDAO) GetOrderCoupons(orderID uint) ([]*model.OrderCoupon, error) {
	var orderCoupons []*model.OrderCoupon
	err := d.db.Preload("Coupon").Preload("UserCoupon").Where("order_id = ?", orderID).Find(&orderCoupons).Error
	return orderCoupons, err
}

// GetPendingOrders 获取待支付订单
func (d *OrderDAO) GetPendingOrders(timeoutHours int) ([]*model.Order, error) {
	var orders []*model.Order
	err := d.db.Where("status = ? AND created_at < NOW() - INTERVAL ? HOUR",
		model.OrderStatusPending, timeoutHours).Find(&orders).Error
	return orders, err
}

// DeleteOrder 删除订单（物理删除，谨慎使用）
func (d *OrderDAO) DeleteOrder(orderID uint) error {
	return d.db.Delete(&model.Order{}, orderID).Error
}

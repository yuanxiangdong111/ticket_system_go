package service

import (
	"errors"
	"fmt"
	"ticket_system/internal/dao"
	"ticket_system/internal/model"
	"ticket_system/pkg/util"
	"time"
)

// OrderService 订单服务
type OrderService struct {
	orderDAO        *dao.OrderDAO
	couponDAO       *dao.CouponDAO
	userCouponDAO   *dao.UserCouponDAO
	ticketDAO       *dao.TicketDAO
	couponService   *CouponService
}

// NewOrderService 创建订单服务实例
func NewOrderService() *OrderService {
	return &OrderService{
		orderDAO:        dao.NewOrderDAO(),
		couponDAO:       dao.NewCouponDAO(),
		userCouponDAO:   dao.NewUserCouponDAO(),
		ticketDAO:       dao.NewTicketDAO(),
		couponService:   NewCouponService(),
	}
}

// OrderTicket 订单门票项
type OrderTicket struct {
	TicketID uint `json:"ticket_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,min=1"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Tickets      []OrderTicket `json:"tickets" binding:"required"`
	UserCouponIDs []uint       `json:"user_coupon_ids"`
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID uint, req *CreateOrderRequest) (*model.Order, error) {
	// 1. 验证门票并计算总金额
	var totalAmount float64
	var orderItems []*model.OrderItem
	var ticketIDs []uint
	var quantityChanges []int

	for _, ticketReq := range req.Tickets {
		ticket, err := s.ticketDAO.GetByID(ticketReq.TicketID)
		if err != nil {
			return nil, fmt.Errorf("门票不存在: %v", err)
		}
		if ticket.Status != 1 {
			return nil, errors.New("门票已下架")
		}
		if ticket.Stock < ticketReq.Quantity {
			return nil, errors.New("库存不足")
		}

		itemAmount := ticket.Price * float64(ticketReq.Quantity)
		totalAmount += itemAmount

		orderItem := &model.OrderItem{
			TicketID:   ticket.ID,
			Quantity:   ticketReq.Quantity,
			Price:      ticket.Price,
			TotalPrice: itemAmount,
		}
		orderItems = append(orderItems, orderItem)

		ticketIDs = append(ticketIDs, ticket.ID)
		quantityChanges = append(quantityChanges, -ticketReq.Quantity)
	}

	// 2. 处理优惠券
	var finalPrice = totalAmount
	var discountAmount float64
	var usedCoupons []*model.UserCoupon

	if len(req.UserCouponIDs) > 0 {
		validCoupons, err := s.couponService.CheckCouponsAvailable(userID, req.UserCouponIDs)
		if err != nil {
			return nil, err
		}

		if len(validCoupons) > 0 {
			finalPrice, discountAmount, err = s.couponService.CalculateFinalPrice(totalAmount, validCoupons)
			if err != nil {
				return nil, err
			}
			usedCoupons = validCoupons
		}
	}

	// 3. 预减库存（使用事务）
	if err := s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges); err != nil {
		return nil, errors.New("库存更新失败")
	}

	// 4. 创建订单
	orderNo := s.generateOrderNo()
	order := &model.Order{
		OrderNo:        orderNo,
		UserID:         userID,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		PayAmount:      finalPrice,
		Status:         model.OrderStatusPending,
		Items:          orderItems,
	}

	// 5. 使用事务创建订单和相关数据
	tx := s.orderDAO.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// 回滚库存
			for i := range ticketIDs {
				quantityChanges[i] = -quantityChanges[i]
			}
			s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
		}
	}()

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		// 回滚库存
		for i := range ticketIDs {
			quantityChanges[i] = -quantityChanges[i]
		}
		s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
		return nil, err
	}

	// 创建订单详情
	for _, item := range orderItems {
		item.OrderID = order.ID
		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			// 回滚库存
			for i := range ticketIDs {
				quantityChanges[i] = -quantityChanges[i]
			}
			s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
			return nil, err
		}
	}

	// 创建订单优惠券关联
	for _, userCoupon := range usedCoupons {
		orderCoupon := &model.OrderCoupon{
			OrderID:      order.ID,
			CouponID:     userCoupon.CouponID,
			UserCouponID: userCoupon.ID,
			Discount:     0,
		}

		// 计算该优惠券的实际折扣
		discount, _ := s.couponService.CalculateCouponDiscount(totalAmount, &userCoupon.Coupon)
		orderCoupon.Discount = discount

		if err := tx.Create(orderCoupon).Error; err != nil {
			tx.Rollback()
			// 回滚库存
			for i := range ticketIDs {
				quantityChanges[i] = -quantityChanges[i]
			}
			s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
			return nil, err
		}

		// 更新用户优惠券状态
		if err := tx.Model(&model.UserCoupon{}).Where("id = ?", userCoupon.ID).
			Update("status", model.UserCouponStatusUsed).Error; err != nil {
			tx.Rollback()
			// 回滚库存
			for i := range ticketIDs {
				quantityChanges[i] = -quantityChanges[i]
			}
			s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		// 回滚库存
		for i := range ticketIDs {
			quantityChanges[i] = -quantityChanges[i]
		}
		s.ticketDAO.BatchUpdateStock(ticketIDs, quantityChanges)
		return nil, err
	}

	return order, nil
}

// generateOrderNo 生成订单号
func (s *OrderService) generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("TS%s%06d", now.Format("20060102150405"), util.GenerateRandomInt(100000, 999999))
}

// GetOrderByID 根据ID获取订单
func (s *OrderService) GetOrderByID(orderID uint) (*model.Order, error) {
	return s.orderDAO.GetByID(orderID)
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	return s.orderDAO.GetByOrderNo(orderNo)
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(userID uint, status int8, offset, limit int) ([]*model.Order, int64, error) {
	return s.orderDAO.GetByUserID(userID, status, offset, limit)
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(orderID uint) error {
	order, err := s.orderDAO.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("订单状态不允许支付")
	}

	now := time.Now()
	return s.orderDAO.db.Model(&model.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":   model.OrderStatusPaid,
			"pay_time": now,
		}).Error
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderID uint) error {
	order, err := s.orderDAO.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("订单状态不允许取消")
	}

	// 使用事务
	tx := s.orderDAO.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新订单状态
	if err := tx.Model(&model.Order{}).Where("id = ?", orderID).
		Update("status", model.OrderStatusCancelled).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 回滚库存
	items, err := s.orderDAO.GetOrderItems(orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, item := range items {
		if err := tx.Model(&model.Ticket{}).Where("id = ?", item.TicketID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3. 回滚优惠券状态
	orderCoupons, err := s.orderDAO.GetOrderCoupons(orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, oc := range orderCoupons {
		if err := tx.Model(&model.UserCoupon{}).Where("id = ?", oc.UserCouponID).
			Update("status", model.UserCouponStatusUnused).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// AutoCancelOrder 自动取消超时订单
func (s *OrderService) AutoCancelOrder(timeoutHours int) error {
	orders, err := s.orderDAO.GetPendingOrders(timeoutHours)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if err := s.CancelOrder(order.ID); err != nil {
			util.Error("自动取消订单失败", util.WithField("order_id", order.ID), util.WithError(err))
		}
	}

	return nil
}

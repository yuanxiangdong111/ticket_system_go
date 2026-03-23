package service

import (
	"errors"
	"fmt"
	"ticket_system/internal/dao"
	"ticket_system/internal/model"
	"ticket_system/pkg/database"
	"ticket_system/pkg/redis"
	"ticket_system/pkg/util"
	"time"

	"gorm.io/gorm"
)

// SeckillService 秒杀服务
type SeckillService struct {
	orderDAO     *dao.OrderDAO
	ticketDAO    *dao.TicketDAO
	orderService *OrderService
}

// NewSeckillService 创建秒杀服务实例
func NewSeckillService() *SeckillService {
	return &SeckillService{
		orderDAO:     dao.NewOrderDAO(),
		ticketDAO:    dao.NewTicketDAO(),
		orderService: NewOrderService(),
	}
}

// CreateSeckillActivity 创建秒杀活动
func (s *SeckillService) CreateSeckillActivity(activity *model.SeckillActivity) error {
	// 检查门票是否存在
	ticket, err := s.ticketDAO.GetByID(activity.TicketID)
	if err != nil {
		return errors.New("门票不存在")
	}

	// 检查时间设置
	if activity.StartTime.After(activity.EndTime) {
		return errors.New("开始时间必须早于结束时间")
	}
	if activity.EndTime.Before(time.Now()) {
		return errors.New("结束时间必须晚于当前时间")
	}

	// 检查秒杀价格是否合理
	if activity.Price >= ticket.Price {
		return errors.New("秒杀价格必须低于原价")
	}

	// 创建秒杀活动
	tx := database.DB.Begin()
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建秒杀库存记录
	stock := &model.SeckillStock{
		ActivityID: activity.ID,
		TotalStock: activity.TotalStock,
		UsedStock:  0,
	}
	if err := tx.Create(stock).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	// 预热库存到Redis
	go s.preheatStock(activity.ID, activity.TotalStock)

	return nil
}

// preheatStock 预热库存到Redis
func (s *SeckillService) preheatStock(activityID uint, stock int) {
	stockKey := fmt.Sprintf("seckill:stock:%d", activityID)
	if err := redis.Set(stockKey, stock, 24*time.Hour); err != nil {
		util.Error("预热秒杀库存失败", util.WithField("activity_id", activityID), util.WithError(err))
	}
}

// GetSeckillActivityByID 根据ID获取秒杀活动
func (s *SeckillService) GetSeckillActivityByID(activityID uint) (*model.SeckillActivity, error) {
	var activity model.SeckillActivity
	err := database.DB.Preload("Ticket").First(&activity, activityID).Error
	return &activity, err
}

// ListSeckillActivities 获取秒杀活动列表
func (s *SeckillService) ListSeckillActivities(offset, limit int) ([]*model.SeckillActivity, int64, error) {
	var activities []*model.SeckillActivity
	var total int64

	query := database.DB.Preload("Ticket").Model(&model.SeckillActivity{})
	query.Count(&total).Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities)

	return activities, total, nil
}

// GetActiveSeckillActivities 获取进行中的秒杀活动
func (s *SeckillService) GetActiveSeckillActivities() ([]*model.SeckillActivity, error) {
	var activities []*model.SeckillActivity
	now := time.Now()
	err := database.DB.Preload("Ticket").
		Where("status = ? AND start_time <= ? AND end_time >= ?",
			model.SeckillActivityStatusActive, now, now).Find(&activities).Error
	return activities, err
}

// UpdateSeckillActivityStatus 更新秒杀活动状态
func (s *SeckillService) UpdateSeckillActivityStatus(activityID uint, status int8) error {
	return database.DB.Model(&model.SeckillActivity{}).Where("id = ?", activityID).Update("status", status).Error
}

// CreateSeckillOrder 创建秒杀订单
func (s *SeckillService) CreateSeckillOrder(userID uint, activityID uint, quantity int) (string, error) {
	// 1. 检查秒杀活动是否有效
	activity, err := s.GetSeckillActivityByID(activityID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	if activity.Status != model.SeckillActivityStatusActive {
		return "", errors.New("秒杀活动未开始或已结束")
	}
	if now.Before(activity.StartTime) {
		return "", errors.New("秒杀活动尚未开始")
	}
	if now.After(activity.EndTime) {
		return "", errors.New("秒杀活动已结束")
	}

	// 2. 检查用户是否已购买过该秒杀活动
	hasPurchased, err := s.checkUserHasPurchased(userID, activityID)
	if err != nil {
		return "", err
	}
	if hasPurchased {
		return "", errors.New("您已经参与过该秒杀活动")
	}

	// 3. 使用Redis预减库存
	stockKey := fmt.Sprintf("seckill:stock:%d", activityID)

	// 使用Lua脚本原子性地检查和减少库存
	luaScript := `
		local stock = redis.call('GET', KEYS[1])
		if stock and tonumber(stock) >= tonumber(ARGV[1]) then
			redis.call('DECRBY', KEYS[1], ARGV[1])
			return 1
		else
			return 0
		end
	`

	result, err := redis.Eval(luaScript, []string{stockKey}, quantity)
	if err != nil {
		return "", err
	}
	if result.(int64) == 0 {
		return "", errors.New("库存不足")
	}

	// 4. 用户购买标记（防止重复购买）
	userPurchaseKey := fmt.Sprintf("seckill:user:%d:%d", userID, activityID)
	hasSet, err := redis.SetNX(userPurchaseKey, 1, activity.EndTime.Sub(now))
	if err != nil {
		// 回滚库存
		redis.IncrBy(stockKey, int64(quantity))
		return "", err
	}
	if !hasSet {
		// 回滚库存
		redis.IncrBy(stockKey, int64(quantity))
		return "", errors.New("您已经参与过该秒杀活动")
	}

	// 5. 创建订单（异步处理）
	orderNo, err := s.createSeckillOrderAsync(userID, activity, quantity)
	if err != nil {
		// 回滚库存
		redis.IncrBy(stockKey, int64(quantity))
		redis.Del(userPurchaseKey)
		return "", err
	}

	// 6. 更新数据库库存（异步）
	go s.updateDatabaseStock(activityID, quantity)

	return orderNo, nil
}

// checkUserHasPurchased 检查用户是否已购买过该秒杀活动
func (s *SeckillService) checkUserHasPurchased(userID uint, activityID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&model.OrderCoupon{}).
		Joins("JOIN orders ON orders.id = order_coupons.order_id").
		Where("orders.user_id = ? AND order_coupons.coupon_id IN (SELECT id FROM coupons WHERE type = ? AND id = ?)",
			userID, model.CouponTypeSeckill, activityID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// createSeckillOrderAsync 异步创建秒杀订单
func (s *SeckillService) createSeckillOrderAsync(userID uint, activity *model.SeckillActivity, quantity int) (string, error) {
	// 先临时修改门票价格为秒杀价
	ticket, err := s.ticketDAO.GetByID(activity.TicketID)
	if err != nil {
		return "", err
	}
	originalPrice := ticket.Price
	ticket.Price = activity.Price

	// 创建订单（这里简化处理，实际应该更复杂）
	orderNo := fmt.Sprintf("SK%s%06d", time.Now().Format("20060102150405"), util.GenerateRandomInt(100000, 999999))
	totalAmount := originalPrice * float64(quantity)
	payAmount := activity.Price * float64(quantity)

	order := &model.Order{
		OrderNo:        orderNo,
		UserID:         userID,
		TotalAmount:    totalAmount,
		DiscountAmount: totalAmount - payAmount,
		PayAmount:      payAmount,
		Status:         model.OrderStatusPending,
	}

	// 创建订单
	if err := s.orderDAO.Create(order); err != nil {
		return "", err
	}

	// 创建订单详情
	orderItem := &model.OrderItem{
		OrderID:    order.ID,
		TicketID:   activity.TicketID,
		Quantity:   quantity,
		Price:      activity.Price,
		TotalPrice: payAmount,
	}
	if err := s.orderDAO.CreateOrderItem(orderItem); err != nil {
		return "", err
	}

	return orderNo, nil
}

// updateDatabaseStock 更新数据库中的秒杀库存
func (s *SeckillService) updateDatabaseStock(activityID uint, quantity int) {
	tx := database.DB.Begin()
	if err := tx.Model(&model.SeckillStock{}).Where("activity_id = ?", activityID).
		Update("used_stock", gorm.Expr("used_stock + ?", quantity)).Error; err != nil {
		tx.Rollback()
		util.Error("更新秒杀库存失败", util.WithField("activity_id", activityID), util.WithError(err))
		return
	}
	tx.Commit()
}

// GetSeckillStock 获取秒杀库存
func (s *SeckillService) GetSeckillStock(activityID uint) (int, error) {
	stockKey := fmt.Sprintf("seckill:stock:%d", activityID)
	stockStr, err := redis.Get(stockKey)
	if err != nil {
		// 从数据库查询
		var stock model.SeckillStock
		if err := database.DB.Where("activity_id = ?", activityID).First(&stock).Error; err != nil {
			return 0, err
		}
		availableStock := stock.TotalStock - stock.UsedStock
		// 预热到Redis
		redis.Set(stockKey, availableStock, 1*time.Hour)
		return availableStock, nil
	}
	return util.StringToInt(stockStr), nil
}

// SyncSeckillStock 同步秒杀库存（Redis -> 数据库）
func (s *SeckillService) SyncSeckillStock(activityID uint) error {
	stockKey := fmt.Sprintf("seckill:stock:%d", activityID)
	stockStr, err := redis.Get(stockKey)
	if err != nil {
		return err
	}

	redisStock := util.StringToInt(stockStr)

	var dbStock model.SeckillStock
	if err := database.DB.Where("activity_id = ?", activityID).First(&dbStock).Error; err != nil {
		return err
	}

	// 计算已使用的库存
	usedStock := dbStock.TotalStock - redisStock
	if usedStock > dbStock.UsedStock {
		return database.DB.Model(&model.SeckillStock{}).Where("activity_id = ?", activityID).
			Update("used_stock", usedStock).Error
	}

	return nil
}

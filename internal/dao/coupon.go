package dao

import (
	"errors"
	"ticket_system/internal/model"
	"ticket_system/pkg/database"
	"ticket_system/pkg/redis"
	"time"

	"gorm.io/gorm"
)

// CouponDAO 优惠券数据访问对象
type CouponDAO struct {
	db *gorm.DB
}

// NewCouponDAO 创建优惠券DAO实例
func NewCouponDAO() *CouponDAO {
	return &CouponDAO{db: database.DB}
}

// Create 创建优惠券
func (d *CouponDAO) Create(coupon *model.Coupon) error {
	return d.db.Create(coupon).Error
}

// GetByID 根据ID获取优惠券
func (d *CouponDAO) GetByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	err := d.db.First(&coupon, id).Error
	return &coupon, err
}

// List 获取优惠券列表
func (d *CouponDAO) List(offset, limit int, couponType int8, status int8) ([]*model.Coupon, int64, error) {
	var coupons []*model.Coupon
	var total int64

	query := d.db.Model(&model.Coupon{})
	if couponType > 0 {
		query = query.Where("type = ?", couponType)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total).Offset(offset).Limit(limit).Order("created_at DESC").Find(&coupons)
	return coupons, total, nil
}

// Update 更新优惠券
func (d *CouponDAO) Update(coupon *model.Coupon) error {
	return d.db.Save(coupon).Error
}

// UpdateUsedCount 更新已使用数量
func (d *CouponDAO) UpdateUsedCount(couponID uint, increment int) error {
	return d.db.Model(&model.Coupon{}).Where("id = ?", couponID).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", increment)).Error
}

// GetAvailableCoupons 获取可领取的优惠券
func (d *CouponDAO) GetAvailableCoupons() ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	now := time.Now()
	err := d.db.Where("status = ? AND start_time <= ? AND end_time >= ? AND used_count < total_count",
		model.CouponStatusAvailable, now, now).Find(&coupons).Error
	return coupons, err
}

// UserCouponDAO 用户优惠券数据访问对象
type UserCouponDAO struct {
	db *gorm.DB
}

// NewUserCouponDAO 创建用户优惠券DAO实例
func NewUserCouponDAO() *UserCouponDAO {
	return &UserCouponDAO{db: database.DB}
}

// Create 创建用户优惠券
func (d *UserCouponDAO) Create(userCoupon *model.UserCoupon) error {
	return d.db.Create(userCoupon).Error
}

// GetByID 根据ID获取用户优惠券
func (d *UserCouponDAO) GetByID(id uint) (*model.UserCoupon, error) {
	var userCoupon model.UserCoupon
	err := d.db.Preload("Coupon").First(&userCoupon, id).Error
	return &userCoupon, err
}

// GetByUserID 根据用户ID获取用户优惠券列表
func (d *UserCouponDAO) GetByUserID(userID uint, status int8) ([]*model.UserCoupon, error) {
	var userCoupons []*model.UserCoupon
	query := d.db.Preload("Coupon").Where("user_id = ?", userID)
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&userCoupons).Error
	return userCoupons, err
}

// CheckUserHasCoupon 检查用户是否已领取该优惠券
func (d *UserCouponDAO) CheckUserHasCoupon(userID, couponID uint) (bool, error) {
	var count int64
	err := d.db.Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count > 0, err
}

// UpdateStatus 更新用户优惠券状态
func (d *UserCouponDAO) UpdateStatus(id uint, status int8) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == model.UserCouponStatusUsed {
		now := time.Now()
		updates["used_time"] = &now
	}
	return d.db.Model(&model.UserCoupon{}).Where("id = ?", id).Updates(updates).Error
}

// GetAvailableUserCoupons 获取用户可用优惠券
func (d *UserCouponDAO) GetAvailableUserCoupons(userID uint) ([]*model.UserCoupon, error) {
	var userCoupons []*model.UserCoupon
	now := time.Now()
	err := d.db.Preload("Coupon").
		Where("user_id = ? AND status = ?", userID, model.UserCouponStatusUnused).
		Find(&userCoupons).Error
	if err != nil {
		return nil, err
	}

	var result []*model.UserCoupon
	for _, uc := range userCoupons {
		if now.After(uc.Coupon.StartTime) && now.Before(uc.Coupon.EndTime) {
			result = append(result, uc)
		}
	}
	return result, nil
}

// UpdateExpiredCoupons 更新过期优惠券状态
func (d *UserCouponDAO) UpdateExpiredCoupons() error {
	now := time.Now()
	return d.db.Model(&model.UserCoupon{}).
		Where("status = ? AND id IN (SELECT uc.id FROM user_coupons uc JOIN coupons c ON uc.coupon_id = c.id WHERE c.end_time < ?)",
			model.UserCouponStatusUnused, now).
		Update("status", model.UserCouponStatusExpired).Error
}

// LockCoupon 锁定优惠券（用于订单创建）
func (d *UserCouponDAO) LockCoupon(userID, userCouponID uint) (bool, error) {
	lockKey := "coupon:lock:" + string(rune(userCouponID))
	locked, err := redis.SetNX(lockKey, 1, 30*time.Second)
	if err != nil {
		return false, err
	}
	if !locked {
		return false, nil
	}

	var userCoupon model.UserCoupon
	err = d.db.Where("id = ? AND user_id = ? AND status = ?", userCouponID, userID, model.UserCouponStatusUnused).First(&userCoupon).Error
	if err != nil {
		redis.Del(lockKey)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UnlockCoupon 解锁优惠券
func (d *UserCouponDAO) UnlockCoupon(userCouponID uint) error {
	lockKey := "coupon:lock:" + string(rune(userCouponID))
	return redis.Del(lockKey)
}

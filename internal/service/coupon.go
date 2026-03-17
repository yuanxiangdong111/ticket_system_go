package service

import (
	"errors"
	"math"
	"ticket_system/internal/dao"
	"ticket_system/internal/model"
	"ticket_system/pkg/util"
	"time"
)

// CouponService 优惠券服务
type CouponService struct {
	couponDAO        *dao.CouponDAO
	userCouponDAO    *dao.UserCouponDAO
	ticketDAO        *dao.TicketDAO
}

// NewCouponService 创建优惠券服务实例
func NewCouponService() *CouponService {
	return &CouponService{
		couponDAO:        dao.NewCouponDAO(),
		userCouponDAO:    dao.NewUserCouponDAO(),
		ticketDAO:        dao.NewTicketDAO(),
	}
}

// CreateCoupon 创建优惠券
func (s *CouponService) CreateCoupon(coupon *model.Coupon) error {
	if coupon.Type == model.CouponTypeCash && coupon.MinAmount <= 0 {
		return errors.New("满减券需要设置最低消费金额")
	}
	if coupon.Type == model.CouponTypeDiscount && (coupon.Discount <= 0 || coupon.Discount >= 1) {
		return errors.New("折扣券折扣必须在0-1之间")
	}
	if coupon.StartTime.After(coupon.EndTime) {
		return errors.New("开始时间必须早于结束时间")
	}

	return s.couponDAO.Create(coupon)
}

// GetCouponByID 根据ID获取优惠券
func (s *CouponService) GetCouponByID(couponID uint) (*model.Coupon, error) {
	return s.couponDAO.GetByID(couponID)
}

// ListCoupons 获取优惠券列表
func (s *CouponService) ListCoupons(offset, limit int, couponType int8, status int8) ([]*model.Coupon, int64, error) {
	return s.couponDAO.List(offset, limit, couponType, status)
}

// GetAvailableCoupons 获取可领取的优惠券
func (s *CouponService) GetAvailableCoupons() ([]*model.Coupon, error) {
	return s.couponDAO.GetAvailableCoupons()
}

// ReceiveCoupon 用户领取优惠券
func (s *CouponService) ReceiveCoupon(userID uint, couponID uint) error {
	// 检查用户是否已领取过该优惠券
	hasCoupon, err := s.checkUserHasCoupon(userID, couponID)
	if err != nil {
		return err
	}
	if hasCoupon {
		return errors.New("您已经领取过该优惠券")
	}

	// 检查优惠券是否可领取
	coupon, err := s.couponDAO.GetByID(couponID)
	if err != nil {
		return err
	}

	now := time.Now()
	if coupon.Status != model.CouponStatusAvailable {
		return errors.New("优惠券不可用")
	}
	if now.Before(coupon.StartTime) || now.After(coupon.EndTime) {
		return errors.New("优惠券不在领取时间范围内")
	}
	if coupon.UsedCount >= coupon.TotalCount {
		return errors.New("优惠券已被领完")
	}

	// 创建用户优惠券
	userCoupon := &model.UserCoupon{
		UserID:   userID,
		CouponID: couponID,
		Status:   model.UserCouponStatusUnused,
	}

	if err := s.userCouponDAO.Create(userCoupon); err != nil {
		return err
	}

	// 更新优惠券已使用数量
	if err := s.couponDAO.UpdateUsedCount(couponID, 1); err != nil {
		return err
	}

	return nil
}

// CheckUserHasCoupon 检查用户是否已领取过优惠券
func (s *CouponService) checkUserHasCoupon(userID, couponID uint) (bool, error) {
	var count int64
	err := s.userCouponDAO.db.Model(&model.UserCoupon{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CalculateFinalPrice 计算最终价格（优惠券叠加使用）
func (s *CouponService) CalculateFinalPrice(originalPrice float64, coupons []*model.UserCoupon) (float64, float64, error) {
	var finalPrice = originalPrice
	var discountAmount = 0.0
	var hasSeckillCoupon = false

	// 分离优惠券类型
	var discountCoupons []*model.Coupon // 折扣券
	var cashCoupons []*model.Coupon    // 满减券

	for _, userCoupon := range coupons {
		coupon := &userCoupon.Coupon
		if coupon.Type == model.CouponTypeSeckill {
			if hasSeckillCoupon {
				return 0, 0, errors.New("只能使用一张秒杀券")
			}
			hasSeckillCoupon = true
			finalPrice = coupon.Discount
			discountAmount = originalPrice - finalPrice
		} else if coupon.Type == model.CouponTypeDiscount {
			discountCoupons = append(discountCoupons, coupon)
		} else if coupon.Type == model.CouponTypeCash {
			cashCoupons = append(cashCoupons, coupon)
		}
	}

	// 如果使用了秒杀券，不能再使用其他优惠券
	if hasSeckillCoupon {
		return finalPrice, discountAmount, nil
	}

	// 计算折扣券优惠
	if len(discountCoupons) > 0 {
		// 选择最优折扣券（最大折扣）
		bestDiscount := 1.0
		for _, coupon := range discountCoupons {
			if coupon.Discount < bestDiscount {
				bestDiscount = coupon.Discount
			}
		}
		// 应用折扣
		discountAmount = originalPrice * (1 - bestDiscount)
		finalPrice = originalPrice * bestDiscount
	}

	// 计算满减券优惠
	if len(cashCoupons) > 0 {
		var totalCashDiscount = 0.0
		for _, coupon := range cashCoupons {
			// 检查满减条件
			if originalPrice >= coupon.MinAmount {
				totalCashDiscount += coupon.Discount
			}
		}
		// 应用满减
		finalPrice -= totalCashDiscount
		discountAmount += totalCashDiscount
	}

	// 确保价格不低于0
	if finalPrice < 0 {
		finalPrice = 0
	}

	return finalPrice, discountAmount, nil
}

// CalculateCouponDiscount 计算单个优惠券的折扣
func (s *CouponService) CalculateCouponDiscount(originalPrice float64, coupon *model.Coupon) (float64, error) {
	var discount = 0.0

	switch coupon.Type {
	case model.CouponTypeCash:
		if originalPrice >= coupon.MinAmount {
			discount = coupon.Discount
		}
	case model.CouponTypeDiscount:
		discount = originalPrice * (1 - coupon.Discount)
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
	case model.CouponTypeSeckill:
		discount = originalPrice - coupon.Discount
	}

	return discount, nil
}

// GetUserCoupons 获取用户优惠券列表
func (s *CouponService) GetUserCoupons(userID uint, status int8) ([]*model.UserCoupon, error) {
	coupons, err := s.userCouponDAO.GetByUserID(userID, status)
	if err != nil {
		return nil, err
	}

	// 过滤已过期但状态未更新的优惠券
	var validCoupons []*model.UserCoupon
	now := time.Now()
	for _, uc := range coupons {
		if uc.Status == model.UserCouponStatusUnused && now.After(uc.Coupon.EndTime) {
			uc.Status = model.UserCouponStatusExpired
			go s.userCouponDAO.UpdateStatus(uc.ID, model.UserCouponStatusExpired)
		}
		validCoupons = append(validCoupons, uc)
	}

	return validCoupons, nil
}

// UpdateExpiredCoupons 更新过期优惠券状态
func (s *CouponService) UpdateExpiredCoupons() error {
	return s.userCouponDAO.UpdateExpiredCoupons()
}

// CheckCouponsAvailable 检查优惠券是否可用
func (s *CouponService) CheckCouponsAvailable(userID uint, userCouponIDs []uint) ([]*model.UserCoupon, error) {
	var validCoupons []*model.UserCoupon

	for _, userCouponID := range userCouponIDs {
		userCoupon, err := s.userCouponDAO.GetByID(userCouponID)
		if err != nil {
			continue
		}

		if userCoupon.UserID != userID {
			continue
		}

		if userCoupon.Status != model.UserCouponStatusUnused {
			continue
		}

		now := time.Now()
		if now.Before(userCoupon.Coupon.StartTime) || now.After(userCoupon.Coupon.EndTime) {
			continue
		}

		validCoupons = append(validCoupons, userCoupon)
	}

	return validCoupons, nil
}

// FilterAvailableCoupons 过滤可用于指定价格的优惠券
func (s *CouponService) FilterAvailableCoupons(userID uint, price float64) ([]*model.UserCoupon, error) {
	availableCoupons, err := s.userCouponDAO.GetAvailableUserCoupons(userID)
	if err != nil {
		return nil, err
	}

	var validCoupons []*model.UserCoupon
	for _, uc := range availableCoupons {
		if uc.Coupon.Type == model.CouponTypeCash {
			if price >= uc.Coupon.MinAmount {
				validCoupons = append(validCoupons, uc)
			}
		} else if uc.Coupon.Type == model.CouponTypeDiscount {
			if uc.Coupon.MinAmount <= 0 || price >= uc.Coupon.MinAmount {
				validCoupons = append(validCoupons, uc)
			}
		} else if uc.Coupon.Type == model.CouponTypeSeckill {
			validCoupons = append(validCoupons, uc)
		}
	}

	return validCoupons, nil
}

// CalculateOptimalCouponCombination 计算最优优惠券组合
func (s *CouponService) CalculateOptimalCouponCombination(originalPrice float64, coupons []*model.UserCoupon) ([]*model.UserCoupon, float64, float64, error) {
	var bestCombination []*model.UserCoupon
	var bestFinalPrice = originalPrice
	var bestDiscountAmount = 0.0

	// 尝试所有可能的优惠券组合
	for mask := 1; mask < (1 << uint(len(coupons))); mask++ {
		var currentCombination []*model.UserCoupon
		var hasSeckill = false

		for i := 0; i < len(coupons); i++ {
			if mask&(1<<uint(i)) != 0 {
				coupon := coupons[i]
				if coupon.Coupon.Type == model.CouponTypeSeckill {
					if hasSeckill {
						// 已经有秒杀券，不能再添加
						currentCombination = nil
						break
					}
					hasSeckill = true
				}
				currentCombination = append(currentCombination, coupon)
			}
		}

		if currentCombination == nil {
			continue
		}

		// 计算当前组合的价格
		price, discount, err := s.CalculateFinalPrice(originalPrice, currentCombination)
		if err != nil {
			continue
		}

		// 更新最优组合
		if price < bestFinalPrice {
			bestFinalPrice = price
			bestDiscountAmount = discount
			bestCombination = currentCombination
		} else if price == bestFinalPrice && len(currentCombination) < len(bestCombination) {
			// 价格相同情况下，使用优惠券数量更少的组合
			bestCombination = currentCombination
		}
	}

	// 如果没有找到更好的组合，使用原价
	if bestCombination == nil {
		bestCombination = []*model.UserCoupon{}
	}

	return bestCombination, bestFinalPrice, bestDiscountAmount, nil
}

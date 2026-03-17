package controller

import (
	"net/http"
	"strconv"
	"ticket_system/internal/model"
	"ticket_system/internal/service"
	"ticket_system/pkg/util"

	"github.com/gin-gonic/gin"
)

// CouponController 优惠券控制器
type CouponController struct {
	couponService *service.CouponService
}

// NewCouponController 创建优惠券控制器实例
func NewCouponController() *CouponController {
	return &CouponController{
		couponService: service.NewCouponService(),
	}
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Name        string  `json:"name" binding:"required"`
	Type        int8    `json:"type" binding:"required,oneof=1 2 3"`
	Discount    float64 `json:"discount" binding:"required,gt=0"`
	MinAmount   float64 `json:"min_amount"`
	MaxDiscount float64 `json:"max_discount"`
	TotalCount  int     `json:"total_count" binding:"required,gt=0"`
	StartTime   string  `json:"start_time" binding:"required"`
	EndTime     string  `json:"end_time" binding:"required"`
}

// CreateCoupon 创建优惠券（管理员）
func (c *CouponController) CreateCoupon(ctx *gin.Context) {
	var req CreateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 解析时间
	startTime, err := util.ParseTime(req.StartTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := util.ParseTime(req.EndTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	coupon := &model.Coupon{
		Name:        req.Name,
		Type:        req.Type,
		Discount:    req.Discount,
		MinAmount:   req.MinAmount,
		MaxDiscount: req.MaxDiscount,
		TotalCount:  req.TotalCount,
		UsedCount:   0,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      model.CouponStatusAvailable,
	}

	if err := c.couponService.CreateCoupon(coupon); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建优惠券失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    coupon,
	})
}

// GetCoupon 获取优惠券详情
func (c *CouponController) GetCoupon(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	coupon, err := c.couponService.GetCouponByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "优惠券不存在",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    coupon,
	})
}

// ListCoupons 获取优惠券列表
func (c *CouponController) ListCoupons(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	couponType, _ := strconv.Atoi(ctx.Query("type"))
	status, _ := strconv.Atoi(ctx.Query("status"))

	offset := (page - 1) * pageSize

	coupons, total, err := c.couponService.ListCoupons(offset, pageSize, int8(couponType), int8(status))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取优惠券列表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  coupons,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}

// GetAvailableCoupons 获取可领取的优惠券列表
func (c *CouponController) GetAvailableCoupons(ctx *gin.Context) {
	coupons, err := c.couponService.GetAvailableCoupons()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取优惠券失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    coupons,
	})
}

// ReceiveCouponRequest 领取优惠券请求
type ReceiveCouponRequest struct {
	CouponID uint `json:"coupon_id" binding:"required"`
}

// ReceiveCoupon 用户领取优惠券
func (c *CouponController) ReceiveCoupon(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var req ReceiveCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := c.couponService.ReceiveCoupon(userID.(uint), req.CouponID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "领取成功",
	})
}

// GetUserCoupons 获取用户优惠券列表
func (c *CouponController) GetUserCoupons(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	status, _ := strconv.Atoi(ctx.Query("status"))

	coupons, err := c.couponService.GetUserCoupons(userID.(uint), int8(status))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户优惠券失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    coupons,
	})
}

// CalculatePriceRequest 计算价格请求
type CalculatePriceRequest struct {
	TotalAmount   float64 `json:"total_amount" binding:"required,gt=0"`
	UserCouponIDs []uint  `json:"user_coupon_ids"`
}

// CalculatePrice 计算最终价格
func (c *CouponController) CalculatePrice(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var req CalculatePriceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	validCoupons, err := c.couponService.CheckCouponsAvailable(userID.(uint), req.UserCouponIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查优惠券失败",
		})
		return
	}

	finalPrice, discountAmount, err := c.couponService.CalculateFinalPrice(req.TotalAmount, validCoupons)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "计算成功",
		"data": gin.H{
			"total_amount":    req.TotalAmount,
			"discount_amount": discountAmount,
			"final_price":     finalPrice,
		},
	})
}

// GetOptimalCoupons 获取最优优惠券组合
func (c *CouponController) GetOptimalCoupons(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	totalAmount, _ := strconv.ParseFloat(ctx.Query("total_amount"), 64)
	if totalAmount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请输入有效的金额",
		})
		return
	}

	// 获取用户可用优惠券
	availableCoupons, err := c.couponService.FilterAvailableCoupons(userID.(uint), totalAmount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取优惠券失败",
		})
		return
	}

	// 计算最优组合
	bestCoupons, bestFinalPrice, bestDiscountAmount, err := c.couponService.CalculateOptimalCouponCombination(totalAmount, availableCoupons)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "计算最优组合失败",
		})
		return
	}

	var couponIDs []uint
	for _, coupon := range bestCoupons {
		couponIDs = append(couponIDs, coupon.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"total_amount":     totalAmount,
			"discount_amount":  bestDiscountAmount,
			"final_price":      bestFinalPrice,
			"user_coupon_ids":  couponIDs,
			"coupons":          bestCoupons,
		},
	})
}

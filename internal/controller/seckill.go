package controller

import (
	"net/http"
	"strconv"
	"ticket_system/internal/model"
	"ticket_system/internal/service"
	"ticket_system/pkg/util"

	"github.com/gin-gonic/gin"
)

// SeckillController 秒杀控制器
type SeckillController struct {
	seckillService *service.SeckillService
}

// NewSeckillController 创建秒杀控制器实例
func NewSeckillController() *SeckillController {
	return &SeckillController{
		seckillService: service.NewSeckillService(),
	}
}

// CreateSeckillActivityRequest 创建秒杀活动请求
type CreateSeckillActivityRequest struct {
	Name          string  `json:"name" binding:"required"`
	TicketID      uint    `json:"ticket_id" binding:"required"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	TotalStock    int     `json:"total_stock" binding:"required,gt=0"`
	StartTime     string  `json:"start_time" binding:"required"`
	EndTime       string  `json:"end_time" binding:"required"`
	Description   string  `json:"description"`
}

// CreateSeckillActivity 创建秒杀活动（管理员）
func (c *SeckillController) CreateSeckillActivity(ctx *gin.Context) {
	var req CreateSeckillActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

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

	activity := &model.SeckillActivity{
		Name:            req.Name,
		TicketID:        req.TicketID,
		Price:           req.Price,
		TotalStock:      req.TotalStock,
		AvailableStock:  req.TotalStock,
		StartTime:       startTime,
		EndTime:         endTime,
		Status:          model.SeckillActivityStatusInactive,
		Description:     req.Description,
	}

	if err := c.seckillService.CreateSeckillActivity(activity); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建秒杀活动失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    activity,
	})
}

// GetSeckillActivity 获取秒杀活动详情
func (c *SeckillController) GetSeckillActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	activity, err := c.seckillService.GetSeckillActivityByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "秒杀活动不存在",
		})
		return
	}

	// 获取当前库存
	stock, _ := c.seckillService.GetSeckillStock(uint(id))
	activity.AvailableStock = stock

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    activity,
	})
}

// ListSeckillActivities 获取秒杀活动列表
func (c *SeckillController) ListSeckillActivities(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	offset := (page - 1) * pageSize

	activities, total, err := c.seckillService.ListSeckillActivities(offset, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取秒杀活动列表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":       activities,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

// GetActiveSeckillActivities 获取进行中的秒杀活动
func (c *SeckillController) GetActiveSeckillActivities(ctx *gin.Context) {
	activities, err := c.seckillService.GetActiveSeckillActivities()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取秒杀活动失败",
		})
		return
	}

	// 获取每个活动的实时库存
	for _, activity := range activities {
		stock, _ := c.seckillService.GetSeckillStock(activity.ID)
		activity.AvailableStock = stock
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    activities,
	})
}

// UpdateSeckillActivityStatus 更新秒杀活动状态（管理员）
func (c *SeckillController) UpdateSeckillActivityStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	var req struct {
		Status int8 `json:"status" binding:"required,oneof=0 1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := c.seckillService.UpdateSeckillActivityStatus(uint(id), req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// CreateSeckillOrderRequest 创建秒杀订单请求
type CreateSeckillOrderRequest struct {
	ActivityID uint `json:"activity_id" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,min=1,max=10"`
}

// CreateSeckillOrder 创建秒杀订单
func (c *SeckillController) CreateSeckillOrder(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var req CreateSeckillOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	orderNo, err := c.seckillService.CreateSeckillOrder(userID.(uint), req.ActivityID, req.Quantity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "秒杀成功",
		"data": gin.H{
			"order_no": orderNo,
		},
	})
}

// GetSeckillStock 获取秒杀库存
func (c *SeckillController) GetSeckillStock(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	stock, err := c.seckillService.GetSeckillStock(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取库存失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"stock": stock,
		},
	})
}

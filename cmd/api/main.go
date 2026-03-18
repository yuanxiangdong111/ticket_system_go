package main

import (
	"fmt"
	"ticket_system/internal/controller"
	"ticket_system/internal/model"
	"ticket_system/pkg/database"
	"ticket_system/pkg/redis"
	"ticket_system/pkg/util"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("            门票系统启动中...")
	fmt.Println("==================================================")
	fmt.Println()

	// 初始化日志
	logConfig := &util.LogConfig{
		Level:      "info",
		Filename:   "logs/ticket_system.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
	}
	if err := util.InitLogger(logConfig); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		return
	}
	defer util.Sync()
	fmt.Println("[1] 日志初始化完成")

	// 初始化数据库
	fmt.Println("[2] 正在连接 MySQL...")
	dbConfig := &database.Config{
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "123456",
		DBName:       "ticket_system",
		Charset:      "utf8mb4",
		MaxOpenConns: 100,
		MaxIdleConns: 10,
	}
	fmt.Printf("    配置: host=%s port=%d user=%s db=%s\n",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DBName)

	if err := database.Init(dbConfig); err != nil {
		util.Fatal("数据库连接失败", util.WithError(err))
		return
	}
	defer database.Close()
	fmt.Println("[3] MySQL 连接成功")

	fmt.Println("[4] 开始数据库迁移...")
	// 自动迁移和初始化数据
	if err := model.Migrate(); err != nil {
		util.Fatal("数据库迁移失败", util.WithError(err))
		return
	}
	fmt.Println("[5] 数据库迁移完成")

	fmt.Println("[6] 开始数据初始化...")
	if err := model.InitData(); err != nil {
		util.Fatal("数据初始化失败", util.WithError(err))
		return
	}
	fmt.Println("[7] 数据初始化完成")

	fmt.Println("[8] 正在连接 Redis...")
	// 初始化Redis
	redisConfig := &redis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 100,
	}
	if err := redis.Init(redisConfig); err != nil {
		util.Fatal("Redis连接失败", util.WithError(err))
		return
	}
	defer redis.Close()
	fmt.Println("[9] Redis 连接成功")

	// 设置Gin模式
	fmt.Println("[10] 启动 API 服务...")
	gin.SetMode("debug")

	// 创建Gin引擎
	r := gin.Default()

	// 注册路由
	registerRoutes(r)

	// 启动服务
	port := 8080
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Printf("    服务启动成功！端口: %d\n", port)
	fmt.Println("==================================================")
	fmt.Println()
	fmt.Println("健康检查: http://localhost:8080/health")
	fmt.Println()

	util.Info("服务启动成功", util.WithField("port", port))
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		util.Fatal("服务启动失败", util.WithError(err))
	}
}

// registerRoutes 注册路由
func registerRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "服务正常",
		})
	})

	// API v1路由组
	apiV1 := r.Group("/api/v1")
	{
		// 优惠券路由
		couponController := controller.NewCouponController()
		coupons := apiV1.Group("/coupons")
		{
			coupons.GET("", couponController.ListCoupons)
			coupons.GET("/available", couponController.GetAvailableCoupons)
			coupons.GET("/:id", couponController.GetCoupon)
			coupons.POST("", couponController.CreateCoupon)
			coupons.POST("/receive", couponController.ReceiveCoupon)
			coupons.POST("/calculate", couponController.CalculatePrice)
			coupons.GET("/optimal/recommendation", couponController.GetOptimalCoupons)
		}

		// 用户优惠券路由
		userCoupons := apiV1.Group("/user/coupons")
		{
			userCoupons.GET("", couponController.GetUserCoupons)
		}

		// 订单路由
		orderController := controller.NewOrderController()
		orders := apiV1.Group("/orders")
		{
			orders.POST("", orderController.CreateOrder)
			orders.GET("", orderController.GetOrders)
			orders.GET("/:id", orderController.GetOrder)
			orders.PUT("/:id/pay", orderController.PayOrder)
			orders.PUT("/:id/cancel", orderController.CancelOrder)
		}

		// 秒杀路由
		seckillController := controller.NewSeckillController()
		seckills := apiV1.Group("/seckill")
		{
			seckills.GET("/activities", seckillController.ListSeckillActivities)
			seckills.GET("/activities/active", seckillController.GetActiveSeckillActivities)
			seckills.GET("/activities/:id", seckillController.GetSeckillActivity)
			seckills.GET("/activities/:id/stock", seckillController.GetSeckillStock)
			seckills.POST("/activities", seckillController.CreateSeckillActivity)
			seckills.PUT("/activities/:id/status", seckillController.UpdateSeckillActivityStatus)
			seckills.POST("/order", seckillController.CreateSeckillOrder)
		}
	}
}

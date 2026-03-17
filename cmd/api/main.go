package main

import (
	"fmt"
	"ticket_system/config"
	"ticket_system/internal/controller"
	"ticket_system/internal/model"
	"ticket_system/pkg/database"
	"ticket_system/pkg/redis"
	"ticket_system/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 初始化日志
	logConfig := &util.LogConfig{
		Level:      viper.GetString("log.level"),
		Filename:   viper.GetString("log.filename"),
		MaxSize:    viper.GetInt("log.max_size"),
		MaxBackups: viper.GetInt("log.max_backups"),
		MaxAge:     viper.GetInt("log.max_age"),
	}
	if err := util.InitLogger(logConfig); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		return
	}
	defer util.Sync()

	util.Info("门票系统启动中...")

	// 初始化数据库
	dbConfig := &database.Config{
		Host:         viper.GetString("database.host"),
		Port:         viper.GetInt("database.port"),
		User:         viper.GetString("database.user"),
		Password:     viper.GetString("database.password"),
		DBName:       viper.GetString("database.dbname"),
		Charset:      viper.GetString("database.charset"),
		MaxOpenConns: viper.GetInt("database.max_open_conns"),
		MaxIdleConns: viper.GetInt("database.max_idle_conns"),
	}
	if err := database.Init(dbConfig); err != nil {
		util.Fatal("数据库连接失败", util.WithError(err))
		return
	}
	defer database.Close()

	// 自动迁移和初始化数据
	if err := model.Migrate(); err != nil {
		util.Fatal("数据库迁移失败", util.WithError(err))
		return
	}

	if err := model.InitData(); err != nil {
		util.Fatal("数据初始化失败", util.WithError(err))
		return
	}

	// 初始化Redis
	redisConfig := &redis.Config{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	}
	if err := redis.Init(redisConfig); err != nil {
		util.Fatal("Redis连接失败", util.WithError(err))
		return
	}
	defer redis.Close()

	// 设置Gin模式
	gin.SetMode(viper.GetString("server.mode"))

	// 创建Gin引擎
	r := gin.Default()

	// 注册路由
	registerRoutes(r)

	// 启动服务
	port := viper.GetInt("server.port")
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

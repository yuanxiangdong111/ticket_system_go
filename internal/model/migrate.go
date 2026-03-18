package model

import (
	"ticket_system/pkg/database"
	"ticket_system/pkg/util"
)

// Migrate 自动迁移数据库表
func Migrate() error {
	util.Info("开始数据库迁移...")

	// 自动迁移表结构
	err := database.DB.AutoMigrate(
		&User{},
		&Category{},
		&Ticket{},
		&Coupon{},
		&UserCoupon{},
		&Order{},
		&OrderItem{},
		&OrderCoupon{},
		&SeckillActivity{},
		&SeckillStock{},
	)

	if err != nil {
		return err
	}

	util.Info("数据库迁移完成")
	return nil
}

// InitData 初始化数据
func InitData() error {
	util.Info("开始初始化数据...")

	// 检查是否已有数据
	var userCount int64
	database.DB.Model(&User{}).Count(&userCount)
	if userCount > 0 {
		util.Info("数据已存在，跳过初始化")
		return nil
	}

	// 初始化管理员用户
	admin := &User{
		Username: "admin",
		Password: "e10adc3949ba59abbe56e057f20f883e", // 123456 的 MD5
		Email:    "admin@example.com",
		Phone:    "13800138000",
		Nickname: "管理员",
		Status:   1,
	}
	if err := database.DB.Create(admin).Error; err != nil {
		return err
	}

	// 初始化分类
	categories := []*Category{
		{Name: "演唱会", ParentID: 0, Sort: 1, Status: 1},
		{Name: "体育赛事", ParentID: 0, Sort: 2, Status: 1},
		{Name: "话剧歌剧", ParentID: 0, Sort: 3, Status: 1},
		{Name: "展览展会", ParentID: 0, Sort: 4, Status: 1},
		{Name: "旅游景点", ParentID: 0, Sort: 5, Status: 1},
	}
	if err := database.DB.CreateInBatches(categories, 10).Error; err != nil {
		return err
	}

	// 初始化门票
	tickets := []*Ticket{
		{
			Title:         "周杰伦演唱会门票",
			Description:   "周杰伦2024年世界巡回演唱会",
			CategoryID:    1,
			Price:         880.00,
			OriginalPrice: 1280.00,
			Stock:         1000,
			Sold:          0,
			Status:        1,
			Sort:          1,
		},
		{
			Title:         "中超联赛门票",
			Description:   "中超联赛广州队vs上海队",
			CategoryID:    2,
			Price:         180.00,
			OriginalPrice: 280.00,
			Stock:         5000,
			Sold:          0,
			Status:        1,
			Sort:          2,
		},
		{
			Title:         "话剧《雷雨》",
			Description:   "经典话剧演出",
			CategoryID:    3,
			Price:         120.00,
			OriginalPrice: 180.00,
			Stock:         200,
			Sold:          0,
			Status:        1,
			Sort:          3,
		},
	}
	if err := database.DB.CreateInBatches(tickets, 10).Error; err != nil {
		return err
	}

	// 初始化优惠券
	now := database.DB.NowFunc()
	coupons := []*Coupon{
		{
			Name:       "满100减20",
			Type:       CouponTypeCash,
			Discount:   20.00,
			MinAmount:  100.00,
			TotalCount: 1000,
			UsedCount:  0,
			StartTime:  now,
			EndTime:    now.AddDate(0, 1, 0), // 30天后
			Status:     CouponStatusAvailable,
		},
		{
			Name:        "8折优惠券",
			Type:        CouponTypeDiscount,
			Discount:    0.80,
			MaxDiscount: 50.00,
			TotalCount:  500,
			UsedCount:   0,
			StartTime:   now,
			EndTime:     now.AddDate(0, 0, 15), // 15天后
			Status:      CouponStatusAvailable,
		},
		{
			Name:       "秒杀券 - 50元",
			Type:       CouponTypeSeckill,
			Discount:   50.00,
			TotalCount: 100,
			UsedCount:  0,
			StartTime:  now,
			EndTime:    now.Add(time.Hour * 1), // 1小时后
			Status:     CouponStatusAvailable,
		},
	}
	if err := database.DB.CreateInBatches(coupons, 10).Error; err != nil {
		return err
	}

	util.Info("数据初始化完成")
	return nil
}

-- 门票系统数据库初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS ticket_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE ticket_system;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `nickname` varchar(50) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-正常 0-禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 分类表
CREATE TABLE IF NOT EXISTS `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '分类名称',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父分类ID',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-启用 0-禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分类表';

-- 门票表
CREATE TABLE IF NOT EXISTS `tickets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(200) NOT NULL COMMENT '标题',
  `description` text COMMENT '描述',
  `image` varchar(255) DEFAULT NULL COMMENT '图片',
  `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
  `price` decimal(10,2) NOT NULL COMMENT '价格',
  `original_price` decimal(10,2) DEFAULT NULL COMMENT '原价',
  `stock` int NOT NULL DEFAULT '0' COMMENT '库存',
  `sold` int NOT NULL DEFAULT '0' COMMENT '销量',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-上架 0-下架',
  `start_time` datetime DEFAULT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `location` varchar(255) DEFAULT NULL COMMENT '地点',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门票表';

-- 优惠券表
CREATE TABLE IF NOT EXISTS `coupons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '优惠券名称',
  `type` tinyint NOT NULL COMMENT '优惠券类型: 1-满减券 2-折扣券 3-秒杀券',
  `discount` decimal(10,2) NOT NULL COMMENT '折扣金额或折扣比例',
  `min_amount` decimal(10,2) DEFAULT NULL COMMENT '满减券最低消费金额',
  `max_discount` decimal(10,2) DEFAULT NULL COMMENT '折扣券最大折扣金额',
  `total_count` int NOT NULL COMMENT '总数量',
  `used_count` int NOT NULL DEFAULT '0' COMMENT '已使用数量',
  `start_time` datetime NOT NULL COMMENT '有效期开始时间',
  `end_time` datetime NOT NULL COMMENT '有效期结束时间',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-可用 0-不可用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type_status` (`type`,`status`),
  KEY `idx_start_time` (`start_time`),
  KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券表';

-- 用户优惠券关联表
CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `coupon_id` bigint unsigned NOT NULL COMMENT '优惠券ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-未使用 2-已使用 3-已过期',
  `obtain_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  `used_time` datetime DEFAULT NULL COMMENT '使用时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id_status` (`user_id`,`status`),
  KEY `idx_coupon_id` (`coupon_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券关联表';

-- 订单表
CREATE TABLE IF NOT EXISTS `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no` varchar(32) NOT NULL COMMENT '订单号',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `total_amount` decimal(10,2) NOT NULL COMMENT '总金额',
  `discount_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '优惠金额',
  `pay_amount` decimal(10,2) NOT NULL COMMENT '实际支付金额',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-待支付 2-已支付 3-已取消 4-已退款',
  `pay_time` datetime DEFAULT NULL COMMENT '支付时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id_status` (`user_id`,`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

-- 订单详情表
CREATE TABLE IF NOT EXISTS `order_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `ticket_id` bigint unsigned NOT NULL COMMENT '门票ID',
  `quantity` int NOT NULL DEFAULT '1' COMMENT '数量',
  `price` decimal(10,2) NOT NULL COMMENT '单价',
  `total_price` decimal(10,2) NOT NULL COMMENT '小计',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_ticket_id` (`ticket_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单详情表';

-- 订单优惠券关联表
CREATE TABLE IF NOT EXISTS `order_coupons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `coupon_id` bigint unsigned NOT NULL COMMENT '优惠券ID',
  `user_coupon_id` bigint unsigned NOT NULL COMMENT '用户优惠券ID',
  `discount` decimal(10,2) NOT NULL COMMENT '优惠金额',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_coupon_id` (`coupon_id`),
  KEY `idx_user_coupon_id` (`user_coupon_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单优惠券关联表';

-- 秒杀活动表
CREATE TABLE IF NOT EXISTS `seckill_activities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '活动名称',
  `ticket_id` bigint unsigned NOT NULL COMMENT '门票ID',
  `price` decimal(10,2) NOT NULL COMMENT '秒杀价格',
  `total_stock` int NOT NULL COMMENT '总库存',
  `available_stock` int NOT NULL COMMENT '可用库存',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态: 0-未开始/已结束 1-进行中',
  `description` text COMMENT '活动描述',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ticket_id` (`ticket_id`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`),
  KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 秒杀库存表
CREATE TABLE IF NOT EXISTS `seckill_stock` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint unsigned NOT NULL COMMENT '活动ID',
  `total_stock` int NOT NULL DEFAULT '0' COMMENT '总库存',
  `used_stock` int NOT NULL DEFAULT '0' COMMENT '已使用库存',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀库存表';

-- 初始化管理员用户
INSERT INTO `users` (`username`, `password`, `email`, `phone`, `nickname`, `avatar`, `status`)
VALUES ('admin', 'e10adc3949ba59abbe56e057f20f883e', 'admin@example.com', '13800138000', '管理员', '', 1);

-- 初始化分类
INSERT INTO `categories` (`name`, `parent_id`, `sort`, `status`)
VALUES ('演唱会', 0, 1, 1),
       ('体育赛事', 0, 2, 1),
       ('话剧歌剧', 0, 3, 1),
       ('展览展会', 0, 4, 1),
       ('旅游景点', 0, 5, 1);

-- 初始化一些测试数据
INSERT INTO `tickets` (`title`, `description`, `image`, `category_id`, `price`, `original_price`, `stock`, `sold`, `status`)
VALUES ('周杰伦演唱会门票', '周杰伦2024年世界巡回演唱会', '', 1, 880.00, 1280.00, 1000, 0, 1),
       ('中超联赛门票', '中超联赛广州队vs上海队', '', 2, 180.00, 280.00, 5000, 0, 1),
       ('话剧《雷雨》', '经典话剧演出', '', 3, 120.00, 180.00, 200, 0, 1);

-- 初始化优惠券
INSERT INTO `coupons` (`name`, `type`, `discount`, `min_amount`, `max_discount`, `total_count`, `used_count`, `start_time`, `end_time`, `status`)
VALUES ('满100减20', 1, 20.00, 100.00, NULL, 1000, 0, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY), 1),
       ('8折优惠券', 2, 0.80, NULL, 50.00, 500, 0, NOW(), DATE_ADD(NOW(), INTERVAL 15 DAY), 1),
       ('秒杀券 - 50元', 3, 50.00, NULL, NULL, 100, 0, NOW(), DATE_ADD(NOW(), INTERVAL 1 HOUR), 1);

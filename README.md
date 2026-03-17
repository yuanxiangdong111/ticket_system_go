# 门票系统

## 项目概述

基于 Go 语言开发的门票系统，支持多种类型的优惠券（满减券、折扣券、秒杀券），并实现了优惠券叠加使用的规则。系统采用 MySQL 存储核心数据，Redis 作为缓存和秒杀活动的高性能支持。

## 技术栈

- **后端**: Go 1.21+
- **数据库**: MySQL 8.0+
- **缓存/秒杀支持**: Redis 7.0+
- **ORM**: GORM
- **HTTP 框架**: Gin
- **配置管理**: Viper
- **日志**: Zap
- **定时任务**: Gocron
- **API 文档**: Swagger

## 项目结构

```
ticket_system/
├── cmd/                    # 命令行入口
│   ├── api/               # API 服务入口
│   └── migrate/           # 数据库迁移工具
├── config/                # 配置文件
├── internal/              # 内部包
│   ├── controller/        # API 控制器
│   ├── dao/               # 数据访问层
│   ├── model/             # 数据模型
│   ├── service/           # 业务逻辑层
│   └── middleware/        # 中间件
├── pkg/                   # 公共包
│   ├── cache/             # Redis 缓存操作
│   ├── database/          # MySQL 数据库连接
│   ├── redis/             # Redis 连接
│   ├── util/              # 工具函数
│   └── validator/         # 参数校验
├── scripts/               # 脚本文件
├── test/                  # 测试文件
├── go.mod                 # Go 模块依赖
└── go.sum                 # Go 依赖锁定
```

## 功能特性

### 1. 优惠券系统

- **多种优惠券类型**:
  - 满减券: 满足一定金额条件后减免固定金额
  - 折扣券: 按折扣比例减免金额
  - 秒杀券: 特定时间内的超低价优惠券，数量有限

- **优惠券叠加规则**:
  - 满减券和折扣券可以叠加使用
  - 秒杀券不可与其他优惠券叠加
  - 自动计算最优优惠券组合

- **管理功能**:
  - 创建、删除、更新优惠券
  - 查看优惠券统计信息
  - 自动过期处理
  - 优惠券状态管理

### 2. 订单系统

- **完整的订单流程**:
  - 创建订单
  - 支付订单
  - 取消订单
  - 订单查询

- **支付集成**:
  - 支持多种支付方式
  - 支付状态跟踪
  - 超时自动取消

- **订单管理**:
  - 订单详情查询
  - 订单状态管理
  - 订单统计

### 3. 秒杀系统

- **高性能秒杀**:
  - Redis 预减库存
  - 分布式锁防止超卖
  - 异步订单处理

- **秒杀活动**:
  - 活动创建和管理
  - 库存预热
  - 实时库存查询

### 4. 用户系统

- **用户认证**:
  - JWT 认证
  - 登录/注册
  - 权限管理

- **用户信息**:
  - 个人信息管理
  - 地址管理
  - 历史订单

## 快速开始

### 1. 环境准备

- 安装 Go 1.21+
- 安装 MySQL 8.0+ 和 Redis 7.0+
- 克隆项目代码

### 2. 数据库配置

1. 创建数据库:
   ```sql
   CREATE DATABASE ticket_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. 配置环境变量或修改 `config/config.yaml` 文件

### 3. 项目启动

```bash
# 进入项目目录
cd "e:\opencode测试\门票系统"

# 安装依赖
go mod tidy

# 启动服务
cd cmd/api
go run main.go
```

### 4. API 文档

启动服务后，访问 `http://localhost:8080/swagger/index.html` 查看 API 文档

## 核心API

### 优惠券相关接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/coupons` | 获取优惠券列表 |
| GET | `/api/v1/coupons/available` | 获取可领取的优惠券 |
| POST | `/api/v1/coupons` | 创建优惠券（管理员） |
| POST | `/api/v1/coupons/receive` | 领取优惠券 |
| GET | `/api/v1/user/coupons` | 获取用户优惠券列表 |
| POST | `/api/v1/coupons/calculate` | 计算价格（含优惠券） |
| GET | `/api/v1/coupons/optimal/recommendation` | 获取最优优惠券组合 |

### 订单相关接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/orders` | 创建订单 |
| GET | `/api/v1/orders` | 获取订单列表 |
| GET | `/api/v1/orders/:id` | 获取订单详情 |
| PUT | `/api/v1/orders/:id/pay` | 支付订单 |
| PUT | `/api/v1/orders/:id/cancel` | 取消订单 |

### 秒杀相关接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/seckill/activities` | 获取秒杀活动列表 |
| GET | `/api/v1/seckill/activities/active` | 获取进行中的秒杀活动 |
| POST | `/api/v1/seckill/activities` | 创建秒杀活动 |
| POST | `/api/v1/seckill/order` | 创建秒杀订单 |
| GET | `/api/v1/seckill/activities/:id/stock` | 获取秒杀库存 |

## 优惠券叠加规则说明

### 规则总结

- **秒杀券优先级最高**: 如果使用了秒杀券，则不能使用其他类型的优惠券
- **折扣券**: 按折扣比例计算优惠，有最大折扣限制
- **满减券**: 满足最低金额条件后减免固定金额，可以叠加使用
- **优惠顺序**: 先计算折扣券，再计算满减券

### 计算示例

**场景**: 原价 100 元，使用一张 8 折优惠券（最高减 20 元）和一张满 50 减 10 的优惠券

**计算过程**:
1. 先计算折扣券: 100 * 0.8 = 80 元
2. 再计算满减券: 80 - 10 = 70 元
3. 最终价格: 70 元

## 秒杀系统设计

### 技术架构

- **Redis 预减库存**: 在秒杀开始前，将库存信息预热到 Redis 中
- **Lua 脚本原子操作**: 使用 Lua 脚本实现原子性库存检查和扣减
- **异步订单处理**: 秒杀成功后，异步创建订单，提高响应速度
- **限流防刷**: 限制用户请求频率，防止恶意请求
- **库存回补**: 订单取消时，及时回补库存

### 秒杀流程

1. 用户点击秒杀按钮
2. 系统检查活动是否有效
3. Redis 检查并预减库存
4. 用户购买标记
5. 异步创建订单
6. 更新数据库库存

## 部署建议

### 开发环境

```bash
# 开发环境建议配置
- Go 1.21+
- MySQL 8.0+
- Redis 7.0+
- 4GB 内存
- 2 核 CPU
```

### 生产环境

```bash
# 生产环境建议配置
- 多台应用服务器负载均衡
- MySQL 主从复制
- Redis 集群
- 监控和日志收集
- 自动扩容
```

## 监控和日志

### 日志配置

```yaml
log:
  level: info
  filename: logs/ticket_system.log
  max_size: 100
  max_backups: 3
  max_age: 28
```

### 监控指标

- **请求量**: QPS、响应时间
- **错误率**: 5xx、4xx 错误
- **系统资源**: CPU、内存、磁盘
- **数据库**: 连接数、查询耗时
- **Redis**: 内存使用、连接数

## 开发规范

### 代码风格

- 使用 go fmt 格式化代码
- 遵循 Go 语言规范
- 使用有意义的变量和函数名
- 添加适当的注释

### API 设计

- 遵循 RESTful API 规范
- 统一响应格式
- 错误处理规范
- 参数校验

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

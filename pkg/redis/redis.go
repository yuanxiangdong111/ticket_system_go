package redis

import (
	"context"
	"fmt"
	"ticket_system/pkg/util"
	"time"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client
var ctx = context.Background()

// Config Redis配置
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// Init 初始化Redis连接
func Init(config *Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return err
	}

	util.Info("Redis连接成功")
	return nil
}

// Close 关闭Redis连接
func Close() error {
	return Client.Close()
}

// Get 获取key
func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Set 设置key
func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// Del 删除key
func Del(key string) error {
	return Client.Del(ctx, key).Err()
}

// Exists 检查key是否存在
func Exists(key string) (bool, error) {
	result, err := Client.Exists(ctx, key).Result()
	return result > 0, err
}

// HGet 获取hash字段
func HGet(key, field string) (string, error) {
	return Client.HGet(ctx, key, field).Result()
}

// HSet 设置hash字段
func HSet(key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

// HGetAll 获取所有hash字段
func HGetAll(key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

// IncrBy 增加数值
func IncrBy(key string, value int64) (int64, error) {
	return Client.IncrBy(ctx, key, value).Result()
}

// DecrBy 减少数值
func DecrBy(key string, value int64) (int64, error) {
	return Client.DecrBy(ctx, key, value).Result()
}

// Eval 执行Lua脚本
func Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return Client.Eval(ctx, script, keys, args...).Result()
}

// SetNX 设置key（仅当key不存在时）
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return Client.SetNX(ctx, key, value, expiration).Result()
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

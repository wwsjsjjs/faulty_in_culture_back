// Package cache 提供 Redis 缓存操作功能
// 用于缓存用户登录信息和排名数据，提高查询性能
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache Redis 缓存管理器
type Cache struct {
	client *redis.Client // Redis 客户端
	ctx    context.Context
}

var (
	// 缓存实例
	cacheInstance *Cache
)

// InitCache 初始化 Redis 缓存连接
// 返回 error 如果连接失败
func InitCache() error {
	ctx := context.Background()

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址
		Password: "",               // 密码（如果设置了密码）
		DB:       0,                // 使用默认数据库
	})

	// 测试连接
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis 连接失败: %v", err)
	}

	cacheInstance = &Cache{
		client: client,
		ctx:    ctx,
	}

	return nil
}

// GetCache 获取缓存实例
func GetCache() *Cache {
	return cacheInstance
}

// Set 设置缓存键值，带过期时间
// key: 缓存键
// value: 缓存值（会自动序列化为 JSON）
// expiration: 过期时间
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	// 序列化为 JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %v", err)
	}

	// 设置缓存
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Get 获取缓存值
// key: 缓存键
// dest: 目标对象指针（会自动反序列化）
// 返回 error 如果键不存在或反序列化失败
func (c *Cache) Get(key string, dest interface{}) error {
	// 获取缓存数据
	data, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return err
	}

	// 反序列化
	return json.Unmarshal([]byte(data), dest)
}

// Delete 删除缓存键
func (c *Cache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Exists 检查键是否存在
func (c *Cache) Exists(key string) (bool, error) {
	result, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetWithoutExpiration 设置永久缓存（不过期）
func (c *Cache) SetWithoutExpiration(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %v", err)
	}
	return c.client.Set(c.ctx, key, data, 0).Err()
}

// ClearAll 清空所有缓存（慎用！）
func (c *Cache) ClearAll() error {
	return c.client.FlushDB(c.ctx).Err()
}

// SetExpiration 为已存在的键设置过期时间
func (c *Cache) SetExpiration(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

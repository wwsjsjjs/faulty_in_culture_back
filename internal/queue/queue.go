package queue

import (
	"context"
	"encoding/json"
	"faulty_in_culture/go_back/internal/logger"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	StreamName       = "message:stream"
	ConsumerGroup    = "message:group"
	ConsumerName     = "message:consumer:1"
	OfflineKeyPrefix = "offline:result:"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

// MessagePayload 消息任务载体
type MessagePayload struct {
	TaskID      string `json:"task_id"`
	UserID      uint   `json:"user_id"`
	Message     string `json:"message"`
	ProcessTime int64  `json:"process_time"` // Unix 时间戳，何时应该处理
}

// InitQueue 初始化 Redis Streams 队列
func InitQueue(redisAddr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis 连接失败: %v", err)
	}

	// 创建消费者组（如果不存在）
	err := rdb.XGroupCreateMkStream(ctx, StreamName, ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %v", err)
	}

	logger.Info("Redis Streams 队列已初始化")
	return nil
}

// EnqueueDelayedMessage 入队延迟消息任务
func EnqueueDelayedMessage(taskID string, userID uint, message string, delay time.Duration) error {
	processTime := time.Now().Add(delay).Unix()

	payload := MessagePayload{
		TaskID:      taskID,
		UserID:      userID,
		Message:     message,
		ProcessTime: processTime,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 添加到 Stream
	args := &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}

	msgID, err := rdb.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("添加消息到 Stream 失败: %v", err)
	}

	logger.Info("任务已入队", zap.String("taskID", taskID), zap.Uint("userID", userID), zap.Duration("delay", delay), zap.String("msgID", msgID))
	return nil
}

// StartWorker 启动 Redis Streams 消费者
func StartWorker(handler func(context.Context, *MessagePayload) error) {
	go func() {
		logger.Info("Redis Streams worker 已启动")

		for {
			// 读取消息（阻塞模式，等待新消息）
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    ConsumerGroup,
				Consumer: ConsumerName,
				Streams:  []string{StreamName, ">"},
				Count:    10,
				Block:    1 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue // 没有新消息
				}
				logger.Error("读取消息失败", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			// 处理消息
			for _, stream := range streams {
				for _, message := range stream.Messages {
					go processMessage(message, handler)
				}
			}
		}
	}()
}

// processMessage 处理单个消息
func processMessage(msg redis.XMessage, handler func(context.Context, *MessagePayload) error) {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		logger.Warn("消息格式错误", zap.String("msgID", msg.ID))
		rdb.XAck(ctx, StreamName, ConsumerGroup, msg.ID)
		return
	}

	var payload MessagePayload
	if err := json.Unmarshal([]byte(dataStr), &payload); err != nil {
		logger.Error("解析消息失败", zap.Error(err))
		rdb.XAck(ctx, StreamName, ConsumerGroup, msg.ID)
		return
	}

	// 检查是否到达处理时间
	now := time.Now().Unix()
	if now < payload.ProcessTime {
		// 还没到处理时间，延迟处理
		waitTime := time.Duration(payload.ProcessTime-now) * time.Second
		time.Sleep(waitTime)
	}

	// 执行处理逻辑
	if err := handler(ctx, &payload); err != nil {
		logger.Error("处理消息失败", zap.String("taskID", payload.TaskID), zap.Error(err))
		// 这里可以实现重试逻辑
	}

	// 确认消息已处理
	rdb.XAck(ctx, StreamName, ConsumerGroup, msg.ID)
}

// StoreOfflineMessage 存储离线消息到 Redis
func StoreOfflineMessage(taskID, message string) error {
	key := fmt.Sprintf("%s%s", OfflineKeyPrefix, taskID)
	return rdb.Set(ctx, key, message, 1*time.Hour).Err()
}

// GetOfflineMessage 获取离线消息
func GetOfflineMessage(taskID string) (string, error) {
	key := fmt.Sprintf("%s%s", OfflineKeyPrefix, taskID)
	return rdb.Get(ctx, key).Result()
}

// DeleteOfflineMessage 删除已读离线消息
func DeleteOfflineMessage(taskID string) error {
	key := fmt.Sprintf("%s%s", OfflineKeyPrefix, taskID)
	return rdb.Del(ctx, key).Err()
}

// GetUserOfflineMessages 获取用户的所有离线消息 key
func GetUserOfflineMessages(userID string) ([]string, error) {
	pattern := fmt.Sprintf("%s*", OfflineKeyPrefix)
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Shutdown 关闭队列
func Shutdown() {
	if rdb != nil {
		rdb.Close()
		logger.Info("Redis 连接已关闭")
	}
}

package scheduler

import (
	"faulty_in_culture/go_back/internal/logger"
	"time"

	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/models"

	"go.uber.org/zap"
)

// StartMessageCleanupScheduler 启动定时清理过期消息的调度器
// 从配置文件读取清理时间和清理策略
func StartMessageCleanupScheduler() {
	go func() {
		// 从配置读取清理时间（默认凌晨 2 点）
		cleanupHour := config.AppConfig.Message.CleanupScheduleHour
		if cleanupHour < 0 || cleanupHour > 23 {
			cleanupHour = 2 // 默认凌晨 2 点
		}

		// 计算下一次执行时间
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), cleanupHour, 0, 0, 0, now.Location())
		if now.After(next) {
			// 如果今天的清理时间已过，则设置为明天
			next = next.Add(24 * time.Hour)
		}
		duration := next.Sub(now)

		logger.Info("定时清理任务已启动", zap.String("next_time", next.Format("2006-01-02 15:04:05")))

		// 等待到下一次执行时间
		time.Sleep(duration)

		// 然后每 24 小时执行一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// 立即执行一次清理
		cleanupOldMessages()

		for range ticker.C {
			cleanupOldMessages()
		}
	}()
}

// cleanupOldMessages 清理过期的已完成消息
func cleanupOldMessages() {
	// 从配置读取清理天数
	cleanupDays := config.AppConfig.Message.CleanupDays
	if cleanupDays <= 0 {
		cleanupDays = 30 // 默认 30 天
	}

	cutoffTime := time.Now().AddDate(0, 0, -cleanupDays)

	result := database.DB.Where("status = ? AND processed_at < ?", "completed", cutoffTime).
		Delete(&models.Message{})

	if result.Error != nil {
		logger.Warn("清理过期消息失败", zap.Error(result.Error))
	} else {
		logger.Warn("清理过期消息完成", zap.Int64("deleted", result.RowsAffected), zap.Int("days", cleanupDays))
	}
}

// CleanupFailedMessages 清理失败消息（可选，手动调用或定时调用）
func CleanupFailedMessages() {
	// 从配置读取失败消息清理天数
	failedCleanupDays := config.AppConfig.Message.FailedCleanupDays
	if failedCleanupDays <= 0 {
		failedCleanupDays = 7 // 默认 7 天
	}

	cutoffTime := time.Now().AddDate(0, 0, -failedCleanupDays)

	result := database.DB.Where("status = ? AND created_at < ?", "failed", cutoffTime).
		Delete(&models.Message{})

	if result.Error != nil {
		logger.Warn("清理失败消息失败", zap.Error(result.Error))
	} else {
		logger.Warn("清理失败消息完成", zap.Int64("deleted", result.RowsAffected), zap.Int("days", failedCleanupDays))
	}
}

package scheduler

import (
	"log"
	"time"

	"github.com/yourusername/ranking-api/internal/config"
	"github.com/yourusername/ranking-api/internal/database"
	"github.com/yourusername/ranking-api/internal/models"
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

		log.Printf("定时清理任务已启动，下次执行时间: %s", next.Format("2006-01-02 15:04:05"))

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
		log.Printf("清理过期消息失败: %v", result.Error)
	} else {
		log.Printf("清理过期消息完成，删除了 %d 条记录（%d 天前的已完成消息）", result.RowsAffected, cleanupDays)
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
		log.Printf("清理失败消息失败: %v", result.Error)
	} else {
		log.Printf("清理失败消息完成，删除了 %d 条记录（%d 天前的失败消息）", result.RowsAffected, failedCleanupDays)
	}
}

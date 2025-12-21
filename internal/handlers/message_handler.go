package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"faulty_in_culture/go_back/internal/config"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/dto"

	"github.com/yourusername/ranking-api/internal/models"
	"github.com/yourusername/ranking-api/internal/queue"
	"github.com/yourusername/ranking-api/internal/vo"
	ws "github.com/yourusername/ranking-api/internal/websocket"
)

var (
	wsManager *ws.Manager
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域，生产环境需严格配置
		},
	}
)

// InitMessageHandler 初始化消息处理器
func InitMessageHandler(manager *ws.Manager) {
	wsManager = manager
}

// SendMessage 发送消息接口
// @Summary 发送延迟消息
// @Description 发送一段消息，配置时间后返回
// @Tags message
// @Accept json
// @Produce json
// @Param data body dto.SendMessageRequest true "消息内容"
// @Success 200 {object} map[string]string
// @Router /api/send-message [post]
func SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	taskID := uuid.New().String()

	// 创建消息记录（pending 状态）
	// 注意：这里不保存返回内容，只保存请求消息
	message := models.Message{
		TaskID:  taskID,
		UserID:  req.UserID,
		Content: req.Message, // 这是请求消息
		Status:  "pending",
	}
	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建消息记录失败"})
		return
	}

	// 入队延迟任务（从配置读取延迟时间）
	delaySeconds := config.AppConfig.Message.DelaySeconds
	if delaySeconds <= 0 {
		delaySeconds = 10 // 默认 10 秒
	}
	err := queue.EnqueueDelayedMessage(taskID, req.UserID, req.Message, time.Duration(delaySeconds)*time.Second)
	if err != nil {
		// 入队失败，更新消息状态为 failed
		database.DB.Model(&message).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "任务入队失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"message": fmt.Sprintf("消息已接收，将在%d秒后返回", delaySeconds),
	})
}

// QueryResult 查询消息结果接口
// @Summary 查询消息结果
// @Description 根据任务ID查询延迟消息结果
// @Tags message
// @Produce json
// @Param task_id query string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/query-result [get]
func QueryResult(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 task_id 参数"})
		return
	}

	result, err := queue.GetOfflineMessage(taskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "pending", "message": "结果尚未就绪"})
		return
	}

	// 删除已读消息
	queue.DeleteOfflineMessage(taskID)

	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"result": result,
	})
}

// HandleWebSocket WebSocket 连接处理
func HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 user_id 参数"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	// 注册连接
	wsManager.Register(userID, conn)
	defer wsManager.Unregister(userID)

	// 设置 PongHandler，收到 Pong 时更新活跃时间
	conn.SetPongHandler(func(appData string) error {
		wsManager.UpdateActiveTime(userID)
		return nil
	})

	// 连接建立后，推送所有离线消息
	go pushOfflineMessages(userID, conn)

	// 保持连接，监听前端消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败:", err)
			break
		}
		// 收到消息时也更新活跃时间
		wsManager.UpdateActiveTime(userID)
		log.Printf("收到来自用户 %s 的消息: %s", userID, string(message))
	}
}

// pushOfflineMessages 推送所有离线消息
func pushOfflineMessages(userID string, conn *websocket.Conn) {
	keys, err := queue.GetUserOfflineMessages(userID)
	if err != nil {
		log.Println("获取离线消息失败:", err)
		return
	}

	for _, key := range keys {
		// 从 key 中提取 taskID
		taskID := key[len("offline:result:"):]
		result, err := queue.GetOfflineMessage(taskID)
		if err != nil {
			continue
		}

		// 推送消息
		msg := map[string]interface{}{
			"task_id": taskID,
			"result":  result,
			"type":    "offline",
		}
		data, _ := json.Marshal(msg)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("推送离线消息失败:", err)
			continue
		}

		// 删除已推送的离线消息
		queue.DeleteOfflineMessage(taskID)
		log.Printf("已推送离线消息: userID=%s, taskID=%s", userID, taskID)
	}
}

// ProcessDelayedMessage Redis Streams 任务处理函数
func ProcessDelayedMessage(ctx context.Context, payload *queue.MessagePayload) error {
	log.Printf("处理延迟任务: taskID=%s, userID=%s", payload.TaskID, payload.UserID)

	// 模拟处理逻辑，生成返回结果（实际应用中可能是复杂的处理）
	// 这里假设返回的文本与请求不一样
	resultMessage := fmt.Sprintf("处理完成: %s [processed at %s]",
		payload.Message,
		time.Now().Format("2006-01-02 15:04:05"))

	// 更新消息状态为 completed，并写入返回结果
	now := time.Now()
	if err := database.DB.Model(&models.Message{}).
		Where("task_id = ?", payload.TaskID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"content":      resultMessage, // 更新为返回的文本
			"processed_at": &now,
		}).Error; err != nil {
		log.Printf("更新消息状态失败: %v", err)
	}

	// 检查用户是否在线
	if wsManager.IsOnline(payload.UserID) {
		// 在线，通过 WebSocket 推送
		msg := map[string]interface{}{
			"task_id": payload.TaskID,
			"result":  resultMessage,
			"type":    "realtime",
		}
		data, _ := json.Marshal(msg)
		if err := wsManager.SendMessage(payload.UserID, data); err != nil {
			log.Printf("WebSocket 推送失败: %v", err)
			// 推送失败，存储为离线消息
			return queue.StoreOfflineMessage(payload.TaskID, resultMessage)
		}
		log.Printf("已通过 WebSocket 推送消息: userID=%s, taskID=%s", payload.UserID, payload.TaskID)
	} else {
		// 离线，存储到 Redis
		if err := queue.StoreOfflineMessage(payload.TaskID, resultMessage); err != nil {
			return fmt.Errorf("存储离线消息失败: %v", err)
		}
		log.Printf("用户离线，消息已存储: userID=%s, taskID=%s", payload.UserID, payload.TaskID)
	}

	return nil
}

// GetMessages 获取用户的历史消息列表
// @Summary 获取历史消息列表
// @Description 获取指定用户的历史消息记录，支持分页和状态筛选
// @Tags message
// @Produce json
// @Param user_id query string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param status query string false "状态筛选" Enums(pending, completed, failed)
// @Success 200 {object} vo.MessageListResponse
// @Failure 400 {object} map[string]string
// @Router /api/messages [get]
func GetMessages(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 user_id 参数"})
		return
	}

	var req dto.GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 默认分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	offset := (req.Page - 1) * req.Limit

	// 构建查询
	query := database.DB.Model(&models.Message{}).Where("user_id = ?", userID)
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取分页数据
	var messages []models.Message
	if err := query.Order("created_at DESC").
		Limit(req.Limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 转换为 VO
	var messageResponses []vo.MessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, vo.MessageResponse{
			ID:          msg.ID,
			TaskID:      msg.TaskID,
			UserID:      msg.UserID,
			Content:     msg.Content,
			Status:      msg.Status,
			ProcessedAt: msg.ProcessedAt,
			CreatedAt:   msg.CreatedAt,
			UpdatedAt:   msg.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, vo.MessageListResponse{
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
		Messages: messageResponses,
	})
}

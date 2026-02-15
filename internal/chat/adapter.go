// Package chat - AI聊天模块适配器
// 功能：提供AI客户端适配器，实现与外部AI服务的集成
package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"faulty_in_culture/go_back/internal/infra/config"
	"faulty_in_culture/go_back/internal/infra/logger"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ============================================================
// AI客户端适配器 - 实现AIClient接口
// 使用腾讯混元API（兼容OpenAI接口规范）
// ============================================================

// AIClientAdapter AI客户端适配器
type AIClientAdapter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewAIClient 创建AI客户端实例
func NewAIClient() AIClient {
	cfg := &config.GlobalConfig.AI
	return &AIClientAdapter{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenAI兼容的请求结构
type chatRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
}

// OpenAI兼容的响应结构
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat 调用AI聊天接口（腾讯混元API，兼容OpenAI规范）
func (a *AIClientAdapter) Chat(ctx context.Context, messages []map[string]string) (string, error) {
	// 检查API Key
	if a.apiKey == "" {
		logger.Warn("AI API Key未配置，返回模拟回复")
		return "AI功能未配置，请设置HUNYUAN_API_KEY环境变量", nil
	}

	// 转换消息格式
	reqMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		reqMessages[i] = map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		}
	}

	// 构建请求
	reqBody := chatRequest{
		Model:    a.model,
		Messages: reqMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("序列化请求失败", zap.Error(err))
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := a.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("创建请求失败", zap.Error(err))
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	// 发送请求
	logger.Info("调用AI接口",
		zap.String("url", url),
		zap.String("model", a.model),
		zap.Int("message_count", len(messages)))

	resp, err := a.client.Do(req)
	if err != nil {
		logger.Error("调用AI接口失败", zap.Error(err))
		return "", fmt.Errorf("调用AI接口失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("读取响应失败", zap.Error(err))
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		logger.Error("AI接口返回错误",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return "", fmt.Errorf("AI接口错误 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		logger.Error("解析响应失败",
			zap.Error(err),
			zap.String("response", string(body)))
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取AI回复
	if len(chatResp.Choices) == 0 {
		logger.Error("AI响应中没有回复内容")
		return "", fmt.Errorf("AI响应中没有回复内容")
	}

	content := chatResp.Choices[0].Message.Content
	logger.Info("AI回复成功",
		zap.Int("prompt_tokens", chatResp.Usage.PromptTokens),
		zap.Int("completion_tokens", chatResp.Usage.CompletionTokens),
		zap.Int("total_tokens", chatResp.Usage.TotalTokens))

	return content, nil
}

// ChatStream 流式调用AI聊天接口（支持实时推送）
func (a *AIClientAdapter) ChatStream(ctx context.Context, messages []map[string]string, callback func(chunk string)) (string, error) {
	// 检查API Key
	if a.apiKey == "" {
		logger.Warn("AI API Key未配置，返回模拟回复")
		mockReply := "AI功能未配置，请设置HUNYUAN_API_KEY环境变量"
		if callback != nil {
			callback(mockReply)
		}
		return mockReply, nil
	}

	// 转换消息格式
	reqMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		reqMessages[i] = map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		}
	}

	// 构建请求（启用流式传输）
	reqBody := map[string]interface{}{
		"model":    a.model,
		"messages": reqMessages,
		"stream":   true, // 启用流式传输
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("序列化请求失败", zap.Error(err))
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := a.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("创建请求失败", zap.Error(err))
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	logger.Info("调用AI接口（流式）",
		zap.String("url", url),
		zap.String("model", a.model),
		zap.Int("message_count", len(messages)))

	resp, err := a.client.Do(req)
	if err != nil {
		logger.Error("调用AI接口失败", zap.Error(err))
		return "", fmt.Errorf("调用AI接口失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("AI接口返回错误",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return "", fmt.Errorf("AI接口错误 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	// 读取流式响应
	var fullContent string
	reader := io.Reader(resp.Body)
	buffer := make([]byte, 4096)

	logger.Info("开始接收AI流式响应")

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			chunk := string(buffer[:n])

			// 解析SSE格式的数据
			lines := bytes.Split([]byte(chunk), []byte("\n"))
			for _, line := range lines {
				if len(line) == 0 {
					continue
				}

				// SSE数据格式: "data: {...}"
				if bytes.HasPrefix(line, []byte("data: ")) {
					jsonStr := bytes.TrimPrefix(line, []byte("data: "))

					// 跳过 [DONE] 标记
					if bytes.Equal(jsonStr, []byte("[DONE]")) {
						logger.Info("接收到流式结束标记")
						continue
					}

					// 解析JSON
					var streamResp struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
							FinishReason *string `json:"finish_reason"`
						} `json:"choices"`
					}

					if err := json.Unmarshal(jsonStr, &streamResp); err == nil {
						if len(streamResp.Choices) > 0 {
							content := streamResp.Choices[0].Delta.Content
							if content != "" {
								fullContent += content
								if callback != nil {
									callback(content)
								}
								logger.Debug("收到AI chunk", zap.String("content", content))
							}
						}
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				logger.Info("AI流式响应接收完成",
					zap.Int("total_length", len(fullContent)))
				break
			}
			logger.Error("读取流式响应失败", zap.Error(err))
			return fullContent, fmt.Errorf("读取流式响应失败: %w", err)
		}
	}

	if fullContent == "" {
		logger.Warn("AI返回空内容")
		return "", fmt.Errorf("AI返回空内容")
	}

	return fullContent, nil
}

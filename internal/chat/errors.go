package chat

import "errors"

var (
	ErrSessionNotFound = errors.New("会话不存在")
	ErrUnauthorized    = errors.New("无权访问此会话")
	ErrMessageTooMany  = errors.New("对话过长，请创建新对话")
	ErrInvalidIndex    = errors.New("消息序号无效")
)

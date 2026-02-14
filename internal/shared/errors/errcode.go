// Package errors 提供统一的错误码和错误消息管理
// 功能：集中管理所有业务错误码，提供一致的错误响应格式
package errors

// 错误码定义 - 使用分段管理
const (
	// 通用错误码 0-9999
	Success       = 0
	ServerError   = 10000
	InvalidParams = 10001
	NotFound      = 10002

	// 用户相关错误 20000-20999
	UserNotFound      = 20001
	UserAlreadyExists = 20002
	InvalidPassword   = 20003
	Unauthorized      = 20004
	TokenExpired      = 20005
	TokenInvalid      = 20006

	// 排行榜相关错误 30000-30999
	InvalidRankType   = 30001
	RankingNotFound   = 30002
	InvalidScore      = 30003
	UpdateScoreFailed = 30004

	// 聊天相关错误 40000-40999
	SessionNotFound    = 40001
	MessageTooLong     = 40002
	SessionLimitExceed = 40003
	AIServiceError     = 40004

	// 存档相关错误 50000-50999
	InvalidSlotNumber = 50001
	SaveGameNotFound  = 50002
	SaveGameExists    = 50003
	SlotNotAvailable  = 50004
)

// messages 错误码对应的消息映射
var messages = map[int]string{
	Success:       "操作成功",
	ServerError:   "服务器内部错误",
	InvalidParams: "参数错误",
	NotFound:      "资源不存在",

	UserNotFound:      "用户不存在",
	UserAlreadyExists: "用户名已存在",
	InvalidPassword:   "用户名或密码错误",
	Unauthorized:      "未授权，请先登录",
	TokenExpired:      "Token已过期",
	TokenInvalid:      "Token无效",

	InvalidRankType:   "排行榜类型无效",
	RankingNotFound:   "排行榜记录不存在",
	InvalidScore:      "分数无效",
	UpdateScoreFailed: "更新分数失败",

	SessionNotFound:    "会话不存在",
	MessageTooLong:     "消息内容过长",
	SessionLimitExceed: "会话数量超过限制",
	AIServiceError:     "AI服务调用失败",

	InvalidSlotNumber: "存档槽位号无效",
	SaveGameNotFound:  "存档不存在",
	SaveGameExists:    "存档已存在",
	SlotNotAvailable:  "存档槽位不可用",
}

// GetMessage 根据错误码获取对应的错误消息
func GetMessage(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}

// Error 标准错误结构
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *Error) Error() string {
	return e.Message
}

// New 创建新的错误
func New(code int) *Error {
	return &Error{
		Code:    code,
		Message: GetMessage(code),
	}
}

// NewWithMessage 创建带自定义消息的错误
func NewWithMessage(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

package user

import "errors"

// ============================================================
// 领域错误定义
// 设计模式：领域驱动设计(DDD) - 领域错误
// 职责：集中定义用户领域的所有业务错误
// ============================================================

var (
	// 认证相关错误
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户名已存在")
	ErrInvalidPassword   = errors.New("用户名或密码错误")
	ErrUnauthorized      = errors.New("未授权")

	// 分数相关错误
	ErrInvalidRankType = errors.New("排行榜类型无效，必须在1-9之间")
	ErrInvalidScore    = errors.New("分数无效")

	// 系统错误
	ErrHashPassword  = errors.New("密码加密失败")
	ErrCreateUser    = errors.New("创建用户失败")
	ErrUpdateScore   = errors.New("更新分数失败")
	ErrDatabaseError = errors.New("数据库操作失败")
)

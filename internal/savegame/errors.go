package savegame

import "errors"

var (
	// ErrSaveGameNotFound 存档未找到
	ErrSaveGameNotFound = errors.New("存档不存在")
	// ErrInvalidSlotNumber 无效的槽位号
	ErrInvalidSlotNumber = errors.New("槽位号必须在1-6之间")
	// ErrUnauthorized 无权限访问
	ErrUnauthorized = errors.New("无权限访问该存档")
)

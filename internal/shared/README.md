# Shared模块设计说明

## 📁 目录结构

```
internal/shared/
├── infra/              # 基础设施层（Infrastructure）
│   ├── config/        # 配置管理（YAML加载）
│   ├── db/            # 数据库连接池（GORM + MySQL）
│   ├── cache/         # Redis缓存
│   ├── logger/        # 日志系统（Zap）
│   └── ws/            # WebSocket管理器
│
├── middleware/         # HTTP中间件
│   ├── auth.go        # JWT认证中间件
│   └── limiter.go     # 限流中间件（Redis）
│
└── pkg/               # 项目内部工具包
    ├── response/      # HTTP响应封装
    │   └── response.go
    ├── security/      # 安全工具（密码哈希、Token）
    │   └── auth.go
    └── convert/       # 类型转换（用户ID解析等）
        └── user.go
```

## 🎯 设计原则

### 1️⃣ infra/ - 基础设施层
**特征**：
- 提供底层技术支撑（数据库、缓存、日志等）
- 与第三方服务交互（MySQL、Redis、WebSocket）
- 生命周期管理（连接池、心跳检测）

**示例**：
```go
import "faulty_in_culture/go_back/internal/shared/infra/db"
database := db.GetDB()
```

### 2️⃣ middleware/ - HTTP中间件层
**特征**：
- Gin框架中间件（`gin.HandlerFunc`）
- 请求/响应拦截器
- 认证、限流、日志记录

**示例**：
```go
import "faulty_in_culture/go_back/internal/shared/middleware"
router.Use(middleware.AuthMiddleware())
```

### 3️⃣ pkg/ - 项目内部工具包
**特征**：
- 与业务逻辑耦合（使用internal结构）
- 被多个领域模块复用
- 提供HTTP上下文相关的工具函数

**子包分类**：
- **response/** - 统一响应格式（Code/Msg/Data结构）
- **security/** - 安全工具（密码哈希bcrypt、Token生成、MD5）
- **convert/** - 类型转换（从gin.Context解析用户ID等）

**示例**：
```go
import (
    "faulty_in_culture/go_back/internal/shared/pkg/security"
    "faulty_in_culture/go_back/internal/shared/pkg/convert"
)

hash, _ := security.HashPassword("123456")
userID, _ := convert.GetUserID(c)
```

## 🆚 与项目根目录 pkg/ 的区别

### 项目根目录 `pkg/`
```
pkg/
└── checker/          # ✓ 纯工具，无internal依赖
    └── deps.go       # MySQL/Redis连接检查
```
- **可被外部项目导入**（如其他Go项目）
- **零业务依赖**
- 可以单独开源为独立库

### Internal `shared/pkg/`
```
internal/shared/pkg/
├── response/         # ✗ 依赖项目的错误码定义
├── security/         # ✗ 为本项目认证系统定制
└── convert/          # ✗ 依赖gin.Context等项目框架
```
- **仅本项目可用**（Go编译器强制）
- **有业务依赖**（使用internal路径）
- 与项目架构耦合

## 📝 使用示例

### 在Handler中使用
```go
package user

import (
    "github.com/gin-gonic/gin"
    "faulty_in_culture/go_back/internal/shared/pkg/convert"
    "faulty_in_culture/go_back/internal/shared/pkg/security"
)

func (h *Handler) UpdateScore(c *gin.Context) {
    // 使用convert获取用户ID
    userID, ok := convert.GetUserID(c)
    if !ok {
        c.JSON(400, gin.H{"error": "无效的用户ID"})
        return
    }
    
    // 使用security验证密码
    if !security.CheckPassword(password, user.PasswordHash) {
        c.JSON(401, gin.H{"error": "密码错误"})
        return
    }
}
```

### 在Service中使用
```go
package user

import "faulty_in_culture/go_back/internal/shared/pkg/security"

func (s *Service) Register(username, password string) (*Entity, string, error) {
    // 密码哈希
    hash, err := security.HashPassword(password)
    if err != nil {
        return nil, "", err
    }
    
    // 生成Token
    token := security.GenerateToken(user.ID, user.Username)
    return user, token, nil
}
```

## ⚙️ 依赖关系

```
领域层（user/chat/savegame）
         ↓
    shared/pkg/          ← 业务工具
         ↓
   shared/middleware/    ← HTTP中间件
         ↓
    shared/infra/        ← 基础设施
         ↓
    第三方库（MySQL/Redis/Zap等）
```

## 🔄 迁移说明

**已完成**：
- ✅ `internal/shared/config` → `internal/shared/infra/config`
- ✅ `internal/shared/db` → `internal/shared/infra/db`
- ✅ `internal/shared/cache` → `internal/shared/infra/cache`
- ✅ `internal/shared/logger` → `internal/shared/infra/logger`
- ✅ `internal/shared/ws` → `internal/shared/infra/ws`
- ✅ `internal/shared/response` → `internal/shared/pkg/response`
- ✅ `internal/shared/utils/auth.go` 拆分为：
  - `internal/shared/pkg/security/auth.go`（密码、Token）
  - `internal/shared/pkg/convert/user.go`（ID解析）

**待实现**：
- 🔲 创建 `internal/shared/pkg/types/` 存放公共类型、错误码、常量
- 🔲 考虑添加 `internal/shared/pkg/validator/` 存放自定义验证器

## 📌 最佳实践

1. **infra/** 只提供技术能力，不包含业务逻辑
2. **middleware/** 只做拦截和转发，不修改业务数据
3. **pkg/** 保持函数式编程风格，避免全局状态
4. 新增工具函数前，先判断应该放在哪个子包：
   - 操作数据库? → `infra/db/`
   - HTTP拦截器? → `middleware/`
   - 安全相关? → `pkg/security/`
   - 类型转换? → `pkg/convert/`
   - 其他业务工具? → 考虑创建新的 `pkg/xxx/`

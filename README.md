# 排名系统API

一个基于Go语言的用户排名管理系统后端API，使用Gin框架和GORM。





## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin（高性能 HTTP Web 框架）
- **ORM**: GORM（Go 语言 ORM 库，支持 MySQL/SQLite/PostgreSQL 等）
- **数据库**: MySQL（默认，支持切换 SQLite/PostgreSQL）
- **密码加密**: bcrypt（用户密码安全存储）
- **API 文档**: swaggo/swag 自动生成 Swagger/OpenAPI 3.0 文档
- **接口测试**: Swagger UI（内置，支持在线调试）
- **缓存中间件**: Redis（可选，支持会话、排行榜、验证码等高性能缓存场景）
- **容器化**: Docker（应用容器化，便于部署和环境隔离）
- **日志与监控**: ELK、Prometheus、Grafana（可选，日志采集、监控、可视化）
- **分层结构**: handler、model、dto、vo、route、database 等分层清晰

## 项目结构

```
go_back/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── models/
│   │   └── ranking.go           # 数据模型定义
│   ├── handlers/
│   │   └── ranking_handler.go  # API处理器
│   ├── database/
│   │   └── db.go                # 数据库配置
│   └── routes/
│       └── routes.go            # 路由配置
├── docs/                        # Swagger文档（自动生成）
├── go.mod                       # Go模块依赖
├── go.sum                       # 依赖校验
└── README.md                    # 项目说明
```

## 安装依赖

### 1. 安装Go语言

如果还没有安装Go，请访问 [https://golang.org/dl/](https://golang.org/dl/) 下载并安装Go 1.21或更高版本。

### 2. 验证安装

```bash
go version
```

### 3. 安装项目依赖

```bash
# 已经完成，依赖已在go.mod中
go mod download
```

### 4. 安装Swagger CLI工具

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

**重要**: 确保 `$GOPATH/bin` 在系统PATH中。



## 生成API文档

```bash
swag init -g cmd/server/main.go -o docs
```

## 运行项目

### 方式1: 直接运行

```bash
go run cmd/server/main.go
```

### 方式2: 编译后运行

```bash
# 编译
go build -o ranking-api.exe cmd/server/main.go

# 运行
./ranking-api.exe
```

## 主要接口方法

### 用户相关
- POST `/api/register` 用户注册
- POST `/api/login` 用户登录

### 排名相关
- POST `/api/rankings` 创建新排名
- GET `/api/rankings` 获取排名列表（分页）
- GET `/api/rankings/{id}` 获取单个排名
- PUT `/api/rankings/{id}` 更新排名
- DELETE `/api/rankings/{id}` 删除排名
- GET `/api/rankings/top` 获取排行榜前N名

### 消息/异步任务相关
- POST `/api/send-message` 发送延迟消息（10秒后返回）
- GET `/api/query-result` 查询消息结果
- GET `/api/messages` 获取历史消息列表（支持分页和状态筛选）
- GET `/ws` WebSocket 实时消息推送

### 存档相关 🆕（需要认证）
- GET `/api/savegames` 获取用户所有存档
- GET `/api/savegames/{slot}` 获取指定槽位存档（1-6）
- PUT `/api/savegames/{slot}` 创建或更新存档
- DELETE `/api/savegames/{slot}` 删除存档

### AI聊天相关 🆕（需要认证）
- POST `/api/chat/start` 开始新的AI聊天会话
- POST `/api/chat/send` 发送消息给AI（异步，通过WebSocket返回）
- GET `/api/chat/sessions` 获取所有聊天会话
- GET `/api/chat/{session_id}` 获取聊天历史

详细参数和返回格式请参考 [Swagger 文档](http://localhost:8080/swagger/index.html)。

## 认证说明

需要认证的接口需要在请求头中添加 `Authorization` token：
```
Authorization: {userID}:{username}:{timestamp}
```

示例：
```bash
# 登录获取token
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 使用token访问需要认证的接口
curl -X GET http://localhost:8080/api/savegames \
  -H "Authorization: 1:test:1703059200"
```

## 环境变量

在启动服务前，需要设置以下环境变量：

```bash
# 混元AI API Key（用于AI聊天功能）
export HUNYUAN_API_KEY="your-api-key-here"

# 可选：端口号（默认8080）
export PORT=8080
```

## 消息队列持久化说明

### 持久化方案
本项目使用 **Redis Streams** 作为消息队列，实现了以下持久化机制：

1. **Redis Streams 持久化**：
   - 使用 Redis Streams 作为消息队列，支持消费者组、消息确认等特性
   - Redis 支持 RDB（快照）和 AOF（追加日志）两种持久化方式
   - 推荐配置：开启 AOF 持久化（`appendonly yes`），确保任务不丢失
   - Streams 特性：消息持久化、消费者组、消息确认、自动重试

2. **数据库双重持久化**：
   - **接收时**：消息发送时立即保存到 MySQL（`messages` 表），状态为 `pending`，只保存请求消息
   - **处理时**：队列倒计时结束后，更新消息状态为 `completed`，写入处理后的返回文本
   - 记录字段：消息ID、用户ID、内容（先保存请求，后更新为返回）、状态、处理时间等
   - 即使 Redis 故障，消息记录仍然完整保存

3. **定时清理机制（可配置）**：
   - 清理时间可在 `config.yaml` 中配置
   - 默认每天凌晨 2 点自动清理 30 天前的已完成消息
   - 7 天前的失败消息可手动清理
   - 防止数据库无限增长，保持系统性能

### 配置说明
在 `config.yaml` 中配置消息队列参数：
```yaml
redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

message:
  delay_seconds: 10          # 消息延迟处理时间（秒）
  cleanup_days: 30           # 清理30天前的已完成消息
  failed_cleanup_days: 7     # 清理7天前的失败消息
  cleanup_schedule_hour: 2   # 每天凌晨2点执行清理
```

### Redis 持久化配置建议
在 Redis 配置文件（`redis.conf`）中添加：
```conf
# 开启 AOF 持久化
appendonly yes
appendfsync everysec

# 或使用 RDB 快照（根据需求选择）
save 900 1
save 300 10
save 60 10000
```

### 消息处理流程
1. 用户发送消息 → 保存到数据库（pending，保存请求消息）
2. 消息入队 Redis Streams（带延迟时间）
3. Worker 消费消息，等待延迟时间
4. 处理完成 → 更新数据库（completed，写入返回文本）
5. 通过 WebSocket 推送给在线用户，或存储为离线消息


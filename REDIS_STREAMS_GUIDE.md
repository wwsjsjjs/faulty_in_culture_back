# Redis Streams 消息队列实现说明

## 主要改动

### 1. 配置文件 (config.yaml)
新增 Redis 和消息配置：
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

### 2. 消息队列实现 (internal/queue/queue.go)
- 从 Asynq 迁移到 **Redis Streams**
- 使用消费者组（Consumer Group）实现消息确认和重试
- 支持延迟消息处理
- 自动处理消息确认（XAck）

### 3. 消息存储逻辑
**接收时（SendMessage）**：
- 保存到数据库，状态为 `pending`
- 只保存用户的请求消息内容
- 消息入队到 Redis Streams

**处理时（ProcessDelayedMessage）**：
- 模拟处理逻辑，生成返回结果
- 更新数据库：状态改为 `completed`，内容更新为处理后的结果
- 通过 WebSocket 推送或存储为离线消息

### 4. 定时清理
- 从配置文件读取清理参数
- 默认每天凌晨2点清理30天前的已完成消息
- 支持配置清理时间和天数

## 使用步骤

1. 确保 Redis 已启动并配置好持久化
2. 更新 config.yaml 中的 Redis 配置
3. 启动服务：`go run cmd/server/main.go`
4. 发送消息：POST /api/send-message
5. 查看历史：GET /api/messages?user_id=xxx

## Redis Streams 优势

- ✅ 原生支持消息持久化
- ✅ 消费者组支持自动重试
- ✅ 消息确认机制（XAck）
- ✅ 支持多个消费者并发处理
- ✅ 无需额外依赖（只需 Redis）
- ✅ 性能优异，适合高并发场景

## 注意事项

1. **Redis 持久化**：建议开启 AOF（appendonly yes）
2. **消息延迟**：通过 ProcessTime 字段实现，Worker 会等待到指定时间
3. **错误处理**：保持原有的错误处理逻辑不变
4. **可配置**：所有时间参数都在 config.yaml 中可配置

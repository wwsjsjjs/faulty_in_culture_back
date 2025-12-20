# WebSocket + 延迟消息功能测试指南

## 前置要求

### 1. 安装 Redis

#### Windows
- 下载 Redis for Windows: https://github.com/microsoftarchive/redis/releases
- 或使用 Docker: `docker run -d -p 6379:6379 redis:latest`

#### 启动 Redis
```bash
# 方式1：直接启动 Redis
redis-server

# 方式2：使用 Docker
docker run -d -p 6379:6379 --name redis redis:latest
```

### 2. 验证 Redis 启动
```bash
redis-cli ping
# 应返回 PONG
```

---

## 启动服务

### 1. 启动 Go 后端
```bash
cd d:\zlearning\project\faulty_in_culture\go_back
go run ./cmd/server/main.go
```

### 2. 确认服务运行
- 访问健康检查: http://localhost:8080/health
- 访问 Swagger 文档: http://localhost:8080/swagger/index.html

---

## 测试场景

### 测试 1：在线推送（WebSocket实时推送）

1. 打开测试页面 `test_ws.html`（双击打开或用浏览器打开）
2. 输入用户ID：`user123`
3. 点击 "连接 WebSocket"，状态变为 "已连接"
4. 输入消息内容：`Hello World`
5. 点击 "发送消息"，记录返回的任务ID
6. **等待10秒**，WebSocket 会收到实时推送的消息

**预期结果**：10秒后，在"接收到的消息"区域看到 WebSocket 推送的消息。

---

### 测试 2：离线存储（HTTP轮询查询）

1. 打开测试页面 `test_ws.html`
2. 输入用户ID：`user456`
3. **不连接 WebSocket**
4. 输入消息内容：`Offline Message`
5. 点击 "发送消息"，记录任务ID（如：`abc-123-def`）
6. 等待10秒后，在"查询结果"区域输入任务ID
7. 点击 "查询结果"

**预期结果**：查询返回消息内容 `Offline Message`。

---

### 测试 3：离线消息自动推送（重连推送）

1. 打开测试页面 `test_ws.html`
2. 输入用户ID：`user789`
3. **不连接 WebSocket**
4. 点击 "发送消息"，记录任务ID
5. 等待10秒（此时用户离线，消息会存储到Redis）
6. 点击 "连接 WebSocket"

**预期结果**：连接建立后，立即收到之前的离线消息推送。

---

## 接口测试（使用 Postman/Apifox/curl）

### 1. 发送消息
```bash
curl -X POST http://localhost:8080/api/send-message \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123","message":"Hello from API"}'
```

**响应示例**：
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "消息已接收，将在10秒后返回"
}
```

### 2. 查询结果
```bash
curl "http://localhost:8080/api/query-result?task_id=550e8400-e29b-41d4-a716-446655440000"
```

**响应示例**（未就绪）：
```json
{
  "status": "pending",
  "message": "结果尚未就绪"
}
```

**响应示例**（已完成）：
```json
{
  "status": "completed",
  "result": "Hello from API"
}
```

### 3. WebSocket 连接
使用浏览器或 WebSocket 客户端连接：
```
ws://localhost:8080/ws?user_id=user123
```

**接收消息格式**：
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "result": "Hello from API",
  "type": "realtime"  // 或 "offline"
}
```

---

## 功能验证清单

- [ ] Redis 已启动并可连接
- [ ] Go 服务器成功启动
- [ ] 健康检查返回 200
- [ ] 发送消息返回任务ID
- [ ] WebSocket 成功建立连接
- [ ] 在线用户10秒后收到实时推送
- [ ] 离线消息成功存储到 Redis
- [ ] HTTP 查询可获取离线消息
- [ ] 重连后自动推送离线消息

---

## 常见问题

### 1. Redis 连接失败
```
ERROR: dial tcp [::1]:6379: connectex: No connection could be made
```

**解决方法**：确保 Redis 服务已启动：
```bash
# 检查 Redis
redis-cli ping

# 启动 Redis
redis-server
# 或
docker run -d -p 6379:6379 redis:latest
```

### 2. WebSocket 连接失败

**原因**：未提供 user_id 参数

**解决方法**：连接 URL 必须包含 user_id
```
ws://localhost:8080/ws?user_id=your_user_id
```

### 3. 消息未收到

**排查步骤**：
1. 检查 Redis 是否正常
2. 查看服务器日志，确认任务入队
3. 确认 WebSocket 连接状态
4. 检查用户ID是否一致

---

## 监控与调试

### 查看 Redis 队列
```bash
redis-cli
> KEYS *
> LLEN asynq:default:pending
> GET offline:result:任务ID
```

### 服务器日志关键信息
- `任务已入队`: 消息成功入队
- `用户 xxx 已连接`: WebSocket 连接成功
- `已通过 WebSocket 推送消息`: 实时推送成功
- `用户离线，消息已存储`: 离线存储成功
- `已推送离线消息`: 重连推送成功

---

## 架构说明

```
前端 ───┬──► HTTP /api/send-message ──► asynq 队列 ──► 延迟10秒
        │                                            │
        │                                            ▼
        │                                    判断用户在线?
        │                                       /        \
        │                                    在线        离线
        │                                      │          │
        └──► WebSocket /ws ◄────────── 实时推送    存Redis
                 │                                    │
                 │                                    │
                 └────────── 重连自动推送 ◄───────────┘
                                  │
                         HTTP /api/query-result
```

---

完成测试后，所有功能应正常运行！

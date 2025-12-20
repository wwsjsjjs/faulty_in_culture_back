# 功能测试指南

## 1. 启动 Redis

在运行测试之前，请确保 Redis 已经启动：

```powershell
# 方式1：双击 D:\yy\Redis-8.4.0\start-redis.bat 启动
# 方式2：命令行启动
cd D:\yy\Redis-8.4.0
.\redis-server.exe
```

验证 Redis 是否运行：
```powershell
cd D:\yy\Redis-8.4.0
.\redis-cli.exe ping
# 应该返回 PONG
```

## 2. 启动后端服务器

Redis 启动后，重新启动 Go 服务器：

```powershell
cd D:\zlearning\project\faulty_in_culture\go_back
go run ./cmd/server/main.go
```

成功启动后应该看到：
- ✅ Database connection established (MySQL)
- ✅ Database migration completed
- ✅ Redis cache initialized successfully
- ✅ Asynq 队列已初始化
- ✅ Asynq worker 已启动
- ✅ Server starting on http://localhost:8080

## 3. 测试 Redis 缓存功能

### 3.1 测试用户登录缓存

**第一次登录（从数据库查询）：**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"testuser\",\"password\":\"password123\"}"
```

**第二次登录（从缓存读取，速度更快）：**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"testuser\",\"password\":\"password123\"}"
```

### 3.2 测试排名缓存

**获取排行榜 Top 10（第一次查数据库）：**
```bash
curl http://localhost:8080/api/rankings/top?top=10
```

**再次获取（从缓存读取）：**
```bash
curl http://localhost:8080/api/rankings/top?top=10
```

**获取分页排名（第一次）：**
```bash
curl http://localhost:8080/api/rankings?page=1&limit=10
```

**再次获取同一页（从缓存读取）：**
```bash
curl http://localhost:8080/api/rankings?page=1&limit=10
```

### 3.3 验证缓存失效

创建新排名后，相关缓存应该被清除：

```bash
curl -X POST http://localhost:8080/api/rankings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d "{\"username\":\"newuser\",\"score\":999}"
```

再次查询排行榜，应该返回更新后的数据（从数据库重新查询）：
```bash
curl http://localhost:8080/api/rankings/top?top=10
```

## 4. 测试 WebSocket + 消息队列功能

### 4.1 使用测试页面

打开浏览器，访问 `file:///D:/zlearning/project/faulty_in_culture/go_back/test_ws.html`

**测试步骤：**

1. **连接 WebSocket**
   - 输入用户 ID（如：123）
   - 点击"连接 WebSocket"
   - 应该看到 "WebSocket 已连接"

2. **发送消息（在线接收）**
   - 在"消息内容"输入框输入测试消息（如："Hello World"）
   - 点击"发送消息"
   - 等待 10 秒后，应该在"WebSocket 消息"区域收到推送的消息

3. **发送消息（离线存储）**
   - 点击"断开连接"
   - 在"消息内容"输入框输入新消息（如："Offline Message"）
   - 点击"发送消息"
   - 点击"查询结果"按钮，应该看到消息状态为 "pending"
   - 重新连接 WebSocket，应该自动收到之前的离线消息

### 4.2 使用 curl 测试

**发送消息：**
```bash
curl -X POST http://localhost:8080/api/send-message \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":123,\"content\":\"Test message\"}"
```

返回示例：
```json
{
  "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "message": "消息已提交，将在 10 秒后处理"
}
```

**查询消息结果：**
```bash
curl "http://localhost:8080/api/query-result?task_id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

返回示例（处理中）：
```json
{
  "status": "pending",
  "message": "消息处理中"
}
```

返回示例（已推送）：
```json
{
  "status": "delivered",
  "message": "消息已推送",
  "content": "Test message"
}
```

## 5. 性能对比测试

### 5.1 测试登录性能（缓存 vs 数据库）

**清除缓存：**
```bash
# 使用 redis-cli 清除缓存
cd D:\yy\Redis-8.4.0
.\redis-cli.exe
> FLUSHDB
> exit
```

**对比测试：**
```bash
# 第一次登录（从数据库查询）
time curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"testuser\",\"password\":\"password123\"}"

# 第二次登录（从缓存读取）
time curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"testuser\",\"password\":\"password123\"}"
```

### 5.2 测试排名查询性能

```bash
# 清除缓存
redis-cli FLUSHDB

# 第一次查询（从数据库）
time curl http://localhost:8080/api/rankings/top?top=100

# 第二次查询（从缓存）
time curl http://localhost:8080/api/rankings/top?top=100
```

## 6. Redis 数据查看

### 6.1 查看所有缓存键

```bash
cd D:\yy\Redis-8.4.0
.\redis-cli.exe
> KEYS *
```

应该看到类似：
```
1) "user:username:testuser"
2) "rankings:top:10"
3) "rankings:page:1:limit:10"
```

### 6.2 查看具体缓存内容

```bash
> GET "user:username:testuser"
> GET "rankings:top:10"
```

### 6.3 查看缓存过期时间

```bash
> TTL "user:username:testuser"    # 用户缓存 24 小时 = 86400 秒
> TTL "rankings:top:10"            # 排名缓存 5 分钟 = 300 秒
```

## 7. 故障排查

### 7.1 Redis 连接失败

**错误信息：**
```
Warning: Failed to initialize cache: Redis 连接失败
```

**解决方法：**
1. 确认 Redis 已启动：`.\redis-cli.exe ping`
2. 检查端口 6379 是否被占用：`netstat -ano | findstr 6379`
3. 重启 Redis 服务

### 7.2 Asynq 连接失败

**错误信息：**
```
cannot subscribe to cancelation channel: UNKNOWN: redis pubsub receive error
```

**解决方法：**
- 与 Redis 连接失败相同，确保 Redis 正常运行

### 7.3 WebSocket 连接失败

**错误信息（浏览器控制台）：**
```
WebSocket connection to 'ws://localhost:8080/ws?user_id=123' failed
```

**解决方法：**
1. 确认服务器已启动
2. 检查防火墙设置
3. 确认端口 8080 未被占用

## 8. 预期结果

### 8.1 缓存命中率

正常情况下：
- 首次查询：从数据库获取（较慢，10-50ms）
- 缓存命中：从 Redis 获取（快速，1-5ms）
- 性能提升：5-10 倍

### 8.2 消息延迟推送

- 用户在线：消息发送后 10 秒准时推送
- 用户离线：消息存储在 Redis，用户上线后立即推送
- 任务队列：可靠处理，支持重试和持久化

## 9. 监控建议

### 9.1 Redis 内存使用

```bash
redis-cli INFO memory
```

### 9.2 查看队列状态

```bash
redis-cli
> KEYS asynq:*
> LLEN asynq:queues:default
```

### 9.3 服务器日志

查看控制台输出：
- 缓存命中日志
- 消息处理日志
- WebSocket 连接日志

---

## 测试完成后

请确认以下功能正常：
- ✅ Redis 缓存正常工作
- ✅ 登录缓存有效（24小时）
- ✅ 排名缓存有效（5分钟）
- ✅ 缓存在数据更新时正确失效
- ✅ WebSocket 连接正常
- ✅ 消息延迟 10 秒推送
- ✅ 离线消息正确存储和推送
- ✅ Asynq 队列正常工作

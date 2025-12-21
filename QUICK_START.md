# 快速开始指南

## 1. 启动服务

### 开发模式（日志输出到终端）
```bash
cd d:\zlearning\project\faulty_in_culture\go_back
go run ./cmd/server/main.go
```

或使用编译后的可执行文件：
```bash
.\bin\server.exe
```

### 生产模式（日志输出到文件）
```bash
# Windows PowerShell
$env:LOG_MODE="prod"; $env:GIN_MODE="release"; go run ./cmd/server/main.go

# 或设置环境变量后运行
$env:LOG_MODE="prod"
$env:GIN_MODE="release"
.\bin\server.exe
```

## 2. 验证服务启动

打开浏览器访问：
- 健康检查: http://localhost:8080/health
- Swagger文档: http://localhost:8080/swagger/index.html

## 3. 使用 Apifox 测试 API

### 第一步：用户注册
1. 在 Apifox 中新建 HTTP 请求
2. 方法：POST
3. URL: `http://localhost:8080/api/register`
4. Headers: `Content-Type: application/json`
5. Body (JSON):
```json
{
  "username": "testuser",
  "password": "123456"
}
```
6. 点击发送
7. **保存响应中的 token**，格式类似：`1:testuser:1703152800`

### 第二步：用户登录
1. 方法：POST
2. URL: `http://localhost:8080/api/login`
3. Headers: `Content-Type: application/json`
4. Body (JSON):
```json
{
  "username": "testuser",
  "password": "123456"
}
```

### 第三步：测试需要认证的接口

#### 创建排名
1. 方法：POST
2. URL: `http://localhost:8080/api/rankings`
3. Headers:
   - `Content-Type: application/json`
   - `Authorization: {你的token}` （使用第一步保存的token）
4. Body (JSON):
```json
{
  "user_id": 1,
  "score": 9999,
  "level": 10
}
```

#### 创建存档
1. 方法: PUT
2. URL: `http://localhost:8080/api/savegames/1`
3. Headers:
   - `Content-Type: application/json`
   - `Authorization: {你的token}`
4. Body (JSON):
```json
{
  "slot": 1,
  "user_id": 1,
  "game_data": "{\"level\":5,\"hp\":100,\"items\":[\"sword\",\"shield\"]}"
}
```

### 第四步：测试 WebSocket

1. 在 Apifox 中新建 WebSocket 请求
2. URL: `ws://localhost:8080/ws?user_id=1`
3. 点击"连接"
4. 连接成功后：
   - 会自动收到离线消息（如果有）
   - 每10秒收到一次Ping心跳
   - 可在下方输入框发送消息测试

### 第五步：测试消息队列

#### 发送延迟消息
1. 方法: POST
2. URL: `http://localhost:8080/api/send-message`
3. Headers: `Content-Type: application/json`
4. Body (JSON):
```json
{
  "user_id": "1",
  "message": "这是一条测试消息",
  "delay_seconds": 5
}
```
5. **保存响应中的 task_id**

#### 查询消息处理结果
1. 方法: GET
2. URL: `http://localhost:8080/api/query-result?task_id={task_id}`
3. 等待5秒后查询，应该看到处理完成的消息

## 4. 查看日志

### 开发模式
日志直接显示在终端，格式如下：
```
2025-12-21 15:04:05 INFO    main: 应用程序启动
2025-12-21 15:04:06 INFO    handlers.Register: 用户注册成功 {"user_id": 1, "username": "testuser"}
2025-12-21 15:04:10 INFO    handlers.Login: 登录成功 {"user_id": 1, "username": "testuser"}
```

### 生产模式
查看日志文件：
```bash
# 实时查看日志
Get-Content logs\app.log -Wait

# 查看最后100行
Get-Content logs\app.log -Tail 100
```

## 5. 常见问题

### Q: 服务启动失败，提示数据库连接错误
A: 检查 `config.yaml` 中的数据库配置，确保 MySQL 服务已启动

### Q: Redis 相关错误
A: 确保 Redis 服务已启动：
```bash
# 如果使用 Docker
docker run -d -p 6379:6379 redis

# 检查 Redis 是否运行
redis-cli ping
# 应返回 PONG
```

### Q: 认证失败
A: 确保在 Headers 中添加了正确的 Authorization 字段，值为注册/登录返回的完整 token

### Q: WebSocket 连接失败
A: 检查 URL 中是否包含 user_id 参数

### Q: 限流错误（429 Too Many Requests）
A: 当前限流为60次/分钟，等待一分钟后重试

## 6. 环境变量说明

| 变量名 | 说明 | 默认值 | 可选值 |
|--------|------|--------|--------|
| LOG_MODE | 日志模式 | dev | dev, prod |
| GIN_MODE | Gin运行模式 | debug | debug, release |
| PORT | 服务端口 | 8080 | 任意可用端口 |

## 7. 目录说明

- `logs/` - 生产模式日志文件目录（自动创建）
- `bin/` - 编译后的可执行文件目录
- `docs/` - 文档目录
  - `APIFOX_TESTING_GUIDE.md` - 详细测试指南
  - `swagger.json/yaml` - Swagger API文档

## 8. 下一步

- 阅读 `PROJECT_UPGRADE_SUMMARY.md` 了解项目升级详情
- 阅读 `docs/APIFOX_TESTING_GUIDE.md` 获取完整API测试指南
- 查看 `docs/swagger.json` 查看完整API定义

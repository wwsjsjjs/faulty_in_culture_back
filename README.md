# 排名系统API

一个基于Go语言的用户排名管理系统后端API，使用Gin框架和GORM。

## 技术栈

- **Web框架**: Gin (高性能HTTP Web框架)
- **ORM**: GORM (Go语言ORM库)
- **数据库**: SQLite (轻量级数据库，可轻松切换到MySQL/PostgreSQL)
- **API文档**: Swagger/OpenAPI 3.0
- **语言**: Go 1.21+

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

Windows PowerShell中添加到PATH:
```powershell
$env:Path += ";$env:USERPROFILE\go\bin"
# 或永久添加（需要管理员权限）
[System.Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", [System.EnvironmentVariableTarget]::User)
```

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

### 方式3: 使用go run（推荐开发时使用）

```bash
cd cmd/server
go run main.go
```

服务器将在 `http://localhost:8080` 启动

## API端点

### 基础

- **健康检查**: `GET /health`
- **Swagger文档**: `GET /swagger/index.html`

### 排名管理

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | /api/rankings | 创建新排名 |
| GET | /api/rankings | 获取所有排名（分页） |
| GET | /api/rankings/top | 获取前N名 |
| GET | /api/rankings/:id | 获取单个排名 |
| PUT | /api/rankings/:id | 更新排名 |
| DELETE | /api/rankings/:id | 删除排名 |

## API使用示例

### 1. 创建排名

```bash
curl -X POST http://localhost:8080/api/rankings \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","score":1000}'
```

### 2. 获取所有排名（分页）

```bash
curl http://localhost:8080/api/rankings?page=1&limit=10
```

### 3. 获取前10名

```bash
curl http://localhost:8080/api/rankings/top?top=10
```

### 4. 获取单个排名

```bash
curl http://localhost:8080/api/rankings/1
```

### 5. 更新排名

```bash
curl -X PUT http://localhost:8080/api/rankings/1 \
  -H "Content-Type: application/json" \
  -d '{"score":1500}'
```

### 6. 删除排名

```bash
curl -X DELETE http://localhost:8080/api/rankings/1
```

## PowerShell示例（Windows）

```powershell
# 创建排名
Invoke-RestMethod -Uri "http://localhost:8080/api/rankings" `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"username":"player1","score":1000}'

# 获取所有排名
Invoke-RestMethod -Uri "http://localhost:8080/api/rankings?page=1&limit=10"

# 获取前10名
Invoke-RestMethod -Uri "http://localhost:8080/api/rankings/top?top=10"
```

## 数据库

项目使用SQLite数据库，数据库文件 `ranking.db` 会自动创建在项目根目录。

### 切换到其他数据库

如需使用MySQL或PostgreSQL，修改 `internal/database/db.go`:

#### MySQL示例:

```go
import "gorm.io/driver/mysql"

// 安装驱动
// go get -u gorm.io/driver/mysql

dsn := "user:password@tcp(127.0.0.1:3306)/ranking_db?charset=utf8mb4&parseTime=True&loc=Local"
DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

#### PostgreSQL示例:

```go
import "gorm.io/driver/postgres"

// 安装驱动
// go get -u gorm.io/driver/postgres

dsn := "host=localhost user=gorm password=gorm dbname=ranking_db port=5432 sslmode=disable"
DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

## 环境变量

- `PORT`: 服务器端口（默认: 8080）
- `GIN_MODE`: Gin运行模式 (`debug` 或 `release`，默认: debug）

```bash
# Windows PowerShell
$env:PORT="3000"
$env:GIN_MODE="release"
go run cmd/server/main.go
```

## 测试工具推荐

### 1. Swagger UI（内置）

访问 `http://localhost:8080/swagger/index.html` 可以直接在浏览器中测试API。

### 2. Postman

下载地址: [https://www.postman.com/downloads/](https://www.postman.com/downloads/)

导入Swagger文档:
1. 打开Postman
2. File → Import
3. 输入URL: `http://localhost:8080/swagger/doc.json`

### 3. Thunder Client (VS Code扩展)

在VS Code中安装Thunder Client扩展，直接在编辑器中测试API。

### 4. curl / Invoke-RestMethod

使用命令行工具测试（见上方示例）

## 常见问题

### Q: swag命令找不到？

A: 确保已安装swag并添加到PATH:

```powershell
# 安装
go install github.com/swaggo/swag/cmd/swag@latest

# 检查安装路径
where.exe swag

# 如果找不到，添加到PATH
$env:Path += ";$env:USERPROFILE\go\bin"
```

### Q: 端口被占用？

A: 更改端口:

```bash
$env:PORT="3000"
go run cmd/server/main.go
```

### Q: 数据库连接失败？

A: SQLite不需要额外配置。如果使用MySQL/PostgreSQL，检查:
- 数据库服务是否运行
- 连接字符串是否正确
- 用户权限是否足够

## 开发流程

1. 修改代码
2. 如果修改了API注释，重新生成Swagger文档: `swag init -g cmd/server/main.go -o docs`
3. 重启服务器
4. 在Swagger UI中测试: `http://localhost:8080/swagger/index.html`

## 部署

### 编译生产版本

```bash
# Windows
go build -o ranking-api.exe -ldflags="-s -w" cmd/server/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o ranking-api -ldflags="-s -w" cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o ranking-api -ldflags="-s -w" cmd/server/main.go
```

### 运行生产版本

```bash
# 设置为release模式
$env:GIN_MODE="release"
./ranking-api.exe
```

## 许可

MIT License

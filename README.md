# Go Back - 游戏后端API服务

基于Go语言的游戏后端API系统，采用**简化MVC架构**，按领域垂直切分。

##  简化MVC架构

\\\
internal/
 user/          # 用户领域（完整MVC）
    entity.go       # 实体层
    repository.go   # 数据访问层  
    service.go      # 业务逻辑层
    handler.go      # HTTP控制层
    dto.go          # 数据传输对象
    strategy.go     # 策略模式
    errors.go       # 领域错误
 chat/          # 聊天领域
 savegame/      # 存档领域  
 shared/        # 共享基础设施
    config/     db/
   cache/
    logger/
    ws/
    middleware/
    response/
    utils/
 routes/        # 路由配置
\\\

##  设计模式

user模块展示了8种设计模式：
1. **实体模式** (Entity) - entity.go
2. **DTO/VO模式** - dto.go
3. **策略模式** (Strategy) - strategy.go（9种排行榜）
4. **工厂模式** (Factory) - strategy.go
5. **仓储模式** (Repository) - repository.go
6. **服务层模式** - service.go
7. **依赖注入** (DI) - 构造函数注入
8. **MVC模式** - handler为Controller

##  API接口 (13个)

**用户模块**：注册、登录、排行榜(1-9种)、更新分数  
**聊天模块**：创建对话、发送消息、获取历史、撤回消息
**存档模块**：查询、创建、更新、删除（6个槽位）

文档：http://localhost:8080/swagger/index.html

##  快速开始

### 开发环境
\\\bash
# 1. 安装依赖
go mod download

# 2. 生成Swagger文档
swag init -g cmd/server/main.go -o docs

# 3. 使用开发配置启动（默认）
go run cmd/server/main.go
\\\

### 生产环境
\\\bash
# 使用生产配置启动
cp config.prod.yaml config.yaml
# 修改config.yaml中的敏感信息后启动
go run cmd/server/main.go
\\\

##  配置文件

项目支持**开发环境**和**生产环境**分离配置：

- **config.yaml** - 开发环境配置（默认）
- **config.prod.yaml** - 生产环境配置模板

配置项包括：
- **app**: 应用配置（环境、端口、日志模式、Gin模式）
- **database**: MySQL数据库配置
- **redis**: Redis缓存配置
- **ai**: AI服务配置（腾讯混元等）
- **jwt**: JWT认证配置

**生产环境注意事项**：
- 将敏感信息（密码、密钥）配置为环境变量
- 设置 `app.environment: production`
- 设置 `app.gin_mode: release`
- 设置 `database.auto_create: false`

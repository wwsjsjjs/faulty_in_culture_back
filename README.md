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

\\\
go mod download
swag init -g cmd/server/main.go -o docs
go run cmd/server/main.go
\\\

配置文件 \config.yaml\ 需包含MySQL、Redis、AI配置。

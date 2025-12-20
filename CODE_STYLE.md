# Go 项目代码规范

## 1. 文件与目录结构
- 遵循标准 Go 项目布局，业务代码放在 `internal/`，入口放在 `cmd/`，配置放在根目录。

## 2. 注释规范
- 所有导出类型、函数、方法、接口、结构体字段必须有注释，注释以被注释对象名开头。
- 业务 handler 必须有 Swagger 注释，便于自动生成 API 文档。
- 复杂逻辑、关键分支、魔法数字等必须有行内注释。
- 示例：
  ```go
  // Register 用户注册
  // @Summary 用户注册
  // @Description 用户注册，传入用户名和密码
  // @Tags user
  // @Accept json
  // @Produce json
  // @Param data body dto.UserRegisterRequest true "注册信息"
  // @Success 200 {object} vo.UserVO
  // @Failure 400 {object} map[string]string
  // @Router /api/register [post]
  func Register(c *gin.Context) { ... }
  ```

## 3. 命名规范
- 包名、文件名、变量名、函数名、结构体名、字段名均采用小驼峰或大驼峰，语义清晰。
- 常量全大写，单词下划线分隔。

## 4. 代码风格
- 每个包有独立的 doc.go 文件说明用途。
- 每个 handler、model、dto、vo 文件顶部有用途说明。
- 结构体字段加 json tag。
- 业务逻辑分层清晰，禁止跨层直接调用。

## 5. 依赖管理
- 所有依赖通过 go.mod 管理，禁止私自引入未声明依赖。

## 6. 其他
- 重要变更需更新 README 和 Swagger 注释。
- 代码提交前请 go fmt、go vet 保证格式和静态检查通过。

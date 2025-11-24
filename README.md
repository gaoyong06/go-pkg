# go-pkg

通用 Go 工具包，提供各种实用的中间件和工具函数，用于简化开发流程、统一代码风格和提高代码质量。支持 Gin 和 Kratos 两种框架。

## 目录

- [功能概述](#功能概述)
- [安装方法](#安装方法)
- [模块说明](#模块说明)
  - [错误处理](#错误处理)
  - [分页中间件](#分页中间件)
  - [过滤中间件](#过滤中间件)
- [使用示例](#使用示例)
- [贡献指南](#贡献指南)

## 功能概述

本工具包目前提供以下功能：

1. **统一错误处理**：标准化的错误类型和处理机制，支持错误代码、详细信息和 HTTP 状态码映射
2. **分页中间件**：处理 API 请求中的分页参数，支持标准化的分页响应格式
3. **过滤中间件**：处理 API 请求中的过滤、排序和搜索参数

后续将逐步添加更多功能，如：

- 认证与授权中间件
- 日志中间件
- 配置管理
- 数据库工具
- 缓存工具
- 等等...

## 安装方法

```bash
go get github.com/gaoyong06/go-pkg
```

## 模块说明

本工具包支持以下两种框架：

- **Gin 框架**：适用于传统的 HTTP API 服务
- **Kratos 框架**：适用于微服务架构

### 错误处理

`errors` 包提供了统一的错误处理机制，包括：

- 标准化的错误类型（验证错误、数据库错误、资源不存在错误等）
- 错误详情支持（字段级别的错误信息）
- HTTP 状态码映射
- 统一的错误响应格式

```go
// 创建验证错误
err := errors.NewValidationError("用户数据验证失败", nil).
    AddDetail("name", "用户名不能为空").
    AddDetail("email", "邮箱格式不正确")

// 创建资源不存在错误
err := errors.NewNotFoundError("找不到指定的用户", nil).
    AddDetail("id", "ID为123的用户不存在")

// 在 Gin 框架中使用错误处理中间件
r.Use(error.ErrorHandlerMiddleware())

// 在 Kratos 框架中使用错误处理中间件
server := http.NewServer(
    http.Address(":8000"),
    http.Middleware(error.KratosErrorHandlerMiddleware())
)
```

### 分页中间件

`middleware/pagination` 包提供了处理 API 分页的中间件和工具函数，支持 Gin 和 Kratos 两种框架：

- 提取和验证分页参数（页码、每页数量）
- 计算分页偏移量
- 标准化的分页响应格式
- 分页相关的响应头

```go
// 在 Gin 框架中使用分页中间件
r.GET("/users", pagination.Middleware(), listUsers)

// 在 Gin 处理函数中使用分页参数
func listUsers(c *gin.Context) {
    // 获取分页参数
    page, pageSize := pagination.GetPageParams(c)
    offset := pagination.GetOffset(page, pageSize)
    
    // 查询数据库（示例）
    users, total := db.GetUsers(offset, pageSize)
    
    // 返回分页响应
    pagination.RespondWithPagination(c, users, total)
}

// 在 Kratos 框架中使用分页中间件
server := http.NewServer(
    http.Address(":8000"),
    http.Middleware(pagination.KratosMiddleware())
)

// 在 Kratos 处理函数中使用分页参数
func listUsers(ctx http.Context) error {
    // 获取分页参数
    paginationInfo := pagination.GetPaginationFromContext(ctx)
    page := paginationInfo.Page
    pageSize := paginationInfo.PageSize
    offset := pagination.GetOffset(page, pageSize)
    
    // 查询数据库（示例）
    users, total := db.GetUsers(offset, pageSize)
    
    // 返回分页响应
    result := pagination.NewPaginationResult(users, total, page, pageSize)
    return ctx.JSON(200, result)
}
```

### 过滤中间件

`middleware/filter` 包提供了处理 API 过滤、排序和搜索的中间件和工具函数，支持 Gin 和 Kratos 两种框架：

- 支持多种过滤操作符（等于、不等于、大于、小于、包含等）
- 支持多字段排序（升序、降序）
- 支持全文搜索
- 支持字段选择

```go
// 在 Gin 框架中使用过滤中间件
r.GET("/users", filter.Middleware(), listUsers)

// 在 Gin 处理函数中使用过滤选项
func listUsers(c *gin.Context) {
    // 获取过滤选项
    filterOptions := filter.GetFilterOptions(c)
    
    // 构建查询条件（示例）
    whereClause, args := filter.BuildWhereClause(filterOptions.Filters)
    orderByClause := filter.BuildOrderByClause(filterOptions.Sorts)
    
    // 查询数据库...
}

// 在 Kratos 框架中使用过滤中间件
server := http.NewServer(
    http.Address(":8000"),
    http.Middleware(filter.KratosMiddleware())
)

// 在 Kratos 处理函数中使用过滤选项
func listUsers(ctx http.Context) error {
    // 获取过滤选项
    filterOptions := filter.GetFilterOptionsFromContext(ctx)
    
    // 构建查询条件（示例）
    whereClause, args := filter.BuildWhereClause(filterOptions.Filters)
    orderByClause := filter.BuildOrderByClause(filterOptions.Sorts)
    
    // 查询数据库...
    return ctx.JSON(200, result)
}
```

## 使用示例

完整的使用示例请参考 `examples` 目录下的示例代码：

- `examples/main.go`：基于 Gin 框架的示例
- `examples/kratos_example/main.go`：基于 Kratos 框架的示例

## 贡献指南

欢迎贡献代码或提出改进建议！请遵循以下步骤：

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建一个 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详情请参见 LICENSE 文件

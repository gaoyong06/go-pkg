package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/gaoyong06/go-pkg/middleware/error"
	"github.com/gaoyong06/go-pkg/middleware/filter"
	"github.com/gaoyong06/go-pkg/middleware/pagination"
	"github.com/gin-gonic/gin"
)

// 用户模型
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Company  string `json:"company"`
	Position string `json:"position"`
}

// 模拟数据库中的用户数据
var users = []User{
	{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 30, Company: "新元科技", Position: "技术总监"},
	{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 25, Company: "新元科技", Position: "前端工程师"},
	{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 35, Company: "新元科技", Position: "后端工程师"},
	{ID: 4, Name: "赵六", Email: "zhaoliu@example.com", Age: 28, Company: "新元科技", Position: "产品经理"},
	{ID: 5, Name: "钱七", Email: "qianqi@example.com", Age: 40, Company: "新元科技", Position: "CEO"},
}

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 注册中间件
	r.Use(error.ErrorHandlerMiddleware()) // 错误处理中间件

	// API 路由
	v1 := r.Group("/v1")
	{
		// 用户相关接口
		users := v1.Group("/users")
		{
			// 获取用户列表，使用分页和过滤中间件
			users.GET("", pagination.Middleware(), filter.Middleware(), listUsers)

			// 获取单个用户
			users.GET("/:id", getUser)

			// 创建用户
			users.POST("", createUser)
		}
	}

	// 启动服务器
	fmt.Println("服务器启动在 http://localhost:8080")
	r.Run(":8080")
}

// listUsers 获取用户列表
func listUsers(c *gin.Context) {
	// 从上下文中获取分页参数
	page, pageSize := pagination.GetPageParams(c)
	offset := pagination.GetOffset(page, pageSize)

	// 从上下文中获取过滤选项
	filterOptions := filter.GetFilterOptions(c)

	// 应用过滤条件（实际项目中会转换为数据库查询）
	filteredUsers := filterUsers(users, filterOptions)

	// 应用分页
	start := offset
	end := offset + pageSize
	if start >= len(filteredUsers) {
		// 返回空数组而不是错误
		pagination.RespondWithPagination(c, []User{}, len(filteredUsers))
		return
	}

	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	pagedUsers := filteredUsers[start:end]

	// 返回分页响应
	pagination.RespondWithPagination(c, pagedUsers, len(filteredUsers))
}

// getUser 获取单个用户
func getUser(c *gin.Context) {
	// 获取用户 ID
	id := c.Param("id")
	if id == "" {
		c.Error(errors.NewValidationError("用户 ID 不能为空", nil))
		return
	}

	// 模拟查找用户
	var user *User
	for _, u := range users {
		if fmt.Sprintf("%d", u.ID) == id {
			user = &u
			break
		}
	}

	// 检查用户是否存在
	if user == nil {
		c.Error(errors.NewNotFoundError("用户不存在", nil).AddDetail("id", fmt.Sprintf("ID 为 %s 的用户不存在", id)))
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// createUser 创建用户
func createUser(c *gin.Context) {
	// 解析请求体
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.Error(errors.NewValidationError("无效的用户数据", err))
		return
	}

	// 验证用户数据
	if newUser.Name == "" {
		c.Error(errors.NewValidationError("用户数据验证失败", nil).AddDetail("name", "用户名不能为空"))
		return
	}

	if newUser.Email == "" {
		c.Error(errors.NewValidationError("用户数据验证失败", nil).AddDetail("email", "邮箱不能为空"))
		return
	}

	// 模拟创建用户（实际项目中会保存到数据库）
	newUser.ID = len(users) + 1
	users = append(users, newUser)

	// 返回创建的用户
	c.JSON(http.StatusCreated, gin.H{"data": newUser})
}

// filterUsers 根据过滤条件过滤用户
func filterUsers(users []User, options *filter.FilterOptions) []User {
	if len(options.Filters) == 0 && options.Search == "" {
		return users
	}

	var result []User

	// 应用过滤条件
	for _, user := range users {
		if matchesFilters(user, options.Filters) && matchesSearch(user, options.Search) {
			result = append(result, user)
		}
	}

	// 应用排序
	if len(options.Sorts) > 0 {
		// 实际项目中会实现排序逻辑
		// 这里简化处理
	}

	return result
}

// matchesFilters 检查用户是否匹配过滤条件
func matchesFilters(user User, filters []filter.FilterCondition) bool {
	if len(filters) == 0 {
		return true
	}

	for _, f := range filters {
		switch f.Field {
		case "name":
			if !matchStringFilter(user.Name, f) {
				return false
			}
		case "email":
			if !matchStringFilter(user.Email, f) {
				return false
			}
		case "age":
			if !matchIntFilter(user.Age, f) {
				return false
			}
		case "company":
			if !matchStringFilter(user.Company, f) {
				return false
			}
		case "position":
			if !matchStringFilter(user.Position, f) {
				return false
			}
		}
	}

	return true
}

// matchStringFilter 检查字符串是否匹配过滤条件
func matchStringFilter(value string, condition filter.FilterCondition) bool {
	switch condition.Operator {
	case filter.OperatorEqual:
		return value == condition.Value.(string)
	case filter.OperatorNotEqual:
		return value != condition.Value.(string)
	case filter.OperatorContains:
		return strings.Contains(value, condition.Value.(string))
	case filter.OperatorStartsWith:
		return strings.HasPrefix(value, condition.Value.(string))
	case filter.OperatorEndsWith:
		return strings.HasSuffix(value, condition.Value.(string))
	default:
		return true
	}
}

// matchIntFilter 检查整数是否匹配过滤条件
func matchIntFilter(value int, condition filter.FilterCondition) bool {
	// 将字符串转换为整数
	filterValue, ok := condition.Value.(string)
	if !ok {
		return true
	}

	intValue, err := strconv.Atoi(filterValue)
	if err != nil {
		return true
	}

	switch condition.Operator {
	case filter.OperatorEqual:
		return value == intValue
	case filter.OperatorNotEqual:
		return value != intValue
	case filter.OperatorGreaterThan:
		return value > intValue
	case filter.OperatorGreaterThanEqual:
		return value >= intValue
	case filter.OperatorLessThan:
		return value < intValue
	case filter.OperatorLessThanEqual:
		return value <= intValue
	default:
		return true
	}
}

// matchesSearch 检查用户是否匹配搜索关键词
func matchesSearch(user User, search string) bool {
	if search == "" {
		return true
	}

	// 在各个字段中搜索关键词
	return strings.Contains(strings.ToLower(user.Name), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(user.Email), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(user.Company), strings.ToLower(search)) ||
		strings.Contains(strings.ToLower(user.Position), strings.ToLower(search))
}

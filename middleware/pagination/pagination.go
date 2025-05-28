// Package pagination 提供分页相关的中间件和工具
package pagination

import (
	"math"
	"strconv"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/gin-gonic/gin"
)

// 分页相关的常量
const (
	DefaultPage     = 1   // 默认页码
	DefaultPageSize = 10  // 默认每页数量
	MaxPageSize     = 100 // 最大每页数量
)

// 分页参数的键名
const (
	PageKey     = "page"      // 页码参数名
	PageSizeKey = "page_size" // 每页数量参数名
)

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`       // 当前页码
	PageSize   int `json:"pageSize"`   // 每页数量
	Total      int `json:"total"`      // 总记录数
	TotalPages int `json:"totalPages"` // 总页数
}

// Response 分页响应的标准格式
type Response struct {
	Data       interface{} `json:"data"`                 // 数据
	Pagination *Pagination `json:"pagination,omitempty"` // 分页信息
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total, pageSize int) int {
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// ExtractPaginationParams 从请求中提取分页参数
func ExtractPaginationParams(c *gin.Context) (page, pageSize int, err error) {
	// 获取页码参数
	pageStr := c.DefaultQuery(PageKey, strconv.Itoa(DefaultPage))
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, errors.NewValidationError("无效的页码参数", err).AddDetail(PageKey, "页码必须是大于等于1的整数")
	}

	// 获取每页数量参数
	pageSizeStr := c.DefaultQuery(PageSizeKey, strconv.Itoa(DefaultPageSize))
	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return 0, 0, errors.NewValidationError("无效的每页数量参数", err).AddDetail(PageSizeKey, "每页数量必须是大于等于1的整数")
	}

	// 限制每页最大数量
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return page, pageSize, nil
}

// SetPaginationHeader 设置分页相关的响应头
func SetPaginationHeader(c *gin.Context, pagination *Pagination) {
	c.Header("X-Total-Count", strconv.Itoa(pagination.Total))
	c.Header("X-Page", strconv.Itoa(pagination.Page))
	c.Header("X-Page-Size", strconv.Itoa(pagination.PageSize))
	c.Header("X-Total-Pages", strconv.Itoa(pagination.TotalPages))
}

// NewPagination 创建分页信息
func NewPagination(page, pageSize, total int) *Pagination {
	totalPages := CalculateTotalPages(total, pageSize)
	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// NewResponse 创建分页响应
func NewResponse(data interface{}, pagination *Pagination) *Response {
	return &Response{
		Data:       data,
		Pagination: pagination,
	}
}

// Middleware 分页中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取分页参数
		page, pageSize, err := ExtractPaginationParams(c)
		if err != nil {
			c.Error(err)
			return
		}

		// 将分页参数存储在上下文中
		c.Set(PageKey, page)
		c.Set(PageSizeKey, pageSize)

		c.Next()
	}
}

// GetPageParams 从上下文中获取分页参数
func GetPageParams(c *gin.Context) (page, pageSize int) {
	pageVal, exists := c.Get(PageKey)
	if !exists {
		page = DefaultPage
	} else {
		page = pageVal.(int)
	}

	pageSizeVal, exists := c.Get(PageSizeKey)
	if !exists {
		pageSize = DefaultPageSize
	} else {
		pageSize = pageSizeVal.(int)
	}

	return page, pageSize
}

// GetOffset 根据页码和每页数量计算偏移量
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// RespondWithPagination 返回带分页的响应
func RespondWithPagination(c *gin.Context, data interface{}, total int) {
	page, pageSize := GetPageParams(c)
	pagination := NewPagination(page, pageSize, total)

	// 设置分页相关的响应头
	SetPaginationHeader(c, pagination)

	// 返回响应
	c.JSON(200, NewResponse(data, pagination))
}

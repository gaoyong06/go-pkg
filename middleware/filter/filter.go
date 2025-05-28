// Package filter 提供过滤相关的中间件和工具
package filter

import (
	"strings"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/gin-gonic/gin"
)

// 过滤相关的常量
const (
	// FilterPrefix 过滤参数前缀
	FilterPrefix = "filter."
	// SortKey 排序参数名
	SortKey = "sort"
	// SearchKey 搜索参数名
	SearchKey = "q"
	// FieldsKey 字段选择参数名
	FieldsKey = "fields"
)

// 过滤操作符
const (
	// OperatorEqual 等于
	OperatorEqual = "eq"
	// OperatorNotEqual 不等于
	OperatorNotEqual = "ne"
	// OperatorGreaterThan 大于
	OperatorGreaterThan = "gt"
	// OperatorGreaterThanEqual 大于等于
	OperatorGreaterThanEqual = "gte"
	// OperatorLessThan 小于
	OperatorLessThan = "lt"
	// OperatorLessThanEqual 小于等于
	OperatorLessThanEqual = "lte"
	// OperatorIn 在集合中
	OperatorIn = "in"
	// OperatorNotIn 不在集合中
	OperatorNotIn = "nin"
	// OperatorLike 模糊匹配
	OperatorLike = "like"
	// OperatorContains 包含
	OperatorContains = "contains"
	// OperatorStartsWith 开头
	OperatorStartsWith = "startswith"
	// OperatorEndsWith 结尾
	OperatorEndsWith = "endswith"
	// OperatorIsNull 为空
	OperatorIsNull = "isnull"
	// OperatorIsNotNull 不为空
	OperatorIsNotNull = "isnotnull"
)

// FilterCondition 表示一个过滤条件
type FilterCondition struct {
	Field    string      // 字段名
	Operator string      // 操作符
	Value    interface{} // 值
}

// SortCondition 表示一个排序条件
type SortCondition struct {
	Field     string // 字段名
	Direction string // 排序方向，"asc" 或 "desc"
}

// FilterOptions 包含所有过滤选项
type FilterOptions struct {
	Filters []FilterCondition // 过滤条件
	Sorts   []SortCondition   // 排序条件
	Search  string            // 搜索关键词
	Fields  []string          // 要返回的字段
}

// 存储在上下文中的键
const (
	FilterOptionsKey = "filterOptions"
)

// Middleware 过滤中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取过滤参数
		filterOptions, err := ExtractFilterOptions(c)
		if err != nil {
			c.Error(err)
			return
		}

		// 将过滤选项存储在上下文中
		c.Set(FilterOptionsKey, filterOptions)

		c.Next()
	}
}

// ExtractFilterOptions 从请求中提取过滤选项
func ExtractFilterOptions(c *gin.Context) (*FilterOptions, error) {
	options := &FilterOptions{
		Filters: make([]FilterCondition, 0),
		Sorts:   make([]SortCondition, 0),
		Fields:  make([]string, 0),
	}

	// 提取搜索参数
	options.Search = c.Query(SearchKey)

	// 提取要返回的字段
	fieldsStr := c.Query(FieldsKey)
	if fieldsStr != "" {
		options.Fields = strings.Split(fieldsStr, ",")
	}

	// 提取排序参数
	sortStr := c.Query(SortKey)
	if sortStr != "" {
		sortFields := strings.Split(sortStr, ",")
		for _, field := range sortFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			direction := "asc"
			if strings.HasPrefix(field, "-") {
				direction = "desc"
				field = field[1:]
			}

			options.Sorts = append(options.Sorts, SortCondition{
				Field:     field,
				Direction: direction,
			})
		}
	}

	// 提取过滤参数
	for key, values := range c.Request.URL.Query() {
		if !strings.HasPrefix(key, FilterPrefix) || len(values) == 0 {
			continue
		}

		// 解析字段和操作符
		fieldOp := strings.TrimPrefix(key, FilterPrefix)
		parts := strings.Split(fieldOp, "_")

		var field, operator string
		if len(parts) == 1 {
			// 默认为等于操作符
			field = parts[0]
			operator = OperatorEqual
		} else if len(parts) == 2 {
			field = parts[0]
			operator = parts[1]
		} else {
			return nil, errors.NewValidationError(
				"无效的过滤参数格式",
				nil,
			).AddDetail(key, "过滤参数格式应为 filter.field 或 filter.field_operator")
		}

		// 检查操作符是否有效
		if !isValidOperator(operator) {
			return nil, errors.NewValidationError(
				"无效的过滤操作符",
				nil,
			).AddDetail(key, "不支持的操作符: "+operator)
		}

		// 处理特殊操作符
		value := values[0]
		var processedValue interface{} = value

		switch operator {
		case OperatorIn, OperatorNotIn:
			processedValue = strings.Split(value, ",")
		case OperatorIsNull, OperatorIsNotNull:
			processedValue = value == "true"
		}

		// 添加过滤条件
		options.Filters = append(options.Filters, FilterCondition{
			Field:    field,
			Operator: operator,
			Value:    processedValue,
		})
	}

	return options, nil
}

// isValidOperator 检查操作符是否有效
func isValidOperator(operator string) bool {
	validOperators := map[string]bool{
		OperatorEqual:            true,
		OperatorNotEqual:         true,
		OperatorGreaterThan:      true,
		OperatorGreaterThanEqual: true,
		OperatorLessThan:         true,
		OperatorLessThanEqual:    true,
		OperatorIn:               true,
		OperatorNotIn:            true,
		OperatorLike:             true,
		OperatorContains:         true,
		OperatorStartsWith:       true,
		OperatorEndsWith:         true,
		OperatorIsNull:           true,
		OperatorIsNotNull:        true,
	}

	_, ok := validOperators[operator]
	return ok
}

// GetFilterOptions 从上下文中获取过滤选项
func GetFilterOptions(c *gin.Context) *FilterOptions {
	val, exists := c.Get(FilterOptionsKey)
	if !exists {
		return &FilterOptions{
			Filters: make([]FilterCondition, 0),
			Sorts:   make([]SortCondition, 0),
			Fields:  make([]string, 0),
		}
	}

	return val.(*FilterOptions)
}

// BuildWhereClause 根据过滤条件构建 SQL WHERE 子句
func BuildWhereClause(filters []FilterCondition) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var clauses []string
	var args []interface{}

	for _, filter := range filters {
		switch filter.Operator {
		case OperatorEqual:
			clauses = append(clauses, filter.Field+" = ?")
			args = append(args, filter.Value)
		case OperatorNotEqual:
			clauses = append(clauses, filter.Field+" != ?")
			args = append(args, filter.Value)
		case OperatorGreaterThan:
			clauses = append(clauses, filter.Field+" > ?")
			args = append(args, filter.Value)
		case OperatorGreaterThanEqual:
			clauses = append(clauses, filter.Field+" >= ?")
			args = append(args, filter.Value)
		case OperatorLessThan:
			clauses = append(clauses, filter.Field+" < ?")
			args = append(args, filter.Value)
		case OperatorLessThanEqual:
			clauses = append(clauses, filter.Field+" <= ?")
			args = append(args, filter.Value)
		case OperatorIn:
			values := filter.Value.([]string)
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = "?"
				args = append(args, values[i])
			}
			clauses = append(clauses, filter.Field+" IN ("+strings.Join(placeholders, ",")+")")
		case OperatorNotIn:
			values := filter.Value.([]string)
			placeholders := make([]string, len(values))
			for i := range values {
				placeholders[i] = "?"
				args = append(args, values[i])
			}
			clauses = append(clauses, filter.Field+" NOT IN ("+strings.Join(placeholders, ",")+")")
		case OperatorLike:
			clauses = append(clauses, filter.Field+" LIKE ?")
			args = append(args, "%"+filter.Value.(string)+"%")
		case OperatorContains:
			clauses = append(clauses, filter.Field+" LIKE ?")
			args = append(args, "%"+filter.Value.(string)+"%")
		case OperatorStartsWith:
			clauses = append(clauses, filter.Field+" LIKE ?")
			args = append(args, filter.Value.(string)+"%")
		case OperatorEndsWith:
			clauses = append(clauses, filter.Field+" LIKE ?")
			args = append(args, "%"+filter.Value.(string))
		case OperatorIsNull:
			if filter.Value.(bool) {
				clauses = append(clauses, filter.Field+" IS NULL")
			} else {
				clauses = append(clauses, filter.Field+" IS NOT NULL")
			}
		case OperatorIsNotNull:
			if filter.Value.(bool) {
				clauses = append(clauses, filter.Field+" IS NOT NULL")
			} else {
				clauses = append(clauses, filter.Field+" IS NULL")
			}
		}
	}

	return strings.Join(clauses, " AND "), args
}

// BuildOrderByClause 根据排序条件构建 SQL ORDER BY 子句
func BuildOrderByClause(sorts []SortCondition) string {
	if len(sorts) == 0 {
		return ""
	}

	var clauses []string
	for _, sort := range sorts {
		clauses = append(clauses, sort.Field+" "+strings.ToUpper(sort.Direction))
	}

	return "ORDER BY " + strings.Join(clauses, ", ")
}

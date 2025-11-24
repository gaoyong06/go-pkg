// Package filter 提供过滤相关的中间件和工具
package filter

import (
	"context"
	"strings"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// KratosFilterKey 是 Kratos 上下文中存储过滤选项的键
const KratosFilterKey = "filter_options"

// KratosMiddleware 是一个 Kratos 中间件，用于处理过滤参数
func KratosMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从 HTTP 请求中提取过滤参数
			if tr, ok := http.RequestFromServerContext(ctx); ok {
				// 提取过滤选项
				filterOptions, err := ExtractFilterOptionsFromHTTP(tr)
				if err != nil {
					return nil, err
				}

				// 将过滤选项存储在上下文中
				ctx = context.WithValue(ctx, KratosFilterKey, filterOptions)
			}

			// 调用下一个处理器
			return handler(ctx, req)
		}
	}
}

// ExtractFilterOptionsFromHTTP 从 HTTP 请求中提取过滤选项
func ExtractFilterOptionsFromHTTP(req *http.Request) (*FilterOptions, error) {
	options := &FilterOptions{
		Filters: make([]FilterCondition, 0),
		Sorts:   make([]SortCondition, 0),
		Fields:  make([]string, 0),
	}

	// 获取查询参数
	query := req.URL.Query()

	// 提取搜索参数
	options.Search = query.Get(SearchKey)

	// 提取要返回的字段
	fieldsStr := query.Get(FieldsKey)
	if fieldsStr != "" {
		options.Fields = strings.Split(fieldsStr, ",")
	}

	// 提取排序参数
	sortStr := query.Get(SortKey)
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
	for key, values := range query {
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

		// 处理值
		value := values[0]
		var processedValue interface{} = value

		// 处理特殊操作符
		if operator == OperatorIn || operator == OperatorNotIn {
			processedValue = strings.Split(value, ",")
		} else if operator == OperatorIsNull || operator == OperatorIsNotNull {
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

// GetFilterOptionsFromContext 从 Kratos 上下文中获取过滤选项
func GetFilterOptionsFromContext(ctx context.Context) *FilterOptions {
	val := ctx.Value(KratosFilterKey)
	if val == nil {
		return &FilterOptions{
			Filters: make([]FilterCondition, 0),
			Sorts:   make([]SortCondition, 0),
			Fields:  make([]string, 0),
		}
	}

	return val.(*FilterOptions)
}

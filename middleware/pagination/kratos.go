// Package pagination u63d0u4f9bu5206u9875u76f8u5173u7684u4e2du95f4u4ef6u548cu5de5u5177
package pagination

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// u5206u9875u76f8u5173u7684u5e38u91cf
const (
	// KratosPaginationKey u662f Kratos u4e0au4e0bu6587u4e2du5b58u50a8u5206u9875u4fe1u606fu7684u952e
	KratosPaginationKey = "pagination_info"
	// DefaultPage u9ed8u8ba4u9875u7801
	DefaultPage = 1
	// DefaultPageSize u9ed8u8ba4u6bcfu9875u6570u91cf
	DefaultPageSize = 10
	// MaxPageSize u6700u5927u6bcfu9875u6570u91cf
	MaxPageSize = 100
)

// PaginationInfo u5305u542bu5206u9875u4fe1u606f
type PaginationInfo struct {
	Page     int `json:"page"`      // u5f53u524du9875u7801
	PageSize int `json:"page_size"` // u6bcfu9875u6570u91cf
}

// KratosMiddleware u662fu4e00u4e2a Kratos u4e2du95f4u4ef6uff0cu7528u4e8eu5904u7406u5206u9875u53c2u6570
func KratosMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// u4ece HTTP u8bf7u6c42u4e2du63d0u53d6u5206u9875u53c2u6570
			if tr, ok := http.RequestFromServerContext(ctx); ok {
				query := tr.URL.Query()

				// u89e3u6790u9875u7801
				page := DefaultPage
				if pageStr := query.Get("page"); pageStr != "" {
					if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
						page = p
					}
				}

				// u89e3u6790u6bcfu9875u6570u91cf
				pageSize := DefaultPageSize
				if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
					if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
						pageSize = ps
					}
				}

				// u9650u5236u6bcfu9875u6700u5927u6570u91cf
				if pageSize > MaxPageSize {
					pageSize = MaxPageSize
				}

				// u5c06u5206u9875u4fe1u606fu5b58u50a8u5728u4e0au4e0bu6587u4e2d
				ctx = context.WithValue(ctx, KratosPaginationKey, &PaginationInfo{
					Page:     page,
					PageSize: pageSize,
				})
			}

			// u8c03u7528u4e0bu4e00u4e2au5904u7406u5668
			return handler(ctx, req)
		}
	}
}

// GetPaginationFromContext u4ece Kratos u4e0au4e0bu6587u4e2du83b7u53d6u5206u9875u4fe1u606f
func GetPaginationFromContext(ctx context.Context) *PaginationInfo {
	val := ctx.Value(KratosPaginationKey)
	if val == nil {
		return &PaginationInfo{
			Page:     DefaultPage,
			PageSize: DefaultPageSize,
		}
	}

	return val.(*PaginationInfo)
}

// GetOffset u6839u636eu9875u7801u548cu6bcfu9875u6570u91cfu8ba1u7b97u504fu79fbu91cf
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// PaginationResult u5206u9875u7ed3u679c
type PaginationResult struct {
	Items      interface{} `json:"items"`       // u5f53u524du9875u7684u6570u636eu9879
	Total      int         `json:"total"`       // u603bu6570u636eu91cf
	Page       int         `json:"page"`        // u5f53u524du9875u7801
	PageSize   int         `json:"page_size"`   // u6bcfu9875u6570u91cf
	TotalPages int         `json:"total_pages"` // u603bu9875u6570
}

// NewPaginationResult u521bu5efau5206u9875u7ed3u679c
func NewPaginationResult(items interface{}, total, page, pageSize int) *PaginationResult {
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &PaginationResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

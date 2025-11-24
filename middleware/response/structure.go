// Package response 提供统一响应格式中间件
package response

// ResponseStructure 统一API响应格式
type ResponseStructure struct {
	Success      bool        `json:"success"`      // 请求是否成功
	Data         interface{} `json:"data"`         // 返回数据（成功时）
	ErrorCode    string      `json:"errorCode"`    // 错误代码
	ErrorMessage string      `json:"errorMessage"` // 错误信息
	ShowType     int         `json:"showType"`     // 错误展示类型
	TraceId      string      `json:"traceId"`      // 请求追踪ID
	Host         string      `json:"host"`         // 请求的主机信息
}

// ShowType 定义错误提示类型常量
// SILENT = 0, WARN_MESSAGE = 1, ERROR_MESSAGE = 2, NOTIFICATION = 3, REDIRECT = 9
const (
	// ShowTypeSilent 不提示错误
	ShowTypeSilent = 0
	// ShowTypeWarnMessage 警告信息提示
	ShowTypeWarnMessage = 1
	// ShowTypeErrorMessage 错误信息提示
	ShowTypeErrorMessage = 2
	// ShowTypeNotification 通知提示
	ShowTypeNotification = 3
	// ShowTypeRedirect 页面跳转
	ShowTypeRedirect = 9
)


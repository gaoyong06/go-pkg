// Package errors 提供通用错误码定义
package errors

// 错误码设计规范：
// 格式：SSMMEE (6位数字)
//   SS: 服务标识 (10-99)，每个服务分配一个唯一标识，最多支持 90 个服务
//   MM: 模块标识 (00-99)，每个服务内部按模块划分
//   EE: 模块内部错误序号 (00-99)
//
// 示例：
//   100001 = 10(通用服务) + 00(参数校验模块) + 01(参数无效)
//   110101 = 11(passport 服务) + 01(登录模块) + 01(密码错误)
//
// 设计原则：
//   1. 每个服务最多 100 个模块，每个模块最多 100 个错误码，单服务容量 10,000
//   2. 编号紧凑易读，方便在日志、告警和文档中快速定位
//   3. 扩展模块时优先复用剩余编号，不鼓励一次性预留过大的区间
//
// 服务标识分配（持续补充）：
//   10: 通用服务 (common-service)
//   11: 用户认证服务 (passport-service)
//   12: 支付服务 (payment-service)
//   13: 订阅服务 (subscription-service)
//   14: 通知服务 (notification-service)
//   15: 桌位排期服务 (table-plan-service)
//   17: API Key 服务 (api-key-service)
//   16, 18-99: 预留其他服务
//
// 模块标识建议：
//   00: 参数校验、通用校验
//   01-09: 通用模块（鉴权、系统、资源、业务等）
//   10-49: 服务核心业务模块
//   50-79: 扩展模块
//   80-99: 预留扩展

// 通用错误码 (服务标识 10: 100000-109999)
// 这些错误码适用于所有项目，定义在公共库中
const (
	// 通用模块 - 参数验证 (100000-100099)
	// ErrCodeInvalidArgument 无效参数错误
	ErrCodeInvalidArgument = 100001
	// ErrCodeMissingRequiredField 缺少必填字段
	ErrCodeMissingRequiredField = 100002
	// ErrCodeInvalidFormat 格式错误
	ErrCodeInvalidFormat = 100003
	// ErrCodeOutOfRange 参数超出范围
	ErrCodeOutOfRange = 100004

	// 通用模块 - 权限相关 (100100-100199)
	// ErrCodeUnauthorized 未授权错误
	ErrCodeUnauthorized = 100101
	// ErrCodeForbidden 禁止访问错误
	ErrCodeForbidden = 100102
	// ErrCodeTokenExpired Token已过期
	ErrCodeTokenExpired = 100103
	// ErrCodeTokenInvalid Token无效
	ErrCodeTokenInvalid = 100104

	// 通用模块 - 系统错误 (100200-100299)
	// ErrCodeInternalError 内部错误
	ErrCodeInternalError = 100201
	// ErrCodeServiceUnavailable 服务不可用
	ErrCodeServiceUnavailable = 100202
	// ErrCodeTimeout 请求超时
	ErrCodeTimeout = 100203
	// ErrCodeDatabaseError 数据库错误
	ErrCodeDatabaseError = 100204
	// ErrCodeExternalServiceError 外部服务错误
	ErrCodeExternalServiceError = 100205

	// 通用模块 - 资源相关 (100300-100399)
	// ErrCodeNotFound 资源不存在
	ErrCodeNotFound = 100301
	// ErrCodeAlreadyExists 资源已存在
	ErrCodeAlreadyExists = 100302
	// ErrCodeResourceExhausted 资源耗尽
	ErrCodeResourceExhausted = 100303

	// 通用模块 - 业务逻辑 (100400-100499)
	// ErrCodeOperationNotAllowed 操作不允许
	ErrCodeOperationNotAllowed = 100401
	// ErrCodeBusinessRuleViolation 违反业务规则
	ErrCodeBusinessRuleViolation = 100402
	// ErrCodeInsufficientBalance 余额不足
	ErrCodeInsufficientBalance = 100403

	// ========== API Key 服务错误码 (服务标识 17: 170000-179999) ==========
	// API Key 管理模块 (170000-170099)
	// ErrCodeApiKeyAlreadyExists 用户已存在活跃的 API Key
	ErrCodeApiKeyAlreadyExists = 170001
	// ErrCodeApiKeyGenerateFailed 生成 API Key 失败
	ErrCodeApiKeyGenerateFailed = 170002
	// ErrCodeApiKeyCreateFailed 创建 API Key 失败
	ErrCodeApiKeyCreateFailed = 170003
	// ErrCodeApiKeyNotFound 未找到活跃的 API Key
	ErrCodeApiKeyNotFound = 170004
	// ErrCodeApiKeyNotExists API Key 不存在
	ErrCodeApiKeyNotExists = 170005
	// ErrCodeApiKeyDeleteFailed 删除 API Key 失败
	ErrCodeApiKeyDeleteFailed = 170006
	// ErrCodeApiKeyInvalid 无效或已停用的 API Key
	ErrCodeApiKeyInvalid = 170007
	// ErrCodeApiKeyCheckFailed 检查现有 API Key 失败
	ErrCodeApiKeyCheckFailed = 170008
)

// 服务错误码范围示例（SS 对应服务标识，MM 根据实际模块划分）：
//   用户认证服务 (passport-service, SS=11)
//     110000-110099: 通用模块
//     110100-110199: 登录模块
//     110200-110299: 注册模块
//     ...
//   支付服务 (payment-service, SS=12)
//     120100-120199: 支付单模块
//     120200-120299: 退款模块
//   其他服务以此类推，通过 SS 对齐服务、MM 对齐模块、EE 对齐错误序号

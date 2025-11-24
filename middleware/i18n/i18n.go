// Package i18n 提供国际化（i18n）工具函数
package i18n

import (
	"context"
)

// Translator 翻译器接口
type Translator interface {
	// Translate 翻译文本
	// key: 翻译键，如 "seating.operations.algorithm"
	// templateData: 模板数据（可选）
	Translate(ctx context.Context, key string, templateData map[string]interface{}) string
}

// DefaultTranslator 默认翻译器实现
// 如果找不到翻译，返回 key 本身
type DefaultTranslator struct{}

// NewDefaultTranslator 创建默认翻译器
func NewDefaultTranslator() *DefaultTranslator {
	return &DefaultTranslator{}
}

// Translate 实现 Translator 接口
func (t *DefaultTranslator) Translate(ctx context.Context, key string, templateData map[string]interface{}) string {
	// 默认实现：返回 key 本身
	// 实际项目中应该使用 go-i18n 或其他翻译库
	return key
}


// Package utils 提供通用工具函数
package utils

import "regexp"

// IsValidPhone 验证手机号格式
func IsValidPhone(phone string) bool {
	// 简单的手机号验证：11位数字，以1开头
	matched, _ := regexp.MatchString(`^1\d{10}$`, phone)
	return matched
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	// 简单的邮箱验证
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// GetStringValue 安全地获取字符串指针的值
func GetStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}


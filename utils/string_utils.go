// Package utils 提供通用工具函数
package utils

import (
	"regexp"
	"strings"
)

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

// MaskPhone 对手机号进行脱敏处理
// 显示前3位和后4位，中间用 **** 替代
// 例如：18600024912 -> 186****4912
// 如果手机号长度不足7位，全部用 * 替代（保护隐私）
func MaskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	length := len(phone)
	if length <= 7 {
		// 如果手机号长度不足7位，全部用 * 替代
		return strings.Repeat("*", length)
	}
	// 显示前3位和后4位
	return phone[:3] + "****" + phone[length-4:]
}

// MaskPhoneWithCustomMask 使用自定义掩码对手机号进行脱敏处理
// mask: 自定义掩码字符，默认为****
// 显示前3位和后4位，中间用自定义掩码替代
func MaskPhoneWithCustomMask(phone, mask string) string {
	if phone == "" {
		return ""
	}

	// 如果未指定掩码，使用默认的****
	if mask == "" {
		mask = "****"
	}

	length := len(phone)
	if length <= 7 {
		// 如果手机号长度不足7位，全部用 * 替代
		return strings.Repeat("*", length)
	}
	// 显示前3位和后4位
	return phone[:3] + mask + phone[length-4:]
}

// MaskEmail 对邮箱进行脱敏处理
// 显示 @ 前面的前2位和 @ 后面的域名，中间用 **** 替代
// 例如：gaoyong06@qq.com -> ga****@qq.com
// 如果用户名长度不足2位，全部用 * 替代（保护隐私）
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		// 如果格式不正确，全部用 * 替代
		return strings.Repeat("*", len(email))
	}
	username := parts[0]
	domain := parts[1]
	if len(username) <= 2 {
		// 如果用户名长度不足2位，全部用 * 替代
		return strings.Repeat("*", len(username)) + "@" + domain
	}
	// 显示前2位，后面用 **** 替代
	return username[:2] + "****@" + domain
}

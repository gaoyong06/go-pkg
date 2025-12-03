// Package utils 提供通用工具函数
package utils

// MaskPhone 对手机号进行脱敏处理
// 保留前3位和后4位，中间用****替换
// 例如：13812345678 -> 138****5678
// 对于7位数字：1234567 -> 123****567
func MaskPhone(phone string) string {
	if phone == "" {
		return phone
	}

	// 如果手机号长度小于7位，直接返回原值
	if len(phone) < 7 {
		return phone
	}

	// 对于7位数字，保留前3位和后3位
	if len(phone) == 7 {
		prefix := phone[:3]
		suffix := phone[len(phone)-3:]
		return prefix + "****" + suffix
	}

	// 对于8位及以上，保留前3位和后4位
	prefix := phone[:3]
	suffix := phone[len(phone)-4:]

	return prefix + "****" + suffix
}

// MaskPhoneWithCustomMask 使用自定义掩码对手机号进行脱敏处理
// mask: 自定义掩码字符，默认为****
func MaskPhoneWithCustomMask(phone, mask string) string {
	if phone == "" {
		return phone
	}

	// 如果手机号长度小于7位，直接返回原值
	if len(phone) < 7 {
		return phone
	}

	// 如果未指定掩码，使用默认的****
	if mask == "" {
		mask = "****"
	}

	// 对于7位数字，保留前3位和后3位
	if len(phone) == 7 {
		prefix := phone[:3]
		suffix := phone[len(phone)-3:]
		return prefix + mask + suffix
	}

	// 对于8位及以上，保留前3位和后4位
	prefix := phone[:3]
	suffix := phone[len(phone)-4:]

	return prefix + mask + suffix
}


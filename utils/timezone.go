// Package utils 提供通用工具函数
package utils

import (
	"time"
)

// 默认时区常量
const DefaultTimezone = "UTC"

// GetDefaultTimezone 获取默认时区
func GetDefaultTimezone() string {
	return DefaultTimezone
}

// ConvertToUTC 将指定时区的时间转换为UTC时间
func ConvertToUTC(t time.Time, timezone string) time.Time {
	if timezone == DefaultTimezone {
		return t
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t
	}

	return t.In(loc).UTC()
}

// ConvertFromUTC 将UTC时间转换为指定时区的时间
func ConvertFromUTC(t time.Time, timezone string) time.Time {
	if timezone == DefaultTimezone {
		return t
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t
	}

	return t.In(loc)
}

// ParseTimeString 解析时间字符串
func ParseTimeString(timeStr, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	return time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc)
}

// FormatTimeResponse 格式化时间响应
func FormatTimeResponse(t time.Time, timezone string) string {
	if timezone == DefaultTimezone {
		return t.Format("2006-01-02 15:04:05")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t.Format("2006-01-02 15:04:05")
	}

	return t.In(loc).Format("2006-01-02 15:04:05")
}

// IsValidTimezone 检查时区是否有效
func IsValidTimezone(timezone string) bool {
	if timezone == DefaultTimezone {
		return true
	}

	_, err := time.LoadLocation(timezone)
	return err == nil
}

// GetCurrentUTC 获取当前UTC时间
func GetCurrentUTC() time.Time {
	return time.Now().UTC()
}

// FormatUTC 格式化UTC时间
func FormatUTC(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}


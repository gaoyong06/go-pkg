package utils

import (
	"fmt"
	"time"
)

// ParseDateRange 解析日期范围字符串（YYYY-MM-DD 格式）
func ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format: %w", err)
	}

	// 设置结束时间为当天的 23:59:59
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())

	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date must be before end_date")
	}

	return startTime, endTime, nil
}

// GetPreviousPeriod 获取上一周期的时间范围
func GetPreviousPeriod(startTime, endTime time.Time) (time.Time, time.Time) {
	duration := endTime.Sub(startTime)
	prevEndTime := startTime.Add(-time.Second) // 上一周期的结束时间是当前周期的开始时间减1秒
	prevStartTime := prevEndTime.Add(-duration)
	return prevStartTime, prevEndTime
}

// GetLastNDays 获取最近 N 天的日期范围
func GetLastNDays(n int) (time.Time, time.Time) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -n)
	return startTime, endTime
}

// GetLastNMonths 获取最近 N 个月的日期范围
func GetLastNMonths(n int) (time.Time, time.Time) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, -n, 0)
	return startTime, endTime
}


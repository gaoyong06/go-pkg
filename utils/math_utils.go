package utils

// CalculateChange 计算变化百分比
func CalculateChange(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0.0
		}
		return 100.0 // 从0增长到非0，视为100%增长
	}
	return float64(current-previous) / float64(previous) * 100.0
}

// CalculateROI 计算 ROI（投资回报率）
func CalculateROI(revenue, cost float64) float64 {
	if cost == 0 {
		if revenue > 0 {
			return 999999.0 // 成本为0，收入大于0，ROI为无穷大
		}
		return 0.0
	}
	return (revenue - cost) / cost * 100.0
}

// CalculateConversionRate 计算转化率
func CalculateConversionRate(conversions, visitors int64) float64 {
	if visitors == 0 {
		return 0.0
	}
	return float64(conversions) / float64(visitors) * 100.0
}


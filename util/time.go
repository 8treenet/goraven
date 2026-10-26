package util

import "time"

// StartOfToday 返回今天零点的 time.Time（本地时区）
func StartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Millisecond returns the current Unix timestamp in milliseconds.
func Millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GenLastNDates 生成最近 N 天的日期字符串列表（YYYY-MM-DD），含今天
// 按时间从早到晚排列，例如 GenLastNDates(3) → ["2026-05-11", "2026-05-12", "2026-05-13"]
func GenLastNDates(n int) []string {
	today := StartOfToday()

	result := make([]string, n)
	for i := 0; i < n; i++ {
		d := today.AddDate(0, 0, -(n - 1 - i))
		result[i] = d.Format("2006-01-02")
	}
	return result
}

// GenThisWeekDates 生成本周（周一至周日）的日期字符串列表
func GenThisWeekDates() []string {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	today := StartOfToday()
	monday := today.AddDate(0, 0, -int(weekday)+1)

	result := make([]string, 7)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		result[i] = d.Format("2006-01-02")
	}
	return result
}

package util

import "time"

func StartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func Millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GenLastNDates(n int) []string {
	today := StartOfToday()

	result := make([]string, n)
	for i := 0; i < n; i++ {
		d := today.AddDate(0, 0, -(n - 1 - i))
		result[i] = d.Format("2006-01-02")
	}
	return result
}

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

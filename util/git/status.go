package git

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ParseStatusZ 解析 `git status --porcelain -z -uall` 输出。
func ParseStatusZ(raw string) []Change {
	parts := strings.Split(raw, "\x00")
	changes := make([]Change, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		entry := parts[i]
		if len(entry) < 4 {
			continue
		}
		xy := entry[:2]
		change := Change{Path: entry[3:], Status: statusLetter(xy)}
		if change.Status == "R" || change.Status == "C" {
			if i+1 < len(parts) {
				change.Source = parts[i+1]
				i++
			}
		}
		changes = append(changes, change)
	}
	return changes
}

// statusLetter 将 porcelain 的 XY 码映射为单字母角标。
func statusLetter(xy string) string {
	if xy == "??" {
		return "U"
	}
	x, y := xy[0], xy[1]
	unmerged := func(c byte) bool { return c == 'U' }
	if unmerged(x) || unmerged(y) || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
		return "C"
	}
	switch {
	case x == 'R' || y == 'R':
		return "R"
	case x == 'A':
		return "A"
	case x == 'D' || y == 'D':
		return "D"
	case x == 'M' || y == 'M':
		return "M"
	}
	return "M"
}

// FilterLargeChanges 按阈值过滤大文件，超限文件不参与提交并作为 skipped 返回。
func FilterLargeChanges(dir string, changes []Change, limit int64) (keep, skipped []Change) {
	for _, change := range changes {
		if change.Status == "D" {
			keep = append(keep, change)
			continue
		}
		info, err := os.Stat(filepath.Join(dir, change.Path))
		if err != nil || info.IsDir() {
			keep = append(keep, change)
			continue
		}
		change.Size = info.Size()
		if info.Size() > limit {
			skipped = append(skipped, change)
			continue
		}
		keep = append(keep, change)
	}
	return keep, skipped
}

// NextDailySyncAt 返回严格晚于 now 的最近一次 03:00（服务器本地时区）。
func NextDailySyncAt(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// LogFormat 解析 ParseLog 输出所需的 `git log` pretty 格式串。
const LogFormat = "%H%x1f%h%x1f%an%x1f%ae%x1f%at%x1f%s%x1e"

// ParseLog 解析按 LogFormat 输出的 `git log` 结果。
func ParseLog(raw string) []CommitInfo {
	items := make([]CommitInfo, 0)
	for _, record := range strings.Split(raw, "\x1e") {
		record = strings.Trim(record, "\n")
		if record == "" {
			continue
		}
		fields := strings.Split(record, "\x1f")
		if len(fields) < 6 {
			continue
		}
		ts, _ := strconv.ParseInt(fields[4], 10, 64)
		items = append(items, CommitInfo{
			Hash:      fields[0],
			ShortHash: fields[1],
			Author:    fields[2],
			Email:     fields[3],
			Time:      time.Unix(ts, 0),
			Message:   fields[5],
		})
	}
	return items
}

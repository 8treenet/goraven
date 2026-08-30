package repository_test

import (
	"goraven/backend/po"
	"goraven/backend/repository"
	"testing"
	"time"
)

// nextMonday 构造一个周一固定时刻（本地时区）
func nextMonday(base time.Time, hour, min int) time.Time {
	t := base
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, 1)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, base.Location())
}

func TestCalcNextRunAtOnce(t *testing.T) {
	runAt := time.Date(2026, 8, 28, 10, 30, 0, 0, time.Local)
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeOnce, RunAt: &runAt}

	got, err := repository.CalcNextRunAt(task, time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	if !got.Equal(runAt) {
		t.Fatalf("单次任务 NextRunAt 应等于 RunAt %v，得到 %v", runAt, got)
	}
}

func TestCalcNextRunAtOnceMissingRunAt(t *testing.T) {
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeOnce}
	if _, err := repository.CalcNextRunAt(task, time.Now()); err == nil {
		t.Fatal("单次任务缺少 RunAt 应返回错误")
	}
}

func TestCalcNextRunAtInterval(t *testing.T) {
	now := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 90}

	got, err := repository.CalcNextRunAt(task, now)
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	want := now.Add(90 * time.Minute)
	if !got.Equal(want) {
		t.Fatalf("间隔任务应为 %v，得到 %v", want, got)
	}
}

func TestCalcNextRunAtIntervalInvalid(t *testing.T) {
	now := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 4}
	if _, err := repository.CalcNextRunAt(task, now); err == nil {
		t.Fatal("间隔分钟数 <5 应返回错误")
	}
}

func TestCalcNextRunAtDaily(t *testing.T) {
	// 今天该时刻未到 → 今天
	base := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeDaily, FixedTime: "21:00"}

	got, err := repository.CalcNextRunAt(task, base)
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	want := time.Date(2026, 8, 27, 21, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("今天未过点应为 %v，得到 %v", want, got)
	}

	// 已过 → 明天
	past := time.Date(2026, 8, 27, 22, 0, 0, 0, time.Local)
	got, err = repository.CalcNextRunAt(task, past)
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	want = time.Date(2026, 8, 28, 21, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("已过点应推到明天 %v，得到 %v", want, got)
	}

	// 恰好等于当前时刻 → 严格未来，推到明天
	exact := time.Date(2026, 8, 27, 21, 0, 0, 0, time.Local)
	got, _ = repository.CalcNextRunAt(task, exact)
	if got.Before(exact) || got.Equal(exact) {
		t.Fatalf("结果必须严格晚于当前时刻：now=%v got=%v", exact, got)
	}
}

// TestCalcNextRunAtDailyLenientParse 一位小时的 HH:MM 输入（如 "9:30"）按合法时间解析
func TestCalcNextRunAtDailyLenientParse(t *testing.T) {
	base := time.Date(2026, 8, 27, 5, 0, 0, 0, time.Local)
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeDaily, FixedTime: "9:30"}
	got, err := repository.CalcNextRunAt(task, base)
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	want := time.Date(2026, 8, 27, 9, 30, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("宽松解析应为 %v，得到 %v", want, got)
	}
}

func TestCalcNextRunAtDailyInvalid(t *testing.T) {
	base := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	cases := []struct {
		name string
		task *po.AutomationTask
	}{
		{"空 FixedTime", &po.AutomationTask{ExecType: po.AutomationExecTypeDaily}},
		{"非法格式", &po.AutomationTask{ExecType: po.AutomationExecTypeDaily, FixedTime: "25:99"}},
		{"缺失分钟段", &po.AutomationTask{ExecType: po.AutomationExecTypeDaily, FixedTime: "2130"}},
	}
	for _, c := range cases {
		if _, err := repository.CalcNextRunAt(c.task, base); err == nil {
			t.Fatalf("%s 应返回错误", c.name)
		}
	}
}

func TestCalcNextRunAtWeekly(t *testing.T) {
	// 周四（time.Thursday=4）15:00 为 now；目标是下周一 09:00
	thu := time.Date(2026, 8, 27, 15, 0, 0, 0, time.Local)
	if thu.Weekday() != time.Thursday {
		t.Fatalf("测试前置失败：2026-08-27 应为周四，实际 %v", thu.Weekday())
	}
	wantDay := nextMonday(thu, 9, 0)

	var monday uint8 = 1
	task := &po.AutomationTask{ExecType: po.AutomationExecTypeWeekly, FixedTime: "09:00", Weekday: monday}
	got, err := repository.CalcNextRunAt(task, thu)
	if err != nil {
		t.Fatalf("CalcNextRunAt 返回错误: %v", err)
	}
	if !got.Equal(wantDay) {
		t.Fatalf("跨周推演应为 %v，得到 %v", wantDay, got)
	}

	// 今天就是目标日且时刻未到 → 本周今天
	var thursday uint8 = uint8(time.Thursday)
	sameDayTarget := time.Date(thu.Year(), thu.Month(), thu.Day(), 18, 0, 0, 0, time.Local)
	task.Weekday = thursday
	task.FixedTime = "21:00"
	got, _ = repository.CalcNextRunAt(task, sameDayTarget)
	want := time.Date(sameDayTarget.Year(), sameDayTarget.Month(), sameDayTarget.Day(), 21, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("目标日本过点应为本周 %v，得到 %v", want, got)
	}

	// 目标日但时刻已过（22:00 晚于目标 21:00）→ 下周同刻
	afterTarget := time.Date(thu.Year(), thu.Month(), thu.Day(), 22, 0, 0, 0, time.Local)
	got, _ = repository.CalcNextRunAt(task, afterTarget)
	nextWeekThu := time.Date(thu.Year(), thu.Month(), thu.Day()+7, 21, 0, 0, 0, time.Local)
	if !got.Equal(nextWeekThu) {
		t.Fatalf("目标日已过点应推到下周 %v，得到 %v", nextWeekThu, got)
	}

	// Weekday 越界
	task.Weekday = 7
	if _, err := repository.CalcNextRunAt(task, thu); err == nil {
		t.Fatal("Weekday=7 应返回错误")
	}
}

func TestCalcNextRunAtUnknownExecType(t *testing.T) {
	task := &po.AutomationTask{ExecType: 9}
	if _, err := repository.CalcNextRunAt(task, time.Now()); err == nil {
		t.Fatal("未知执行类型应返回错误")
	}
}

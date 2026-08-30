package tools_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/core/iface"
	"goraven/core/tools"

	"github.com/cloudwego/eino/components/tool"
)

// updateFakeRepo 捕获更新调用，模拟仓储行为
type updateFakeRepo struct {
	stored       *po.AutomationTask
	getErr       error
	updateErr    error
	gotRecompute bool
	updated      *po.AutomationTask
}

func (f *updateFakeRepo) CreateTask(task *po.AutomationTask) error { return nil }

func (f *updateFakeRepo) ListEnabledTasks(userId, title string, page, pageSize int) ([]po.AutomationTask, int64, error) {
	return nil, 0, nil
}

func (f *updateFakeRepo) GetTask(id int, userId string) (*po.AutomationTask, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.stored, nil
}

func (f *updateFakeRepo) UpdateTask(task *po.AutomationTask, recomputeNext bool) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.gotRecompute = recomputeNext
	cp := *task
	f.updated = &cp
	return nil
}

func newUpdateTool(t *testing.T, repo iface.AutomationTaskRepo) tool.InvokableTool {
	t.Helper()
	invokable, err := tools.NewAutomationTaskUpdater("user-001", repo)
	if err != nil {
		t.Fatal(err)
	}
	return invokable
}

func invokeUpdate(t *testing.T, tt tool.InvokableTool, req tools.UpdateAutomationTaskRequest) string {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tt.InvokableRun(context.Background(), string(data))
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	return raw
}

func TestUpdateAutomationTaskScheduleChangedRecomputes(t *testing.T) {
	next := time.Date(2026, 9, 1, 8, 30, 0, 0, time.Local)
	repo := &updateFakeRepo{stored: &po.AutomationTask{
		Id: 5, Title: "旧标题", Requirement: "旧需求", UserId: "user-001",
		ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 30, NextRunAt: next,
		Status: po.AutomationStatusEnabled,
	}}
	tt := newUpdateTool(t, repo)

	raw := invokeUpdate(t, tt, tools.UpdateAutomationTaskRequest{
		TaskId:      5,
		Title:       "每天巡检",
		Requirement: "每天早上整理 downloads 目录",
		ExecType:    po.AutomationExecTypeDaily,
		FixedTime:   "09:30",
	})

	if !repo.gotRecompute {
		t.Fatal("执行计划变更应触发 NextRunAt 重算")
	}
	if repo.updated.ExecType != po.AutomationExecTypeDaily || repo.updated.FixedTime != "09:30" {
		t.Fatalf("计划字段应替换为每日 09:30，得到 %+v", repo.updated)
	}
	// 旧类型的字段应被归一化清零
	if repo.updated.IntervalMinutes != 0 {
		t.Fatalf("切换类型后 interval_minutes 应归零，得到 %d", repo.updated.IntervalMinutes)
	}
	// 暂存字段与状态不因更新而变化
	if repo.updated.Status != po.AutomationStatusEnabled {
		t.Fatalf("状态不应被更新工具修改，得到 %d", repo.updated.Status)
	}
	for _, key := range []string{`"task_id":5`, `"title":"每天巡检"`} {
		if !strings.Contains(raw, key) {
			t.Fatalf("响应缺少 %s：%s", key, raw)
		}
	}
}

func TestUpdateAutomationTaskRequirementOnlyKeepsNextRunAt(t *testing.T) {
	next := time.Date(2026, 9, 1, 8, 30, 0, 0, time.Local)
	repo := &updateFakeRepo{stored: &po.AutomationTask{
		Id: 5, Title: "每日备份", Requirement: "旧需求", UserId: "user-001",
		ExecType: po.AutomationExecTypeDaily, FixedTime: "09:30", NextRunAt: next,
		Status: po.AutomationStatusEnabled,
	}}
	tt := newUpdateTool(t, repo)

	raw := invokeUpdate(t, tt, tools.UpdateAutomationTaskRequest{
		TaskId:      5,
		Title:       "每日备份",
		Requirement: "新需求：整理后附加生成周报",
		ExecType:    po.AutomationExecTypeDaily,
		FixedTime:   "09:30",
	})

	if repo.gotRecompute {
		t.Fatal("仅改标题/需求不应重算 NextRunAt")
	}
	if !repo.updated.NextRunAt.Equal(next) {
		t.Fatalf("NextRunAt 应保持原值 %v，得到 %v", next, repo.updated.NextRunAt)
	}
	if !strings.Contains(raw, `"next_run_at":"2026-09-01 08:30"`) {
		t.Fatalf("响应应回传原下次执行时间：%s", raw)
	}
}

func TestUpdateAutomationTaskDoneRejected(t *testing.T) {
	repo := &updateFakeRepo{stored: &po.AutomationTask{
		Id: 5, Title: "t", Requirement: "r", UserId: "user-001",
		ExecType: po.AutomationExecTypeOnce, Status: po.AutomationStatusDone,
	}}
	tt := newUpdateTool(t, repo)

	data, _ := json.Marshal(tools.UpdateAutomationTaskRequest{
		TaskId: 5, Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeOnce,
		RunAt: time.Now().Add(time.Hour).Format("2006-01-02 15:04"),
	})
	_, err := tt.InvokableRun(context.Background(), string(data))
	if err == nil || !strings.Contains(err.Error(), "已完成") {
		t.Fatalf("已完成任务应拒绝修改，得到 %v", err)
	}
	if repo.updated != nil {
		t.Fatal("已完成任务不应触发写库")
	}
}

func TestUpdateAutomationTaskDisabledAllowed(t *testing.T) {
	repo := &updateFakeRepo{stored: &po.AutomationTask{
		Id: 5, Title: "t", Requirement: "r", UserId: "user-001",
		ExecType: po.AutomationExecTypeDaily, FixedTime: "09:30",
		Status: po.AutomationStatusDisabled,
	}}
	tt := newUpdateTool(t, repo)

	invokeUpdate(t, tt, tools.UpdateAutomationTaskRequest{
		TaskId: 5, Title: "t2", Requirement: "r2",
		ExecType: po.AutomationExecTypeDaily, FixedTime: "10:00",
	})
	if repo.updated == nil || repo.updated.Title != "t2" {
		t.Fatalf("停用任务应允许修改，得到 %+v", repo.updated)
	}
	if !repo.gotRecompute {
		t.Fatal("停用任务的计划变更同样应重算 NextRunAt")
	}
}

func TestUpdateAutomationTaskValidationAndErrors(t *testing.T) {
	weekly := uint8(4)
	cases := []struct {
		name string
		req  tools.UpdateAutomationTaskRequest
		repo *updateFakeRepo
	}{
		{"标题为空", tools.UpdateAutomationTaskRequest{TaskId: 5, Requirement: "r", ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 10}, &updateFakeRepo{}},
		{"每周缺weekday", tools.UpdateAutomationTaskRequest{TaskId: 5, Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeWeekly, FixedTime: "09:00"}, &updateFakeRepo{}},
		{"任务不存在", tools.UpdateAutomationTaskRequest{TaskId: 99, Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeDaily, FixedTime: "09:00"}, &updateFakeRepo{getErr: &assertError{"not found"}}},
		{"仓储错误透传", tools.UpdateAutomationTaskRequest{TaskId: 5, Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeDaily, FixedTime: "09:00"}, &updateFakeRepo{stored: &po.AutomationTask{Id: 5, Status: po.AutomationStatusEnabled}, updateErr: &assertError{"db down"}}},
	}
	for _, c := range cases {
		tt := newUpdateTool(t, c.repo)
		data, err := json.Marshal(c.req)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tt.InvokableRun(context.Background(), string(data)); err == nil {
			t.Fatalf("%s 应返回错误", c.name)
		}
	}
	// 每周任务 weekday 显式传入（含周日 0）应正常通过
	repo := &updateFakeRepo{stored: &po.AutomationTask{
		Id: 5, Title: "t", Requirement: "r", UserId: "user-001",
		ExecType: po.AutomationExecTypeWeekly, FixedTime: "20:00", Weekday: 0,
		Status: po.AutomationStatusEnabled,
	}}
	tt := newUpdateTool(t, repo)
	invokeUpdate(t, tt, tools.UpdateAutomationTaskRequest{
		TaskId: 5, Title: "t", Requirement: "r",
		ExecType: po.AutomationExecTypeWeekly, FixedTime: "20:00", Weekday: &weekly,
	})
	if repo.updated == nil || repo.updated.Weekday != 4 {
		t.Fatalf("weekly 任务应全量回显成功，得到 %+v", repo.updated)
	}
}

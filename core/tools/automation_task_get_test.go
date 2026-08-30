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

// getFakeRepo 提供单任务查询行为的存根
type getFakeRepo struct {
	fakeAutomationRepo
	stored   *po.AutomationTask
	gotId    int
	gotUser  string
	getErr   error
}

func (f *getFakeRepo) GetTask(id int, userId string) (*po.AutomationTask, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	f.gotId = id
	f.gotUser = userId
	return f.stored, nil
}

func newGetTool(t *testing.T, repo iface.AutomationTaskRepo) tool.InvokableTool {
	t.Helper()
	invokable, err := tools.NewAutomationTaskGetter("user-001", repo)
	if err != nil {
		t.Fatal(err)
	}
	return invokable
}

func invokeGet(t *testing.T, tt tool.InvokableTool, taskId int) string {
	t.Helper()
	data, err := json.Marshal(tools.GetAutomationTaskRequest{TaskId: taskId})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tt.InvokableRun(context.Background(), string(data))
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	return raw
}

func TestGetAutomationTaskFullDetail(t *testing.T) {
	next := time.Date(2026, 9, 1, 8, 30, 0, 0, time.Local)
	created := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	repo := &getFakeRepo{stored: &po.AutomationTask{
		Id: 7, Title: "整理下载目录", Requirement: "每天整理 downloads 目录并汇总到周报",
		UserId: "user-001", ExecType: po.AutomationExecTypeDaily, FixedTime: "09:30",
		NextRunAt: next, Created: created, Status: po.AutomationStatusEnabled,
	}}
	tt := newGetTool(t, repo)

	raw := invokeGet(t, tt, 7)

	for _, key := range []string{
		`"task_id":7`, `"title":"整理下载目录"`,
		`"requirement":"每天整理 downloads 目录并汇总到周报"`,
		`"status":1`, `"exec_type":3`, `"fixed_time":"09:30"`,
		`"next_run_at":"2026-09-01 08:30"`, `"created":"2026-08-27 09:00"`,
	} {
		if !strings.Contains(raw, key) {
			t.Fatalf("响应缺少 %s：%s", key, raw)
		}
	}
	// 暂存字段不允许出现；每日任务不携带 weekday
	for _, key := range []string{`"mcp_ids"`, `"skill_ids"`, `"ai_model_id"`, `"persona_id"`, `"weekday":`} {
		if strings.Contains(raw, key) {
			t.Fatalf("响应不允许返回 %s：%s", key, raw)
		}
	}
	if repo.gotId != 7 || repo.gotUser != "user-001" {
		t.Fatalf("仓储应收到 task_id=7 与 user-001，得到 %d/%s", repo.gotId, repo.gotUser)
	}
}

func TestGetAutomationTaskWeeklySundayKeepsZeroWeekday(t *testing.T) {
	next := time.Date(2026, 9, 6, 20, 0, 0, 0, time.Local)
	repo := &getFakeRepo{stored: &po.AutomationTask{
		Id: 4, Title: "周日巡检", Requirement: "r", UserId: "user-001",
		ExecType: po.AutomationExecTypeWeekly, FixedTime: "20:00", Weekday: 0,
		NextRunAt: next, Status: po.AutomationStatusDisabled,
	}}
	tt := newGetTool(t, repo)

	raw := invokeGet(t, tt, 4)

	// 周日任务必须显式输出 "weekday":0（与"无 weekday"可区分），停用任务可查
	if !strings.Contains(raw, `"weekday":0`) || !strings.Contains(raw, `"status":0`) {
		t.Fatalf("周日停用任务应输出 weekday:0 与 status:0：%s", raw)
	}
}

func TestGetAutomationTaskNotFound(t *testing.T) {
	repo := &getFakeRepo{getErr: assertErr("record not found")}
	tt := newGetTool(t, repo)

	data, _ := json.Marshal(tools.GetAutomationTaskRequest{TaskId: 999})
	_, err := tt.InvokableRun(context.Background(), string(data))
	if err == nil || !strings.Contains(err.Error(), "999") {
		t.Fatalf("任务不存在应返回含 ID 的错误，得到 %v", err)
	}
}

func assertErr(msg string) error { return &assertError{msg} }

type assertError struct{ msg string }

func (e *assertError) Error() string { return e.msg }

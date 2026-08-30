package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/core/iface"
	"goraven/core/tools"

	"github.com/cloudwego/eino/components/tool"
)

// fakeAutomationRepo 捕获工具落库的任务，并模拟仓储行为（回填 Id 与 NextRunAt）
type fakeAutomationRepo struct {
	captured *po.AutomationTask
	err      error
}

func (f *fakeAutomationRepo) CreateTask(task *po.AutomationTask) error {
	if f.err != nil {
		return f.err
	}
	f.captured = task
	task.Id = 42
	task.NextRunAt = time.Date(2026, 9, 1, 8, 30, 0, 0, time.Local)
	return nil
}

// ListEnabledTasks 满足 iface.AutomationTaskRepo 的查询存根（创建用例不触发）
func (f *fakeAutomationRepo) ListEnabledTasks(userId, title string, page, pageSize int) ([]po.AutomationTask, int64, error) {
	return nil, 0, nil
}

// GetTask / UpdateTask 满足 iface.AutomationTaskRepo 的存根（创建用例不触发）
func (f *fakeAutomationRepo) GetTask(id int, userId string) (*po.AutomationTask, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeAutomationRepo) UpdateTask(task *po.AutomationTask, recomputeNext bool) error {
	return fmt.Errorf("not implemented")
}

// newTestTool 以默认暂存上下文创建工具，返回 Eino 可调用入口
func newTestTool(repo iface.AutomationTaskRepo, staging tools.AutomationStaging) tool.InvokableTool {
	t, err := tools.NewAutomationTaskCreator("user-001", staging, repo)
	if err != nil {
		panic(err)
	}
	return t
}

func mustJSON(t *testing.T, req tools.CreateAutomationTaskRequest) string {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCreateAutomationTaskHappyPath(t *testing.T) {
	fake := &fakeAutomationRepo{}
	staging := tools.AutomationStaging{
		McpIds:          []int{1, 2},
		SkillIds:        []int{3},
		Project:         "blog",
		SharedProjectId: 2,
		AIModelId:       5,
		PersonaId:       7,
	}
	tt := newTestTool(fake, staging)

	rspRaw, err := tt.InvokableRun(context.Background(), mustJSON(t, tools.CreateAutomationTaskRequest{
		Title:           "每日备份",
		Requirement:     "每天整理 downloads 目录",
		ExecType:        po.AutomationExecTypeInterval,
		IntervalMinutes: 30,
	}))
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	if !strings.Contains(rspRaw, "\"task_id\":42") || !strings.Contains(rspRaw, "2026-09-01 08:30") {
		t.Fatalf("响应应包含任务 ID 与下次执行时间，得到 %s", rspRaw)
	}

	if fake.captured == nil {
		t.Fatal("工具未调用仓储 CreateTask")
	}
	task := fake.captured
	if task.UserId != "user-001" {
		t.Fatalf("UserId 应为 user-001，得到 %s", task.UserId)
	}
	if task.McpIds != "[1,2]" || task.SkillIds != "[3]" {
		t.Fatalf("McpIds/SkillIds 应序列化为 JSON 数组，得到 %q / %q", task.McpIds, task.SkillIds)
	}
	if task.Project != "blog" || task.SharedProjectId != 2 || task.AIModelId != 5 || task.PersonaId != 7 {
		t.Fatalf("暂存字段透传不正确: %+v", task)
	}
	if task.IntervalMinutes != 30 || task.ExecType != po.AutomationExecTypeInterval {
		t.Fatalf("执行参数透传不正确: %+v", task)
	}
	if task.Status != po.AutomationStatusEnabled {
		t.Fatalf("新任务状态应为启用(%d)，得到 %d", po.AutomationStatusEnabled, task.Status)
	}
	if task.Title == "" || task.Requirement == "" {
		t.Fatal("标题与需求必须落库")
	}
}

func TestCreateAutomationTaskEmptyStagingSerializedAsEmptyString(t *testing.T) {
	fake := &fakeAutomationRepo{}
	tt := newTestTool(fake, tools.AutomationStaging{})

	runAt := time.Now().Add(time.Hour).Format("2006-01-02T15:04:05")
	if _, err := tt.InvokableRun(context.Background(), mustJSON(t, tools.CreateAutomationTaskRequest{
		Title:       "t",
		Requirement: "r",
		ExecType:    po.AutomationExecTypeOnce,
		RunAt:       runAt,
	})); err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	if fake.captured == nil || fake.captured.McpIds != "" || fake.captured.SkillIds != "" {
		t.Fatalf("空列表应序列化为空字符串以对齐 session 快照语义，得到 %+v", fake.captured)
	}
}

func TestCreateAutomationTaskValidation(t *testing.T) {
	staging := tools.AutomationStaging{}
	past := time.Now().Add(-time.Hour).Format("2006-01-02 15:04")

	cases := []struct {
		name string
		req  tools.CreateAutomationTaskRequest
	}{
		{"标题为空", tools.CreateAutomationTaskRequest{Requirement: "r", ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 10}},
		{"需求为空", tools.CreateAutomationTaskRequest{Title: "t", ExecType: po.AutomationExecTypeInterval, IntervalMinutes: 10}},
		{"未知执行类型", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: 9}},
		{"单次缺RunAt", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeOnce}},
		{"单次时间格式非法", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeOnce, RunAt: "明天上午十点"}},
		{"单次时间为过去", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeOnce, RunAt: past}},
		{"间隔小于1分钟", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeInterval}},
		{"每天缺FixedTime", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeDaily}},
		{"每天时间格式非法", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeDaily, FixedTime: "2130"}},
		{"每周缺weekday", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeWeekly, FixedTime: "09:00"}},
		{"每周星期越界", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeWeekly, FixedTime: "09:00", Weekday: ptrUint8(7)}},
		{"每周缺FixedTime", tools.CreateAutomationTaskRequest{Title: "t", Requirement: "r", ExecType: po.AutomationExecTypeWeekly, Weekday: ptrUint8(1)}},
	}
	for _, c := range cases {
		fake := &fakeAutomationRepo{}
		tt := newTestTool(fake, staging)
		_, err := tt.InvokableRun(context.Background(), mustJSON(t, c.req))
		if err == nil {
			t.Fatalf("%s 应返回校验错误", c.name)
		}
		if fake.captured != nil {
			t.Fatalf("%s 校验失败时不应写库", c.name)
		}
	}
}

// ptrUint8 构造 *uint8 字面量
func ptrUint8(v uint8) *uint8 { return &v }

func TestCreateAutomationTaskWeeklySundayWeekdayZero(t *testing.T) {
	fake := &fakeAutomationRepo{}
	tt := newTestTool(fake, tools.AutomationStaging{})

	// 每周日任务 weekday=0 必须显式传 0，且能正常创建
	if _, err := tt.InvokableRun(context.Background(), mustJSON(t, tools.CreateAutomationTaskRequest{
		Title:       "周日巡检",
		Requirement: "r",
		ExecType:    po.AutomationExecTypeWeekly,
		FixedTime:   "20:00",
		Weekday:     ptrUint8(0),
	})); err != nil {
		t.Fatalf("weekday=0（周日）应创建成功: %v", err)
	}
	if fake.captured == nil || fake.captured.Weekday != 0 {
		t.Fatalf("weekday=0 应正确落库，得到 %+v", fake.captured)
	}
}

func TestCreateAutomationTaskFutureRunAtFlexibleLayouts(t *testing.T) {
	fake := &fakeAutomationRepo{}
	tt := newTestTool(fake, tools.AutomationStaging{})

	layouts := []string{
		time.Now().Add(time.Hour).Format("2006-01-02T15:04"),
		time.Now().Add(time.Hour).Format("2006-01-02 15:04"),
		time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	for _, v := range layouts {
		if _, err := tt.InvokableRun(context.Background(), mustJSON(t, tools.CreateAutomationTaskRequest{
			Title:       "t",
			Requirement: "r",
			ExecType:    po.AutomationExecTypeOnce,
			RunAt:       v,
		})); err != nil {
			t.Fatalf("布局 %s 应解析成功: %v", v, err)
		}
	}
}

func TestCreateAutomationTaskRepoErrorPropagates(t *testing.T) {
	fake := &fakeAutomationRepo{err: fmt.Errorf("db down")}
	tt := newTestTool(fake, tools.AutomationStaging{})

	runAt := time.Now().Add(time.Hour).Format("2006-01-02T15:04")
	_, err := tt.InvokableRun(context.Background(), mustJSON(t, tools.CreateAutomationTaskRequest{
		Title:       "t",
		Requirement: "r",
		ExecType:    po.AutomationExecTypeOnce,
		RunAt:       runAt,
	}))
	if err == nil || !strings.Contains(err.Error(), "db down") {
		t.Fatalf("仓储错误应向上传播，得到 %v", err)
	}
}

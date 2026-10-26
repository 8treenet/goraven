package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/core/tools"

	"github.com/cloudwego/eino/components/tool"
)

// listFixture 分页查询测试数据（按 Id 排序即分页顺序）
func listFixture() []po.AutomationTask {
	next := time.Date(2026, 9, 1, 8, 30, 0, 0, time.Local)
	runAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.Local)
	created := time.Date(2026, 8, 27, 9, 0, 0, 0, time.Local)
	return []po.AutomationTask{
		{Id: 1, Title: "每日备份", UserId: "user-001", ExecType: po.AutomationExecTypeDaily, FixedTime: "09:30", NextRunAt: next, Created: created},
		{Id: 2, Title: "周报汇总", UserId: "user-001", ExecType: po.AutomationExecTypeWeekly, FixedTime: "18:00", Weekday: 1, NextRunAt: next, Created: created},
		{Id: 3, Title: "单次提醒", UserId: "user-001", ExecType: po.AutomationExecTypeOnce, RunAt: &runAt, NextRunAt: runAt, Created: created},
		{Id: 4, Title: "周日巡检", UserId: "user-001", ExecType: po.AutomationExecTypeWeekly, FixedTime: "20:00", Weekday: 0, NextRunAt: next, Created: created},
	}
}

// listFakeRepo 在创建用 fake 的基础上补充分页查询行为
type listFakeRepo struct {
	fakeAutomationRepo
	data        []po.AutomationTask
	gotUserId   string
	gotTitle    string
	gotPage     int
	gotPageSize int
	err         error
}

func (f *listFakeRepo) ListEnabledTasks(userId, title string, page, pageSize int) ([]po.AutomationTask, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.gotUserId = userId
	f.gotTitle = title
	f.gotPage = page
	f.gotPageSize = pageSize
	filtered := f.data
	if title != "" { // 模拟仓储 title LIKE %title% 行为
		var matched []po.AutomationTask
		for i := range filtered {
			if strings.Contains(filtered[i].Title, title) {
				matched = append(matched, filtered[i])
			}
		}
		filtered = matched
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], int64(len(filtered)), nil
}

func newListTool(t *testing.T, repo *listFakeRepo) tool.InvokableTool {
	t.Helper()
	invokable, err := tools.NewAutomationTaskLister("user-001", repo)
	if err != nil {
		t.Fatal(err)
	}
	return invokable
}

func invokeList(t *testing.T, tt tool.InvokableTool, req tools.ListAutomationTasksRequest) string {
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

func TestListAutomationTasksDefaults(t *testing.T) {
	repo := &listFakeRepo{data: listFixture()}
	tt := newListTool(t, repo)

	raw := invokeList(t, tt, tools.ListAutomationTasksRequest{})

	// 默认分页：page=1, size=10，全量返回
	if repo.gotPage != 1 || repo.gotPageSize != 10 {
		t.Fatalf("默认分页应为 1/10，仓储收到 %d/%d", repo.gotPage, repo.gotPageSize)
	}
	for _, key := range []string{
		`"page":1`, `"page_size":10`, `"total":4`,
		`"title":"每日备份"`, `"title":"周日巡检"`,
		`"task_id":1`, `"task_id":4`,
		`"next_run_at":"2026-09-01 08:30"`,
		`"run_at":"2026-08-30 10:00"`, `"fixed_time":"09:30"`,
		`"weekday":0`, `"weekday":1`,
	} {
		if !strings.Contains(raw, key) {
			t.Fatalf("响应缺少 %s：%s", key, raw)
		}
	}
	// weekday 仅周任务输出：4 条任务中仅 2 条周任务携带该字段
	if got := strings.Count(raw, `"weekday":`); got != 2 {
		t.Fatalf("weekday 应仅出现在 2 条周任务中，得到 %d 次：%s", got, raw)
	}
	// 暂存字段与 Requirement 不允许出现；任务 ID 统一为 task_id，不允许裸 id
	for _, key := range []string{
		`"requirement"`, `"mcp_ids"`, `"skill_ids"`, `"project"`,
		`"shared_project_id"`, `"ai_model_id"`, `"persona_id"`, `"id":`,
	} {
		if strings.Contains(raw, key) {
			t.Fatalf("响应不允许返回 %s：%s", key, raw)
		}
	}
}

func TestListAutomationTasksExplicitPagination(t *testing.T) {
	repo := &listFakeRepo{data: listFixture()}
	tt := newListTool(t, repo)

	raw := invokeList(t, tt, tools.ListAutomationTasksRequest{Page: 2, PageSize: 1})

	if repo.gotPage != 2 || repo.gotPageSize != 1 {
		t.Fatalf("分页参数应为 2/1，仓储收到 %d/%d", repo.gotPage, repo.gotPageSize)
	}
	if !strings.Contains(raw, `"page":2`) || !strings.Contains(raw, `"page_size":1`) {
		t.Fatalf("响应应回显分页信息：%s", raw)
	}
	if !strings.Contains(raw, `"title":"周报汇总"`) || strings.Contains(raw, `"title":"每日备份"`) {
		t.Fatalf("第 2 页应只包含第二条数据：%s", raw)
	}
}

func TestListAutomationTasksTitleFilter(t *testing.T) {
	repo := &listFakeRepo{data: listFixture()}
	tt := newListTool(t, repo)

	raw := invokeList(t, tt, tools.ListAutomationTasksRequest{Title: "周"})

	if repo.gotTitle != "周" {
		t.Fatalf("标题过滤参数应透传给仓储，收到 %q", repo.gotTitle)
	}
	if !strings.Contains(raw, `"title":"周报汇总"`) || !strings.Contains(raw, `"title":"周日巡检"`) {
		t.Fatalf("应包含标题匹配的任务：%s", raw)
	}
	if strings.Contains(raw, `"title":"每日备份"`) || strings.Contains(raw, `"title":"单次提醒"`) {
		t.Fatalf("不匹配的任务不应返回：%s", raw)
	}
	if !strings.Contains(raw, `"total":2`) {
		t.Fatalf("total 应为匹配数 2：%s", raw)
	}
}

func TestListAutomationTasksClampAndNormalize(t *testing.T) {
	repo := &listFakeRepo{data: listFixture()}
	tt := newListTool(t, repo)

	// page<=0 归一为 1；page_size 超过上限裁到 100
	invokeList(t, tt, tools.ListAutomationTasksRequest{Page: -3, PageSize: 250})
	if repo.gotPage != 1 || repo.gotPageSize != 100 {
		t.Fatalf("越界分页应归一为 1/100，仓储收到 %d/%d", repo.gotPage, repo.gotPageSize)
	}
}

func TestListAutomationTasksEmptyResult(t *testing.T) {
	repo := &listFakeRepo{data: []po.AutomationTask{}}
	tt := newListTool(t, repo)

	raw := invokeList(t, tt, tools.ListAutomationTasksRequest{})
	if !strings.Contains(raw, `"items":[]`) {
		t.Fatalf("空结果 items 应序列化为空数组而非 null：%s", raw)
	}
}

func TestListAutomationTasksRepoErrorPropagates(t *testing.T) {
	tt := newListTool(t, &listFakeRepo{err: fmt.Errorf("db down")})

	data, err := json.Marshal(tools.ListAutomationTasksRequest{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = tt.InvokableRun(context.Background(), string(data))
	if err == nil || !strings.Contains(err.Error(), "db down") {
		t.Fatalf("仓储错误应向上传播，得到 %v", err)
	}
}

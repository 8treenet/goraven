package mock

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/config"
)

// TestBuildAutomationTasks 校验任务列表：总数、排序（created DESC）与名称字段填充。
func TestBuildAutomationTasks(t *testing.T) {
	rsp := BuildAutomationTasks(1, 20, nil)
	if rsp == nil {
		t.Fatal("expected non-nil page response")
	}
	if rsp.TotalCount != 4 {
		t.Fatalf("totalCount = %d, want 4", rsp.TotalCount)
	}
	list, ok := rsp.List.([]vo.AutomationTaskItem)
	if !ok {
		t.Fatalf("expected []vo.AutomationTaskItem, got %T", rsp.List)
	}
	if len(list) != 4 {
		t.Fatalf("list len = %d, want 4", len(list))
	}
	// 按创建时间倒序：每日站会总结(18天前) > 季度经营报告(23天前) > 竞品动态监控(27天前) > 每周代码审查(39天前)
	wantOrder := []int{1, 4, 3, 2}
	for i, want := range wantOrder {
		if list[i].Id != want {
			t.Fatalf("list[%d].Id = %d, want %d", i, list[i].Id, want)
		}
	}

	first := list[0]
	// 模型名对齐种子库 displayName（02 SQL），语言无关
	if first.AIModelName != "deepseek-v4-pro" {
		t.Errorf("AIModelName = %q", first.AIModelName)
	}
	if len(first.McpNames) != 2 || first.McpNames[0] != "文件系统" || first.McpNames[1] != "邮件助手" {
		t.Errorf("McpNames = %v", first.McpNames)
	}
	if len(first.SkillNames) != 3 {
		t.Errorf("SkillNames = %v", first.SkillNames)
	}

	// 回归：名称列表必须序列化为 [] 而非 null（前端对 null 取 length 会崩溃）
	for _, item := range list {
		if item.McpNames == nil || item.SkillNames == nil {
			t.Errorf("task %d names must be empty slice, not nil", item.Id)
		}
	}
	data, err := json.Marshal(rsp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(data)
	if strings.Contains(body, `"mcpNames":null`) || strings.Contains(body, `"skillNames":null`) {
		t.Errorf("response must not contain null name lists: %s", body)
	}

	// 分页：每页 2 条第 2 页应为后两个任务
	rsp2 := BuildAutomationTasks(2, 2, nil)
	list2 := rsp2.List.([]vo.AutomationTaskItem)
	if len(list2) != 2 || list2[0].Id != 3 || list2[1].Id != 2 {
		t.Fatalf("page2 = [%d, %d], want [3, 2]", list2[0].Id, list2[1].Id)
	}

	// 状态过滤：启用中（status=1）只有任务 1 和 2
	enabled := uint8(po.AutomationStatusEnabled)
	rsp3 := BuildAutomationTasks(1, 20, &enabled)
	list3 := rsp3.List.([]vo.AutomationTaskItem)
	if rsp3.TotalCount != 2 || len(list3) != 2 {
		t.Fatalf("enabled filter: totalCount = %d, len = %d, want 2", rsp3.TotalCount, len(list3))
	}
	for _, item := range list3 {
		if item.Id != 1 && item.Id != 2 {
			t.Errorf("enabled filter should contain tasks 1/2 only, got task %d", item.Id)
		}
	}
}

// TestBuildAutomationTask 校验详情：requirement 与名称；不存在返回 nil。
func TestBuildAutomationTask(t *testing.T) {
	detail := BuildAutomationTask(2)
	if detail == nil {
		t.Fatal("task 2 should exist")
	}
	if detail.Title != "每周代码审查报告" {
		t.Errorf("Title = %q", detail.Title)
	}
	if detail.PersonaName != "代码审查员" {
		t.Errorf("PersonaName = %q", detail.PersonaName)
	}
	if detail.Requirement == "" {
		t.Error("Requirement should not be empty")
	}

	if BuildAutomationTask(99) != nil {
		t.Error("task 99 should be nil")
	}
}

// TestBuildAutomationExecutions 校验执行记录：ID 从新到旧、分页与时长为正。
func TestBuildAutomationExecutions(t *testing.T) {
	rsp := BuildAutomationExecutions(1, 1, 4)
	list, ok := rsp.List.([]vo.AutomationExecutionItem)
	if !ok {
		t.Fatalf("expected []vo.AutomationExecutionItem, got %T", rsp.List)
	}
	if rsp.TotalCount != 28 {
		t.Fatalf("totalCount = %d, want 28", rsp.TotalCount)
	}
	if rsp.TotalPage != 7 {
		t.Fatalf("totalPage = %d, want 7", rsp.TotalPage)
	}
	if len(list) != 4 {
		t.Fatalf("list len = %d, want 4", len(list))
	}
	// ID 复刻前端 genExecutions：128 起，从新到旧
	if list[0].Id != 128 || list[1].Id != 127 || list[3].Id != 125 {
		t.Fatalf("ids = [%d, %d, %d, %d], want [128, 127, 126, 125]", list[0].Id, list[1].Id, list[2].Id, list[3].Id)
	}
	for _, item := range list {
		if !item.FinishedAt.After(item.StartedAt) {
			t.Errorf("exec %d finishedAt should be after startedAt", item.Id)
		}
	}
	// 最新一条应为今天 09:00
	today9 := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 0, 0, 0, time.Now().Location())
	if !list[0].StartedAt.Equal(today9) {
		t.Errorf("newest startedAt = %v, want %v", list[0].StartedAt, today9)
	}

	// 越界页返回空列表
	empty := BuildAutomationExecutions(1, 99, 4)
	if n := len(empty.List.([]vo.AutomationExecutionItem)); n != 0 {
		t.Errorf("out-of-range page should be empty, got %d", n)
	}
}

// TestBuildAutomationAnswer 校验执行回复：精写表命中、兜底生成与未知任务。
func TestBuildAutomationAnswer(t *testing.T) {
	// 精写内容
	ans := BuildAutomationAnswer(1, 128)
	if !strings.HasPrefix(ans.Answer, "【站会纪要】") {
		t.Errorf("fixed answer = %q", ans.Answer)
	}

	// 兜底：回复模板含任务标题
	ans = BuildAutomationAnswer(1, 105)
	if !strings.Contains(ans.Answer, "每日站会总结") {
		t.Errorf("fallback answer = %q", ans.Answer)
	}
	if ans.Answer == "" {
		t.Error("fallback answer should not be empty")
	}

	// 未知任务返回空结构
	ans = BuildAutomationAnswer(99, 1)
	if ans.Answer != "" {
		t.Errorf("unknown task should return empty answer, got %+v", ans)
	}
}

// TestBuildAutomationEn 英文环境（system.language=en）应输出英文 mock 文案
func TestBuildAutomationEn(t *testing.T) {
	cfg := config.Get()
	orig := cfg.System.Language
	cfg.System.Language = "en"
	defer func() { cfg.System.Language = orig }()

	// 列表：标题 / MCP 名称 / 角色名
	list := BuildAutomationTasks(1, 20, nil).List.([]vo.AutomationTaskItem)
	if list[0].Title != "Daily Standup Summary" {
		t.Errorf("en Title = %q", list[0].Title)
	}
	if len(list[0].McpNames) != 2 || list[0].McpNames[0] != "Filesystem" || list[0].McpNames[1] != "Email Assistant" {
		t.Errorf("en McpNames = %v", list[0].McpNames)
	}
	detail := BuildAutomationTask(2)
	if detail == nil || detail.Title != "Weekly Code Review Report" || detail.PersonaName != "Code Reviewer" {
		t.Errorf("en detail = %+v", detail)
	}
	if detail != nil && !strings.HasPrefix(detail.Requirement, "Every Monday at 10 AM") {
		t.Errorf("en Requirement = %q", detail.Requirement)
	}

	// 精写回复
	ans := BuildAutomationAnswer(1, 128)
	if !strings.HasPrefix(ans.Answer, "[Standup Minutes]") {
		t.Errorf("en fixed answer = %q", ans.Answer)
	}

	// 兜底：回复模板为英文
	ans = BuildAutomationAnswer(1, 105)
	wantAnswer := `Task "Daily Standup Summary" has been executed`
	if !strings.HasPrefix(ans.Answer, wantAnswer) {
		t.Errorf("en fallback answer = %q", ans.Answer)
	}
}

// TestAutomationTaskSchedules 校验调度字段与状态的一致性。
func TestAutomationTaskSchedules(t *testing.T) {
	specs := automationTaskSpecs()
	for _, s := range specs {
		switch s.execType {
		case po.AutomationExecTypeOnce:
			if s.runAt == nil {
				t.Errorf("task %d once-type requires runAt", s.id)
			}
		case po.AutomationExecTypeDaily, po.AutomationExecTypeWeekly:
			if s.fixedTime == "" {
				t.Errorf("task %d requires fixedTime", s.id)
			}
		case po.AutomationExecTypeInterval:
			if s.intervalMinutes < 5 {
				t.Errorf("task %d intervalMinutes must be >= 5", s.id)
			}
		}
		if s.status == po.AutomationStatusEnabled && s.nextRunAt.IsZero() {
			t.Errorf("enabled task %d must have nextRunAt", s.id)
		}
		if s.status == po.AutomationStatusDisabled && !s.nextRunAt.IsZero() {
			t.Errorf("disabled task %d nextRunAt should be zero", s.id)
		}
	}
}

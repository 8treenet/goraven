package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"goraven/backend/po"
	"goraven/config"
	"goraven/core/iface"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UpdateAutomationTaskRequest 更新自动化任务工具的请求参数（全量替换业务字段）
type UpdateAutomationTaskRequest struct {
	TaskId          int    `json:"task_id" jsonschema:"description=The task ID to update"`
	Title           string `json:"title" jsonschema:"description=New brief title shown to the user"`
	Requirement     string `json:"requirement" jsonschema:"description=New self-contained instruction executed at trigger time; include all context, paths and fallback behavior"`
	ExecType        uint8  `json:"exec_type" jsonschema:"description=1=once at run_at, 2=every interval_minutes, 3=daily at fixed_time, 4=weekly on weekday at fixed_time"`
	RunAt           string `json:"run_at,omitempty" jsonschema:"description=exec_type=1 only. Future time; RFC3339 or YYYY-MM-DD HH:MM[:SS]"`
	IntervalMinutes int    `json:"interval_minutes,omitempty" jsonschema:"description=exec_type=2 only. Minutes, min 5"`
	FixedTime       string `json:"fixed_time,omitempty" jsonschema:"description=exec_type=3/4 only. HH:MM, e.g. 09:30"`
	Weekday         *uint8 `json:"weekday,omitempty" jsonschema:"description=exec_type=4 only, required. 0=Sunday..6=Saturday"`
}

// UpdateAutomationTaskResponse 更新自动化任务工具的响应
type UpdateAutomationTaskResponse struct {
	TaskId    int    `json:"task_id" jsonschema:"description=The updated automation task ID"`
	Title     string `json:"title" jsonschema:"description=The task title as stored"`
	NextRunAt string `json:"next_run_at" jsonschema:"description=Formatted next execution time (YYYY-MM-DD HH:MM), relay this to the user"`
}

// AutomationTaskUpdater 更新自动化任务工具
// 全量替换业务字段（标题/需求/执行计划），校验与 create 共用；
// 执行计划变更时重算 NextRunAt，仅改标题或需求时保持原值；
// 暂存上下文（MCP/技能/项目/模型/角色）与状态不由本工具修改；已完成任务拒绝修改。
type AutomationTaskUpdater struct {
	Name   string
	Desc   string
	userID string
	repo   iface.AutomationTaskRepo
}

const (
	UpdateAutomationTaskToolDesc = `Updates an automation task by full replacement. First fetch current fields via goraven_get_automation_task, then send task_id plus ALL of title, requirement, exec_type and its schedule params with your changes applied. Omit fields irrelevant to exec_type. Any schedule change resets next_run_at; changing only title/requirement keeps it. Disabled tasks can be updated, done ones cannot. Confirm with the user first.`

	UpdateAutomationTaskToolDescChinese = `以全量替换更新自动化任务。先用 goraven_get_automation_task 获取当前字段，再提交 task_id 及全部业务字段（标题、需求、exec_type 及对应计划参数），仅应用要修改的部分。与 exec_type 无关的字段不要传。修改任一计划字段会重置下次执行时间；仅改标题/需求则保持。停用任务可改，已完成任务不可改。先与用户确认。`
)

// NewAutomationTaskUpdater 创建更新自动化任务工具
func NewAutomationTaskUpdater(userID string, repo iface.AutomationTaskRepo) (tool.InvokableTool, error) {
	if repo == nil {
		return nil, fmt.Errorf("automation task repo is required")
	}

	desc := UpdateAutomationTaskToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = UpdateAutomationTaskToolDescChinese
	}

	t := &AutomationTaskUpdater{
		Name:   "goraven_update_automation_task",
		Desc:   desc,
		userID: userID,
		repo:   repo,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

// Invoke 全量替换任务业务字段：校验与 create 复用，计划变更时重算下次执行时间
func (u *AutomationTaskUpdater) Invoke(_ context.Context, req *UpdateAutomationTaskRequest) (*UpdateAutomationTaskResponse, error) {
	title := strings.TrimSpace(req.Title)
	requirement := strings.TrimSpace(req.Requirement)
	if title == "" {
		return nil, automationToolErr("title is required", "标题不能为空")
	}
	if requirement == "" {
		return nil, automationToolErr("requirement is required", "需求描述不能为空")
	}

	schedule, err := validateSchedule(&CreateAutomationTaskRequest{
		ExecType:        req.ExecType,
		RunAt:           req.RunAt,
		IntervalMinutes: req.IntervalMinutes,
		FixedTime:       req.FixedTime,
		Weekday:         req.Weekday,
	}, time.Now())
	if err != nil {
		return nil, err
	}

	task, err := u.repo.GetTask(req.TaskId, u.userID)
	if err != nil {
		return nil, automationToolErr(
			"automation task %d not found",
			"自动化任务 %d 不存在或不可访问", req.TaskId)
	}
	if task.Status == po.AutomationStatusDone {
		return nil, automationToolErr(
			"automation task %d is done (completed) and cannot be modified",
			"自动化任务 %d 已完成，不能修改", req.TaskId)
	}

	// 计划字段与库中不一致时重算 NextRunAt；仅改标题/需求时保持原值
	recompute := task.ExecType != req.ExecType ||
		!sameAutomationMinute(task.RunAt, schedule.RunAt) ||
		task.IntervalMinutes != schedule.IntervalMinutes ||
		task.FixedTime != schedule.FixedTime ||
		task.Weekday != schedule.Weekday

	task.Title = title
	task.Requirement = requirement
	task.ExecType = req.ExecType
	task.RunAt = schedule.RunAt
	task.IntervalMinutes = schedule.IntervalMinutes
	task.FixedTime = schedule.FixedTime
	task.Weekday = schedule.Weekday

	if err := u.repo.UpdateTask(task, recompute); err != nil {
		return nil, err
	}
	next := ""
	if !task.NextRunAt.IsZero() {
		next = task.NextRunAt.Format("2006-01-02 15:04")
	}
	return &UpdateAutomationTaskResponse{
		TaskId:    task.Id,
		Title:     task.Title,
		NextRunAt: next,
	}, nil
}

// sameAutomationMinute 以分钟粒度比较执行时间，秒级差异视为未变化
func sameAutomationMinute(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Format("2006-01-02 15:04") == b.Format("2006-01-02 15:04")
}

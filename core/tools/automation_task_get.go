package tools

import (
	"context"
	"fmt"

	"goraven/backend/po"
	"goraven/config"
	"goraven/core/iface"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetAutomationTaskRequest 查询自动化任务详情工具的请求参数
type GetAutomationTaskRequest struct {
	TaskId int `json:"task_id" jsonschema:"description=The task ID to fetch"`
}

// GetAutomationTaskResponse 自动化任务完整详情，含需求描述与状态；
// 字段命名与 create/list/update 工具保持一致，不含会话暂存字段
type GetAutomationTaskResponse struct {
	TaskId          int    `json:"task_id" jsonschema:"description=The task ID"`
	Title           string `json:"title" jsonschema:"description=The task title"`
	Requirement     string `json:"requirement" jsonschema:"description=The full requirement executed at trigger time"`
	Status          uint8  `json:"status" jsonschema:"description=0=disabled, 1=enabled, 2=done"`
	ExecType        uint8  `json:"exec_type" jsonschema:"description=1=once, 2=interval, 3=daily, 4=weekly"`
	RunAt           string `json:"run_at,omitempty" jsonschema:"description=Run time, YYYY-MM-DD HH:MM (once)"`
	IntervalMinutes int    `json:"interval_minutes,omitempty" jsonschema:"description=Interval minutes (interval)"`
	FixedTime       string `json:"fixed_time,omitempty" jsonschema:"description=HH:MM (daily/weekly)"`
	Weekday         *uint8 `json:"weekday,omitempty" jsonschema:"description=0=Sunday..6=Saturday (weekly)"`
	NextRunAt       string `json:"next_run_at" jsonschema:"description=Next run time, YYYY-MM-DD HH:MM"`
	Created         string `json:"created" jsonschema:"description=Creation time, YYYY-MM-DD HH:MM"`
}

// AutomationTaskGetter 查询自动化任务详情工具
// 返回单个任务的完整信息（含需求描述），供修改任务前获取当前全部字段；
// 启用/停用/已完成任务均可查询，暂存字段（MCP/技能/项目/模型/角色）不对外返回。
type AutomationTaskGetter struct {
	Name   string
	Desc   string
	userID string
	repo   iface.AutomationTaskRepo
}

const (
	GetAutomationTaskToolDesc = `Returns one automation task's full detail by task_id, including the complete requirement and status (0=disabled, 1=enabled, 2=done). Any task of the user is queryable. Call it before goraven_update_automation_task to obtain the current field values.`

	GetAutomationTaskToolDescChinese = `按 task_id 返回自动化任务完整详情，含完整需求描述与状态（0=停用，1=启用，2=已完成）。修改任务前先用本工具获取当前字段值。`
)

// NewAutomationTaskGetter 创建查询自动化任务详情工具
func NewAutomationTaskGetter(userID string, repo iface.AutomationTaskRepo) (tool.InvokableTool, error) {
	if repo == nil {
		return nil, fmt.Errorf("automation task repo is required")
	}

	desc := GetAutomationTaskToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = GetAutomationTaskToolDescChinese
	}

	t := &AutomationTaskGetter{
		Name:   "goraven_get_automation_task",
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

// Invoke 按 ID 查询当前用户的自动化任务并映射为完整详情
func (g *AutomationTaskGetter) Invoke(_ context.Context, req *GetAutomationTaskRequest) (*GetAutomationTaskResponse, error) {
	task, err := g.repo.GetTask(req.TaskId, g.userID)
	if err != nil {
		return nil, automationToolErr(
			"automation task %d not found",
			"自动化任务 %d 不存在或不可访问", req.TaskId)
	}

	resp := &GetAutomationTaskResponse{
		TaskId:          task.Id,
		Title:           task.Title,
		Requirement:     task.Requirement,
		Status:          task.Status,
		ExecType:        task.ExecType,
		RunAt:           formatAutomationTime(task.RunAt),
		IntervalMinutes: task.IntervalMinutes,
		FixedTime:       task.FixedTime,
		NextRunAt:       formatAutomationTime(&task.NextRunAt),
		Created:         formatAutomationTime(&task.Created),
	}
	if task.ExecType == po.AutomationExecTypeWeekly {
		weekday := task.Weekday
		resp.Weekday = &weekday
	}
	return resp, nil
}

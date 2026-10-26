package tools

import (
	"context"
	"fmt"
	"time"

	"goraven/backend/po"
	"goraven/config"
	"goraven/core/iface"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ListAutomationTasksRequest 查询自动化任务工具的请求参数
type ListAutomationTasksRequest struct {
	Page     int    `json:"page,omitempty" jsonschema:"description=Page number, default 1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"description=Per page, default 10, max 100"`
	Title    string `json:"title,omitempty" jsonschema:"description=Fuzzy match on task title (LIKE %title%), optional"`
}

// AutomationTaskSummary 自动化任务摘要，不包含需求描述与会话暂存字段；
// 字段命名与 create/get/update 工具保持一致（任务 ID 统一为 task_id）
type AutomationTaskSummary struct {
	TaskId          int    `json:"task_id" jsonschema:"description=The task ID"`
	Title           string `json:"title" jsonschema:"description=The task title"`
	ExecType        uint8  `json:"exec_type" jsonschema:"description=1=once, 2=interval, 3=daily, 4=weekly"`
	RunAt           string `json:"run_at,omitempty" jsonschema:"description=Run time, YYYY-MM-DD HH:MM (once)"`
	IntervalMinutes int    `json:"interval_minutes,omitempty" jsonschema:"description=Interval minutes (interval)"`
	FixedTime       string `json:"fixed_time,omitempty" jsonschema:"description=HH:MM (daily/weekly)"`
	Weekday         *uint8 `json:"weekday,omitempty" jsonschema:"description=0=Sunday..6=Saturday (weekly)"`
	NextRunAt       string `json:"next_run_at" jsonschema:"description=Next run time, YYYY-MM-DD HH:MM"`
	Created         string `json:"created" jsonschema:"description=Creation time, YYYY-MM-DD HH:MM"`
}

// ListAutomationTasksResponse 查询自动化任务工具的响应
type ListAutomationTasksResponse struct {
	Page     int                     `json:"page" jsonschema:"description=Current page number"`
	PageSize int                     `json:"page_size" jsonschema:"description=Items per page"`
	Total    int64                   `json:"total" jsonschema:"description=Total number of enabled tasks for the user"`
	Items    []AutomationTaskSummary `json:"items" jsonschema:"description=Task summaries of the current page"`
}

// AutomationTaskList 查询自动化任务工具
// 仅返回当前用户启用中且未删除的任务，字段为安全白名单：
// 需求描述与 MCP/技能/项目/模型/角色等暂存字段不对外返回。
type AutomationTaskList struct {
	Name   string
	Desc   string
	userID string
	repo   iface.AutomationTaskRepo
}

const (
	ListAutomationTasksToolDesc = `Lists the user's enabled automation tasks ordered by next run time, paginated (default 10, max 100). Optionally filter by fuzzy title match. Returns schedule summaries without the requirement text. Use it to find task IDs.`

	ListAutomationTasksToolDescChinese = `分页列出用户启用中的自动化任务（默认每页 10 条，按下次执行时间排序），可按标题模糊过滤，仅摘要不含需求描述。用于查找 task_id。`
)

// NewAutomationTaskLister 创建查询自动化任务工具
func NewAutomationTaskLister(userID string, repo iface.AutomationTaskRepo) (tool.InvokableTool, error) {
	if repo == nil {
		return nil, fmt.Errorf("automation task repo is required")
	}

	desc := ListAutomationTasksToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = ListAutomationTasksToolDescChinese
	}

	t := &AutomationTaskList{
		Name:   "goraven_list_automation_tasks",
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

// Invoke 分页查询当前用户启用中的自动化任务并映射为安全摘要
func (l *AutomationTaskList) Invoke(ctx context.Context, req *ListAutomationTasksRequest) (*ListAutomationTasksResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tasks, total, err := l.repo.ListEnabledTasks(l.userID, req.Title, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]AutomationTaskSummary, 0, len(tasks))
	for i := range tasks {
		task := &tasks[i]
		summary := AutomationTaskSummary{
			TaskId:          task.Id,
			Title:           task.Title,
			ExecType:        task.ExecType,
			RunAt:           formatAutomationTime(task.RunAt),
			IntervalMinutes: task.IntervalMinutes,
			FixedTime:       task.FixedTime,
			NextRunAt:       formatAutomationTime(&task.NextRunAt),
			Created:         formatAutomationTime(&task.Created),
		}
		if task.ExecType == po.AutomationExecTypeWeekly {
			weekday := task.Weekday
			summary.Weekday = &weekday
		}
		items = append(items, summary)
	}

	return &ListAutomationTasksResponse{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// formatAutomationTime 统一的时间展示格式，零值返回空串
func formatAutomationTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

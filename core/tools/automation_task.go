package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"goraven/backend/po"
	"goraven/config"
	"goraven/core/iface"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// AutomationStaging 自动化任务的暂存上下文，复用创建自动化任务时会话的字段，
// 触发执行时据此创建会话；由 Agent 构建时注入，不暴露给 LLM。
type AutomationStaging struct {
	McpIds          []int // 会话的 MCP 配置 ID 列表
	SkillIds        []int // 会话的技能 ID 列表
	Project         string
	SharedProjectId int
	AIModelId       int
	PersonaId       int
}

// CreateAutomationTaskRequest 创建自动化任务工具的请求参数
type CreateAutomationTaskRequest struct {
	Title           string `json:"title" jsonschema:"description=Brief title shown to the user"`
	Requirement     string `json:"requirement" jsonschema:"description=Self-contained instruction executed at trigger time; include all context, paths and fallback behavior"`
	ExecType        uint8  `json:"exec_type" jsonschema:"description=1=once at run_at, 2=every interval_minutes, 3=daily at fixed_time, 4=weekly on weekday at fixed_time"`
	RunAt           string `json:"run_at,omitempty" jsonschema:"description=exec_type=1 only. Future time; RFC3339 or YYYY-MM-DD HH:MM[:SS]"`
	IntervalMinutes int    `json:"interval_minutes,omitempty" jsonschema:"description=exec_type=2 only. Minutes, min 5"`
	FixedTime       string `json:"fixed_time,omitempty" jsonschema:"description=exec_type=3/4 only. HH:MM, e.g. 09:30"`
	Weekday         *uint8 `json:"weekday,omitempty" jsonschema:"description=exec_type=4 only, required. 0=Sunday..6=Saturday"`
}

// CreateAutomationTaskResponse 创建自动化任务工具的响应
type CreateAutomationTaskResponse struct {
	TaskId    int    `json:"task_id" jsonschema:"description=The created automation task ID"`
	Title     string `json:"title" jsonschema:"description=The task title as stored"`
	NextRunAt string `json:"next_run_at" jsonschema:"description=Formatted next execution time (YYYY-MM-DD HH:MM), relay this to the user"`
}

// CreateAutomationTask 创建自动化任务工具
// 通过对话创建定时任务：执行参数由 LLM 提供并校验，会话上下文（MCP/技能/项目/
// 模型/角色）在构建时注入暂存，持久化通过注入的 AutomationTaskRepo 完成。
type CreateAutomationTask struct {
	Name    string
	Desc    string
	userID  string
	staging AutomationStaging
	repo    iface.AutomationTaskRepo
}

const (
	CreateAutomationTaskToolDesc = `Creates a scheduled automation task. It inherits the current conversation's model, persona, project, MCP servers and skills automatically — never ask the user about them. Provide title, a self-contained requirement, and the schedule. Omit fields irrelevant to exec_type. Confirm the exact schedule with the user first if ambiguous.`

	CreateAutomationTaskToolDescChinese = `创建定时自动化任务。自动继承当前会话的模型、角色、项目、MCP 与技能，无需询问用户。提供标题、自包含的需求与执行计划；与 exec_type 无关的字段不要传。计划有歧义时先与用户确认。`
)

// NewAutomationTaskCreator 创建自动化任务工具
func NewAutomationTaskCreator(userID string, staging AutomationStaging, repo iface.AutomationTaskRepo) (tool.InvokableTool, error) {
	if repo == nil {
		return nil, fmt.Errorf("automation task repo is required")
	}

	desc := CreateAutomationTaskToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = CreateAutomationTaskToolDescChinese
	}

	t := &CreateAutomationTask{
		Name:    "goraven_create_automation_task",
		Desc:    desc,
		userID:  userID,
		staging: staging,
		repo:    repo,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

// automationSchedule 校验后的执行计划字段，create 与 update 共用；
// 仅保留 exec_type 对应的字段，其余归零，避免残留值落库
type automationSchedule struct {
	RunAt           *time.Time
	IntervalMinutes int
	FixedTime       string
	Weekday         uint8
}

// validateSchedule 校验执行计划参数并归一化，create 与 update 共用
func validateSchedule(req *CreateAutomationTaskRequest, now time.Time) (*automationSchedule, error) {
	switch req.ExecType {
	case po.AutomationExecTypeOnce:
		runAt, err := parseRunAt(req.RunAt, now)
		if err != nil {
			return nil, err
		}
		return &automationSchedule{RunAt: &runAt}, nil
	case po.AutomationExecTypeInterval:
		if req.IntervalMinutes < 5 {
			return nil, automationToolErr(
				"interval_minutes must be >= 5",
				"间隔分钟数不能小于 5")
		}
		return &automationSchedule{IntervalMinutes: req.IntervalMinutes}, nil
	case po.AutomationExecTypeDaily:
		if _, err := validateFixedTime(req.FixedTime); err != nil {
			return nil, err
		}
		return &automationSchedule{FixedTime: req.FixedTime}, nil
	case po.AutomationExecTypeWeekly:
		if req.Weekday == nil {
			return nil, automationToolErr(
				"weekday is required for exec_type=4 (0=Sunday, 1=Monday ... 6=Saturday)",
				"每周执行必须提供 weekday（0=周日，1=周一……6=周六）")
		}
		if *req.Weekday > 6 {
			return nil, automationToolErr(
				"weekday must be in [0, 6] (0=Sunday .. 6=Saturday)",
				"星期取值必须在 0-6 之间（0=周日，1=周一……6=周六）")
		}
		if _, err := validateFixedTime(req.FixedTime); err != nil {
			return nil, err
		}
		return &automationSchedule{FixedTime: req.FixedTime, Weekday: *req.Weekday}, nil
	default:
		return nil, automationToolErr(
			"unsupported exec_type: %d (valid: 1=once, 2=interval, 3=daily, 4=weekly)",
			"不支持的执行类型: %d（有效值：1=单次，2=按间隔，3=每天，4=每周）", req.ExecType)
	}
}

// Invoke 校验参数后构造任务并通过注入的仓储持久化
func (c *CreateAutomationTask) Invoke(ctx context.Context, req *CreateAutomationTaskRequest) (*CreateAutomationTaskResponse, error) {
	title := strings.TrimSpace(req.Title)
	requirement := strings.TrimSpace(req.Requirement)
	if title == "" {
		return nil, automationToolErr("title is required", "标题不能为空")
	}
	if requirement == "" {
		return nil, automationToolErr("requirement is required", "需求描述不能为空")
	}

	schedule, err := validateSchedule(req, time.Now())
	if err != nil {
		return nil, err
	}
	task := c.buildTask(title, requirement, req.ExecType, schedule)
	return c.create(ctx, task)
}

// buildTask 组装持久化对象：执行计划来自归一化结果，暂存上下文序列化语义与 session 快照一致（空列表存空串）
func (c *CreateAutomationTask) buildTask(title, requirement string, execType uint8, schedule *automationSchedule) *po.AutomationTask {
	mcpIds := ""
	skillIds := ""
	if len(c.staging.McpIds) > 0 {
		if data, err := json.Marshal(c.staging.McpIds); err == nil {
			mcpIds = string(data)
		}
	}
	if len(c.staging.SkillIds) > 0 {
		if data, err := json.Marshal(c.staging.SkillIds); err == nil {
			skillIds = string(data)
		}
	}

	return &po.AutomationTask{
		Title:           title,
		Requirement:     requirement,
		UserId:          c.userID,
		ExecType:        execType,
		RunAt:           schedule.RunAt,
		IntervalMinutes: schedule.IntervalMinutes,
		FixedTime:       schedule.FixedTime,
		Weekday:         schedule.Weekday,
		McpIds:          mcpIds,
		SkillIds:        skillIds,
		Project:         c.staging.Project,
		SharedProjectId: c.staging.SharedProjectId,
		AIModelId:       c.staging.AIModelId,
		PersonaId:       c.staging.PersonaId,
		Status:          po.AutomationStatusEnabled,
	}
}

// create 持久化并组装响应
func (c *CreateAutomationTask) create(_ context.Context, task *po.AutomationTask) (*CreateAutomationTaskResponse, error) {
	if err := c.repo.CreateTask(task); err != nil {
		return nil, err
	}
	next := ""
	if !task.NextRunAt.IsZero() {
		next = task.NextRunAt.Format("2006-01-02 15:04")
	}
	return &CreateAutomationTaskResponse{
		TaskId:    task.Id,
		Title:     task.Title,
		NextRunAt: next,
	}, nil
}

// parseRunAt 解析单次任务的执行时间，支持 RFC3339 与常用书写格式，且必须晚于当前时刻
func parseRunAt(value string, now time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, automationToolErr(
			"run_at is required for exec_type=1",
			"单次执行必须提供 run_at")
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, now.Location()); err == nil {
			if !t.After(now) {
				return time.Time{}, automationToolErr(
					"run_at must be in the future, got %s", "run_at 必须晚于当前时间，收到 %s", value)
			}
			return t, nil
		}
	}
	return time.Time{}, automationToolErr(
		"run_at is not a valid time, expected RFC3339 or 'YYYY-MM-DD HH:MM[:SS]', got %s",
		"run_at 时间格式无效，应为 RFC3339 或 'YYYY-MM-DD HH:MM[:SS]'，收到 %s", value)
}

// validateFixedTime 校验每天/每周类型的固定时间格式
func validateFixedTime(value string) (time.Time, error) {
	t, err := time.ParseInLocation("15:04", value, time.Local)
	if err != nil {
		return time.Time{}, automationToolErr(
			"fixed_time must be HH:MM format (e.g. 09:30), got %q", "fixed_time 必须为 HH:MM 格式（如 09:30），收到 %s", value)
	}
	return t, nil
}

// automationToolErr 按系统语言返回工具错误信息
func automationToolErr(formatEn, formatZh string, args ...any) error {
	if config.Get().GetLanguage() == "zh" {
		return fmt.Errorf(formatZh, args...)
	}
	return fmt.Errorf(formatEn, args...)
}

package vo

import "time"

// AutomationTaskListReq 自动化任务分页列表请求
type AutomationTaskListReq struct {
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
	Status   *uint8 `url:"status"`   // 状态筛选：0停用 1启用 2已完成，nil不筛选
}

// AutomationTaskItem 自动化任务条目（不含需求描述 requirement）
type AutomationTaskItem struct {
	Id                int        `json:"id"`                // 自动化任务ID
	Title             string     `json:"title"`             // 任务标题
	UserId            string     `json:"userId"`            // 所属用户ID
	ExecType          uint8      `json:"execType"`          // 执行类型：1单次固定时间 2按间隔 3每天固定时间 4每周固定时间
	RunAt             *time.Time `json:"runAt"`             // 单次执行时间（ExecType=1）
	IntervalMinutes   int        `json:"intervalMinutes"`   // 执行间隔分钟数（ExecType=2）
	FixedTime         string     `json:"fixedTime"`         // 固定时间 HH:MM（ExecType=3/4）
	Weekday           uint8      `json:"weekday"`           // 每周执行日 0=周日 1-6=周一至周六（ExecType=4）
	McpIds            string     `json:"mcpIds"`            // MCP配置ID列表（JSON数组）
	SkillIds          string     `json:"skillIds"`          // 技能ID列表（JSON数组）
	Project           string     `json:"project"`           // 项目目录名称
	SharedProjectId   int        `json:"sharedProjectId"`   // 团队项目ID，0表示个人项目
	AIModelId         int        `json:"aiModelId"`         // 使用的模型ID
	PersonaId         int        `json:"personaId"`         // 用户角色ID，0表示未选择
	AIModelName       string     `json:"aiModelName"`       // 模型展示名称（如 "DeepSeek - DeepSeek V3"），0或已删除为空
	PersonaName       string     `json:"personaName"`       // 角色名称，0或已删除为空
	SharedProjectName string     `json:"sharedProjectName"` // 团队项目名称，0或不存在为空
	McpNames          []string   `json:"mcpNames"`          // MCP展示名称列表（按 mcpIds 顺序，缺失跳过）
	SkillNames        []string   `json:"skillNames"`        // 技能名称列表（按 skillIds 顺序，缺失跳过）
	NextRunAt         time.Time  `json:"nextRunAt"`         // 下次执行时间
	Status            uint8      `json:"status"`            // 任务状态：0停用 1启用 2已完成
	Deleted           uint8      `json:"deleted"`           // 软删除：0正常 1删除
	Created           time.Time  `json:"created"`           // 创建时间
	Updated           time.Time  `json:"updated"`           // 更新时间
}

// AutomationTaskDetail 自动化任务详情（含需求描述）
type AutomationTaskDetail struct {
	AutomationTaskItem
	Requirement string `json:"requirement"` // 需求描述
}

// AutomationExecutionItem 自动化任务执行记录条目
type AutomationExecutionItem struct {
	Id         int       `json:"id"`         // 执行记录ID
	StartedAt  time.Time `json:"startedAt"`  // 实际开始执行时间
	FinishedAt time.Time `json:"finishedAt"` // 实际完成时间
}

// AutomationExecutionListReq 执行记录分页请求
type AutomationExecutionListReq struct {
	Page     int `url:"page"`     // 页码
	PageSize int `url:"pageSize"` // 每页条数
}

// AutomationTaskStatusReq 更新自动化任务状态请求
type AutomationTaskStatusReq struct {
	Status uint8 `json:"status"` // 目标状态：0停用 1启用
}

// AutomationTaskRequirementReq 修改自动化任务需求描述请求
type AutomationTaskRequirementReq struct {
	Requirement string `json:"requirement"` // 需求描述，不可为空
}

// AutomationQARsp 执行问答对响应（首条用户问题 + 助手最终回复）
type AutomationQARsp struct {
	Question string `json:"question"` // 用户问题内容
	Answer   string `json:"answer"`   // 助手回复内容
}

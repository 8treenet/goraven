package po

import (
	"time"

	"gorm.io/gorm"
)

// 自动化任务执行类型常量
const (
	AutomationExecTypeOnce     = 1 // 单次固定时间
	AutomationExecTypeInterval = 2 // 按间隔（每 N 分钟）
	AutomationExecTypeDaily    = 3 // 每天固定时间
	AutomationExecTypeWeekly   = 4 // 每周固定时间
)

// 自动化任务状态常量
const (
	AutomationStatusDisabled = 0 // 停用
	AutomationStatusEnabled  = 1 // 启用
	AutomationStatusDone     = 2 // 已完成（单次任务执行完成后）
)

// AutomationTask 自动化任务表（定时任务）
// 任务触发时依据暂存的 McpIds/SkillIds/Project/SharedProjectId 创建会话。
type AutomationTask struct {
	Id              int        `gorm:"primaryKey;column:id;type:int;autoIncrement"`    // 自动化任务ID
	Title           string     `gorm:"column:title;type:varchar(255);not null"`        // 任务标题
	Requirement     string     `gorm:"column:requirement;type:text"`                   // 需求描述
	UserId          string     `gorm:"column:user_id;type:varchar(64);index;not null"` // 所属用户ID
	ExecType        uint8      `gorm:"column:exec_type;not null"`                      // 执行类型：1单次固定时间 2按间隔 3每天固定时间 4每周固定时间
	RunAt           *time.Time `gorm:"column:run_at"`                                  // 单次执行时间（ExecType=1）
	IntervalMinutes int        `gorm:"column:interval_minutes;default:0"`              // 执行间隔分钟数（ExecType=2，最小5分钟）
	FixedTime       string     `gorm:"column:fixed_time;type:varchar(5);default:''"`   // 固定时间 HH:MM（ExecType=3/4）
	Weekday         uint8      `gorm:"column:weekday;default:0"`                       // 每周执行日 0=周日 1-6=周一至周六（ExecType=4）
	McpIds          string     `gorm:"column:mcp_ids;type:text"`                       // 暂存：MCP配置ID列表（JSON数组：[1,2,3]），创建会话时写入session
	SkillIds        string     `gorm:"column:skill_ids;type:text"`                     // 暂存：技能ID列表（JSON数组：[1,2,3]），创建会话时写入session
	Project         string     `gorm:"column:project;type:varchar(255);default:''"`    // 暂存：项目目录名称，创建会话时写入session
	SharedProjectId int        `gorm:"column:shared_project_id;index;default:0"`       // 暂存：团队项目ID，0表示个人项目，创建会话时写入session
	AIModelId       int        `gorm:"column:ai_model_id;index;default:0"`             // 暂存：使用的模型ID，创建会话时写入session
	PersonaId       int        `gorm:"column:persona_id;index;default:0"`              // 暂存：用户角色ID，0表示未选择，创建会话时写入session
	NextRunAt       time.Time  `gorm:"index;not null;column:next_run_at"`              // 下次执行时间（调度器扫描依据），创建/编辑时必算；停用或完成保留原值，靠status过滤
	Status          uint8      `gorm:"column:status;default:0;not null"`               // 任务状态：0停用 1启用 2已完成
	Deleted         uint8      `gorm:"column:deleted;default:0"`                       // 软删除：0正常 1删除
	Created         time.Time  `gorm:"not null;column:created"`
	Updated         time.Time  `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (t *AutomationTask) TableName() string {
	return "automation_task"
}

// BeforeCreate 创建时设置时间戳
func (t *AutomationTask) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.Created = now
	t.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (t *AutomationTask) BeforeSave(tx *gorm.DB) error {
	t.Updated = time.Now()
	return nil
}

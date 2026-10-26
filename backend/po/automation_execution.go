package po

import (
	"time"

	"gorm.io/gorm"
)

// AutomationExecution 自动化任务执行记录表
// 执行成功（会话完成）后才写入，仅作任务与会话的关联记录，供前端查看执行结果；
// 失败原因无需存储，session 与聊天记录中自然可见。
type AutomationExecution struct {
	Id               int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`       // 执行记录ID
	AutomationTaskId int       `gorm:"index;column:automation_task_id;not null"`          // 自动化任务ID
	SessionId        string    `gorm:"index;column:session_id;type:varchar(64);not null"` // 执行产生的会话ID
	StartedAt        time.Time `gorm:"not null;column:started_at"`                        // 实际开始执行时间
	FinishedAt       time.Time `gorm:"not null;column:finished_at"`                       // 实际完成时间
	Created          time.Time `gorm:"not null;column:created"`
	Updated          time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (e *AutomationExecution) TableName() string {
	return "automation_execution"
}

// BeforeCreate 创建时设置时间戳
func (e *AutomationExecution) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	e.Created = now
	e.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (e *AutomationExecution) BeforeSave(tx *gorm.DB) error {
	e.Updated = time.Now()
	return nil
}

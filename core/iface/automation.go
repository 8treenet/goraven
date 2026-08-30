package iface

import "goraven/backend/po"

// AutomationTaskRepo 自动化任务（定时任务）持久化接口
type AutomationTaskRepo interface {
	// CreateTask 创建自动化任务，仓储内部计算并写入 NextRunAt 后落库
	CreateTask(task *po.AutomationTask) error
	// ListEnabledTasks 分页查询用户启用中（status=1）且未删除的自动化任务，按 NextRunAt 升序；
	// title 非空时按标题模糊匹配（LIKE %title%）
	ListEnabledTasks(userId, title string, page, pageSize int) ([]po.AutomationTask, int64, error)
	// GetTask 按 ID 查询用户的自动化任务（校验归属，防止越权），含停用与已完成任务
	GetTask(id int, userId string) (*po.AutomationTask, error)
	// UpdateTask 全量更新任务业务字段（标题/需求/执行计划），next_run_at 随任务写入；
	// recomputeNext 为 true 时依据任务当前计划字段重算 NextRunAt
	UpdateTask(task *po.AutomationTask, recomputeNext bool) error
}

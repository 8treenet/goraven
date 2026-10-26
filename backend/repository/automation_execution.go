package repository

import (
	"goraven/backend/po"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *AutomationExecutionRepository {
			return &AutomationExecutionRepository{}
		})
	})
}

// AutomationExecutionRepository 自动化任务执行记录仓储
type AutomationExecutionRepository struct {
	freedom.Repository
}

// ListByTaskId 分页查询任务的执行记录，按 id 倒序（最新在前）
func (repo *AutomationExecutionRepository) ListByTaskId(taskId int, page, pageSize int) ([]po.AutomationExecution, *PageResult, error) {
	query := repo.db().Model(&po.AutomationExecution{}).
		Where("automation_task_id = ?", taskId).
		Order("id DESC")
	var records []po.AutomationExecution
	pr, err := Paginate(query, &PageQuery{Page: page, PageSize: pageSize}, &records)
	if err != nil {
		return nil, nil, err
	}
	return records, pr, nil
}

// GetById 按 ID 查询执行记录
func (repo *AutomationExecutionRepository) GetById(id int) (*po.AutomationExecution, error) {
	var record po.AutomationExecution
	if err := repo.db().First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// maxExecutionsPerTask 每个任务保留的执行记录上限，防止无限增长
const maxExecutionsPerTask = 50

// CreateExecution 写入执行记录（仅在执行成功、会话完成后调用），
// 并顺带清理旧记录，只保留每个任务最近 maxExecutionsPerTask 条
func (repo *AutomationExecutionRepository) CreateExecution(record *po.AutomationExecution) error {
	if err := repo.db().Create(record).Error; err != nil {
		return err
	}
	// 两步查询清理：先取最近 N 条的 ID 再删除其余，避免 MySQL 不支持的同表子查询删除
	var keepIds []int
	if err := repo.db().Model(&po.AutomationExecution{}).
		Where("automation_task_id = ?", record.AutomationTaskId).
		Order("id DESC").
		Limit(maxExecutionsPerTask).
		Pluck("id", &keepIds).Error; err != nil {
		return err
	}
	return repo.db().
		Where("automation_task_id = ? AND id NOT IN ?", record.AutomationTaskId, keepIds).
		Delete(&po.AutomationExecution{}).Error
}

func (repo *AutomationExecutionRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

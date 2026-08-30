package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goraven/backend/po"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *AutomationTaskRepository {
			return &AutomationTaskRepository{}
		})
	})
}

// AutomationTaskRepository 自动化任务（定时任务）仓储
type AutomationTaskRepository struct {
	freedom.Repository
}

// CreateTask 创建自动化任务：计算并写入 NextRunAt 后落库
func (repo *AutomationTaskRepository) CreateTask(task *po.AutomationTask) error {
	next, err := CalcNextRunAt(task, time.Now())
	if err != nil {
		return err
	}
	task.NextRunAt = next
	return repo.db().Create(task).Error
}

// ListEnabledTasks 分页查询用户启用中（status=1）且未删除的自动化任务，按 NextRunAt 升序；
// title 非空时按标题模糊匹配（LIKE %title%）
func (repo *AutomationTaskRepository) ListEnabledTasks(userId, title string, page, pageSize int) ([]po.AutomationTask, int64, error) {
	query := repo.db().Model(&po.AutomationTask{}).
		Where("user_id = ? AND status = ? AND deleted = 0", userId, po.AutomationStatusEnabled)
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	var tasks []po.AutomationTask
	pr, err := Paginate(query.Order("next_run_at ASC"), &PageQuery{Page: page, PageSize: pageSize}, &tasks)
	if err != nil {
		return nil, 0, err
	}
	return tasks, int64(pr.TotalCount), nil
}

// ListTasks 分页查询用户未删除的自动化任务（含停用/已完成），status 非 nil 时按状态过滤，按创建时间倒序
func (repo *AutomationTaskRepository) ListTasks(userId string, status *uint8, page, pageSize int) ([]po.AutomationTask, *PageResult, error) {
	query := repo.db().Model(&po.AutomationTask{}).
		Where("user_id = ? AND deleted = 0", userId)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query = query.Order("created DESC")
	var tasks []po.AutomationTask
	pr, err := Paginate(query, &PageQuery{Page: page, PageSize: pageSize}, &tasks)
	if err != nil {
		return nil, nil, err
	}
	return tasks, pr, nil
}

// GetTask 按 ID 查询用户的自动化任务（校验归属，防止越权）
func (repo *AutomationTaskRepository) GetTask(id int, userId string) (*po.AutomationTask, error) {
	var task po.AutomationTask
	if err := repo.db().First(&task, "id = ? AND user_id = ? AND deleted = 0", id, userId).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// SoftDelete 软删除任务（deleted=1），扫描条件天然过滤
func (repo *AutomationTaskRepository) SoftDelete(id int, userId string) error {
	return repo.db().Model(&po.AutomationTask{}).
		Where("id = ? AND user_id = ? AND deleted = 0", id, userId).
		Updates(map[string]interface{}{
			"deleted": 1,
			"updated": time.Now(),
		}).Error
}

// UpdateStatus 更新任务状态；nextRunAt 非 nil 时同时刷新下次执行时间
func (repo *AutomationTaskRepository) UpdateStatus(id int, userId string, status uint8, nextRunAt *time.Time) error {
	updates := map[string]interface{}{
		"status":  status,
		"updated": time.Now(),
	}
	if nextRunAt != nil {
		updates["next_run_at"] = *nextRunAt
	}
	return repo.db().Model(&po.AutomationTask{}).
		Where("id = ? AND user_id = ? AND deleted = 0", id, userId).
		Updates(updates).Error
}

// UpdateTask 全量更新任务业务字段（标题/需求/执行计划），next_run_at 随任务写入；
// recomputeNext 为 true 时先依据任务当前计划字段重算 NextRunAt。
// 任务不存在或已删除时报错（RowsAffected 为 0）。
func (repo *AutomationTaskRepository) UpdateTask(task *po.AutomationTask, recomputeNext bool) error {
	if recomputeNext {
		next, err := CalcNextRunAt(task, time.Now())
		if err != nil {
			return err
		}
		task.NextRunAt = next
	}
	result := repo.db().Model(&po.AutomationTask{}).
		Where("id = ? AND user_id = ? AND deleted = 0", task.Id, task.UserId).
		Updates(map[string]interface{}{
			"title":            task.Title,
			"requirement":      task.Requirement,
			"exec_type":        task.ExecType,
			"run_at":           task.RunAt,
			"interval_minutes": task.IntervalMinutes,
			"fixed_time":       task.FixedTime,
			"weekday":          task.Weekday,
			"next_run_at":      task.NextRunAt,
			"updated":          time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("automation task %d not found", task.Id)
	}
	return nil
}

// CalcNextRunAt 计算任务的下次执行时间（创建/编辑时初始化）
// now 传入当前时间便于测试；单次任务直接取 RunAt，间隔任务为 now+间隔，
// 每天/每周锚定最近的未来固定时间点。
func CalcNextRunAt(task *po.AutomationTask, now time.Time) (time.Time, error) {
	switch task.ExecType {
	case po.AutomationExecTypeOnce:
		if task.RunAt == nil {
			return time.Time{}, fmt.Errorf("RunAt is required for once-type automation task")
		}
		return *task.RunAt, nil
	case po.AutomationExecTypeInterval:
		if task.IntervalMinutes < 5 {
			return time.Time{}, fmt.Errorf("interval_minutes must be >= 5")
		}
		return now.Add(time.Duration(task.IntervalMinutes) * time.Minute), nil
	case po.AutomationExecTypeDaily:
		return nextFixedTime(task.FixedTime, nil, now)
	case po.AutomationExecTypeWeekly:
		if task.Weekday > 6 {
			return time.Time{}, fmt.Errorf("weekday must be in [0, 6]")
		}
		return nextFixedTime(task.FixedTime, &task.Weekday, now)
	default:
		return time.Time{}, fmt.Errorf("unknown exec_type: %d", task.ExecType)
	}
}

// nextFixedTime 计算下一个固定时间点（每天 weekday 为 nil；每周时 0=周日 1-6=周一至周六）
func nextFixedTime(fixed string, weekday *uint8, now time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", fixed)
	if err != nil {
		return time.Time{}, fmt.Errorf("fixed_time must be HH:MM format: %v", err)
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	if weekday == nil { // 每天
		if !next.After(now) {
			next = next.AddDate(0, 0, 1) // 今天已过，推到明天
		}
		return next, nil
	}
	// 每周：计算与目标星期的天数差
	diff := (int(*weekday) - int(now.Weekday()) + 7) % 7
	if diff == 0 && !next.After(now) {
		diff = 7 // 今天就是目标日但时刻已过，推到下周
	}
	return next.AddDate(0, 0, diff), nil
}

// automationTaskLockTTL 任务执行锁 TTL：5 分钟，执行完成后主动释放。
// 锁仅兜底防多实例/协程重复触发；周期任务靠执行前的占位推演防重复扫描。
const automationTaskLockTTL = 5 * time.Minute

func automationTaskLockKey(id int) string {
	return fmt.Sprintf("automation_task_lock:%d", id)
}

// ListDue 调度器扫描：查询到点待执行的启用任务（status=1 且未删除且 next_run_at <= now），
// 按 next_run_at 升序取前 limit 条。NextRunAt 为唯一调度依据。
func (repo *AutomationTaskRepository) ListDue(now time.Time, limit int) ([]po.AutomationTask, error) {
	var tasks []po.AutomationTask
	err := repo.db().Model(&po.AutomationTask{}).
		Where("status = ? AND deleted = 0 AND next_run_at <= ?", po.AutomationStatusEnabled, now).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// LockTask 对任务加执行锁（SetNX，TTL 5 分钟）。
// 返回 false 表示其他实例/协程正在执行，调用方直接跳过本轮。
func (repo *AutomationTaskRepository) LockTask(id int) (bool, error) {
	ok, err := repo.Redis().SetNX(context.Background(), automationTaskLockKey(id), time.Now().Unix(), automationTaskLockTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// UnlockTask 释放任务执行锁（执行结束后调用）
func (repo *AutomationTaskRepository) UnlockTask(id int) error {
	return repo.Redis().Del(context.Background(), automationTaskLockKey(id)).Err()
}

// IsTaskLocked 检查任务执行锁是否被占用（本地缓存与 Redis 模式均可用）
func (repo *AutomationTaskRepository) IsTaskLocked(id int) (bool, error) {
	_, err := repo.Redis().Get(context.Background(), automationTaskLockKey(id)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateNextRunAt 更新任务的下次执行时间（执行前占位推演 / 执行后精确推演）
func (repo *AutomationTaskRepository) UpdateNextRunAt(id int, next time.Time) error {
	return repo.db().Model(&po.AutomationTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"next_run_at": next,
			"updated":     time.Now(),
		}).Error
}

// MarkDone 单次任务加锁成功后立即置为已完成，防止任何重复触发；失败不再重试
func (repo *AutomationTaskRepository) MarkDone(id int) error {
	return repo.db().Model(&po.AutomationTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  po.AutomationStatusDone,
			"updated": time.Now(),
		}).Error
}

func (repo *AutomationTaskRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

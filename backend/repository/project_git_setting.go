package repository

import (
	"context"
	"fmt"
	"time"

	"goraven/backend/po"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *ProjectGitSettingRepository {
			return &ProjectGitSettingRepository{}
		})
	})
}

// ProjectGitSettingRepository 项目 Git 配置仓储（个人项目与团队项目共用）
type ProjectGitSettingRepository struct {
	freedom.Repository
}

func (repo *ProjectGitSettingRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

// GetByOwner 按项目查询 Git 配置，不存在时返回 gorm.ErrRecordNotFound
func (repo *ProjectGitSettingRepository) GetByOwner(ownerType uint8, ownerId int) (*po.ProjectGitSetting, error) {
	var setting po.ProjectGitSetting
	if err := repo.db().Where("owner_type = ? AND owner_id = ?", ownerType, ownerId).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// Upsert 保存 Git 配置（Id 为 0 时插入，否则更新）
func (repo *ProjectGitSettingRepository) Upsert(setting *po.ProjectGitSetting) error {
	return repo.db().Save(setting).Error
}

// DeleteByOwner 删除项目的 Git 配置（删除项目时级联调用）
func (repo *ProjectGitSettingRepository) DeleteByOwner(ownerType uint8, ownerId int) error {
	return repo.db().Where("owner_type = ? AND owner_id = ?", ownerType, ownerId).Delete(&po.ProjectGitSetting{}).Error
}

// ListAutoSync 列出所有 Git 项目配置（存在配置行即 Git 项目）
func (repo *ProjectGitSettingRepository) ListAutoSync() ([]po.ProjectGitSetting, error) {
	var settings []po.ProjectGitSetting
	err := repo.db().
		Order("id ASC").
		Find(&settings).Error
	return settings, err
}

// ListByOwners 批量查询项目 Git 配置，用于列表徽标
func (repo *ProjectGitSettingRepository) ListByOwners(ownerType uint8, ownerIds []int) (map[int]*po.ProjectGitSetting, error) {
	result := make(map[int]*po.ProjectGitSetting, len(ownerIds))
	if len(ownerIds) == 0 {
		return result, nil
	}
	var settings []po.ProjectGitSetting
	if err := repo.db().Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIds).Find(&settings).Error; err != nil {
		return nil, err
	}
	for i := range settings {
		result[settings[i].OwnerId] = &settings[i]
	}
	return result, nil
}

// UpdateCloneResult 写入克隆状态与失败原因（克隆中删除项目时静默无效）
func (repo *ProjectGitSettingRepository) UpdateCloneResult(ownerType uint8, ownerId int, state uint8, message string) error {
	return repo.db().Model(&po.ProjectGitSetting{}).
		Where("owner_type = ? AND owner_id = ?", ownerType, ownerId).
		Updates(map[string]interface{}{
			"clone_state":   state,
			"clone_message": message,
		}).Error
}

const projectGitLockTTL = 10 * time.Minute

func projectGitLockKey(ownerType uint8, ownerId int) string {
	return fmt.Sprintf("project_git_lock:%d:%d", ownerType, ownerId)
}

// LockOwner 对项目 Git 操作加锁（SetNX），获取失败表示已有操作进行中
func (repo *ProjectGitSettingRepository) LockOwner(ownerType uint8, ownerId int, token string, ttl time.Duration) (bool, error) {
	return repo.Redis().SetNX(context.Background(), projectGitLockKey(ownerType, ownerId), token, ttl).Result()
}

// UnlockOwner 释放项目 Git 操作锁
func (repo *ProjectGitSettingRepository) UnlockOwner(ownerType uint8, ownerId int) error {
	return repo.Redis().Del(context.Background(), projectGitLockKey(ownerType, ownerId)).Err()
}

package repository

import (
	"time"

	"goraven/backend/po"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *UserProjectRepository {
			return &UserProjectRepository{}
		})
	})
}

// UserProjectRepository 用户个人项目仓储
type UserProjectRepository struct {
	freedom.Repository
}

func (repo *UserProjectRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

// ListByUser 查询用户的个人项目，按更新时间倒序
func (repo *UserProjectRepository) ListByUser(userId string) ([]po.UserProject, error) {
	var projects []po.UserProject
	err := repo.db().Where("user_id = ?", userId).Order("updated DESC").Find(&projects).Error
	return projects, err
}

// GetByID 根据 ID 查询个人项目
func (repo *UserProjectRepository) GetByID(id int) (*po.UserProject, error) {
	var project po.UserProject
	if err := repo.db().First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName 根据用户和项目名查询个人项目
func (repo *UserProjectRepository) GetByName(userId, name string) (*po.UserProject, error) {
	var project po.UserProject
	if err := repo.db().Where("user_id = ? AND project_name = ?", userId, name).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// Create 创建个人项目记录
func (repo *UserProjectRepository) Create(project *po.UserProject) error {
	return repo.db().Create(project).Error
}

// Delete 删除个人项目记录
func (repo *UserProjectRepository) Delete(id int) error {
	return repo.db().Delete(&po.UserProject{}, "id = ?", id).Error
}

// UpdateDescription 更新项目简介
func (repo *UserProjectRepository) UpdateDescription(id int, description string) error {
	return repo.db().Model(&po.UserProject{}).Where("id = ?", id).
		Update("description", description).Error
}

// RenameProject 重命名个人项目并更新其个人会话和自动化任务引用
func (repo *UserProjectRepository) RenameProject(project *po.UserProject, newName string) error {
	updated := time.Now()
	err := repo.db().Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&po.UserProject{}).
			Where("id = ? AND user_id = ? AND project_name = ?", project.Id, project.UserId, project.ProjectName).
			Update("project_name", newName)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		where := "user_id = ? AND project = ? AND shared_project_id = 0"
		updates := map[string]interface{}{"project": newName, "updated": updated}
		if err := tx.Model(&po.Session{}).
			Where(where, project.UserId, project.ProjectName).
			Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&po.AutomationTask{}).
			Where(where, project.UserId, project.ProjectName).
			Updates(updates).Error
	})
	if err == nil {
		project.ProjectName = newName
	}
	return err
}

// ClearProjectReferences 清空个人项目被会话和自动化任务引用的目录名
func (repo *UserProjectRepository) ClearProjectReferences(userId, projectName string) error {
	updated := time.Now()
	return repo.db().Transaction(func(tx *gorm.DB) error {
		where := "user_id = ? AND project = ? AND shared_project_id = 0"
		updates := map[string]interface{}{"project": "", "updated": updated}
		if err := tx.Model(&po.Session{}).
			Where(where, userId, projectName).
			Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&po.AutomationTask{}).
			Where(where, userId, projectName).
			Updates(updates).Error
	})
}

// GetUsernameByUserID 根据用户 ID 查询用户名，包括已删除用户
func (repo *UserProjectRepository) GetUsernameByUserID(userId string) (string, error) {
	var user po.User
	err := repo.db().Select("username").Where("user_id = ?", userId).First(&user).Error
	return user.Username, err
}

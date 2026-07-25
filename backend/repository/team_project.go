package repository

import (
	"context"
	"fmt"
	"goraven/backend/po"
	"goraven/backend/vo"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *SharedProjectRepository {
			return &SharedProjectRepository{}
		})
	})
}

type SharedProjectRepository struct {
	freedom.Repository
}

func (repo *SharedProjectRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

func (repo *SharedProjectRepository) ListAll() ([]po.SharedProject, error) {
	var projects []po.SharedProject
	err := repo.db().Order("updated DESC").Find(&projects).Error
	return projects, err
}

func (repo *SharedProjectRepository) PaginateSharedProjects(req *vo.AdminSharedProjectListReq) ([]po.SharedProject, *PageResult, error) {
	query := repo.db().Model(&po.SharedProject{})
	if req.Search != "" {
		query = query.Where("project_name LIKE ?", "%"+req.Search+"%")
	}
	var projects []po.SharedProject
	pr, err := Paginate(query.Order("updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &projects)
	if err != nil {
		return nil, nil, err
	}
	return projects, pr, nil
}

func (repo *SharedProjectRepository) GetByID(id int) (*po.SharedProject, error) {
	var project po.SharedProject
	err := repo.db().First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (repo *SharedProjectRepository) ListByOwner(ownerId string) ([]po.SharedProject, error) {
	var projects []po.SharedProject
	err := repo.db().Where("owner_id = ?", ownerId).Find(&projects).Error
	return projects, err
}

func (repo *SharedProjectRepository) ListByIDs(ids []int) (map[int]po.SharedProject, error) {
	if len(ids) == 0 {
		return map[int]po.SharedProject{}, nil
	}
	var projects []po.SharedProject
	err := repo.db().Where("id IN ?", ids).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]po.SharedProject, len(projects))
	for i := range projects {
		m[projects[i].Id] = projects[i]
	}
	return m, nil
}

func (repo *SharedProjectRepository) GetByOwnerAndProject(ownerId, projectName string) (*po.SharedProject, error) {
	var project po.SharedProject
	err := repo.db().Where("owner_id = ? AND project_name = ?", ownerId, projectName).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (repo *SharedProjectRepository) Create(project *po.SharedProject) error {
	return repo.db().Create(project).Error
}

func (repo *SharedProjectRepository) Delete(id int) error {
	return repo.db().Delete(&po.SharedProject{}, "id = ?", id).Error
}

func (repo *SharedProjectRepository) UpdateDescription(id int, description string) error {
	return repo.db().Model(&po.SharedProject{}).Where("id = ?", id).
		Update("description", description).Error
}

func (repo *SharedProjectRepository) GetUserByID(userId string) (*po.User, error) {
	var user po.User
	err := repo.db().Where("user_id = ? AND deleted = 0", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *SharedProjectRepository) GetUsersByIDs(userIds []string) (map[string]*po.User, error) {
	if len(userIds) == 0 {
		return map[string]*po.User{}, nil
	}
	var users []po.User
	err := repo.db().Where("user_id IN ? AND deleted = 0", userIds).Find(&users).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]*po.User, len(users))
	for i := range users {
		m[users[i].UserId] = &users[i]
	}
	return m, nil
}

const sharedProjectLockTTL = 30 * time.Minute

func sharedProjectLockKey(id int) string {
	return fmt.Sprintf("shared_project_lock:%d", id)
}

func (repo *SharedProjectRepository) LockSharedProject(id int, sessionId string) (bool, error) {
	ctx := context.Background()
	ok, err := repo.Redis().SetNX(ctx, sharedProjectLockKey(id), sessionId, sharedProjectLockTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (repo *SharedProjectRepository) UnlockSharedProject(id int) error {
	ctx := context.Background()
	return repo.Redis().Del(ctx, sharedProjectLockKey(id)).Err()
}

func (repo *SharedProjectRepository) IncrementVisitAndUpdateLastActive(id int) error {
	return repo.db().Model(&po.SharedProject{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"visit_count":    gorm.Expr("visit_count + 1"),
			"last_active_at": time.Now(),
		}).Error
}

func (repo *SharedProjectRepository) IsSharedProjectLocked(id int) (bool, string, error) {
	ctx := context.Background()
	val, err := repo.Redis().Get(ctx, sharedProjectLockKey(id)).Result()
	if err != nil {
		return false, "", nil
	}
	return true, val, nil
}

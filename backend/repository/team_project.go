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
		initiator.BindRepository(func() *TeamProjectRepository {
			return &TeamProjectRepository{}
		})
	})
}

// TeamProjectRepository 团队项目仓储
type TeamProjectRepository struct {
	freedom.Repository
}

func (repo *TeamProjectRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

// ListAll 查询所有团队项目，按更新时间倒序
func (repo *TeamProjectRepository) ListAll() ([]po.TeamProject, error) {
	var projects []po.TeamProject
	err := repo.db().Order("updated DESC").Find(&projects).Error
	return projects, err
}

// Paginate 分页查询团队项目（可按项目名模糊搜索），按更新时间倒序
func (repo *TeamProjectRepository) Paginate(req *vo.AdminTeamProjectListReq) ([]po.TeamProject, *PageResult, error) {
	query := repo.db().Model(&po.TeamProject{})
	if req.Search != "" {
		query = query.Where("project_name LIKE ?", "%"+req.Search+"%")
	}
	var projects []po.TeamProject
	pr, err := Paginate(query.Order("updated DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &projects)
	if err != nil {
		return nil, nil, err
	}
	return projects, pr, nil
}

// GetByID 根据 ID 查询团队项目
func (repo *TeamProjectRepository) GetByID(id int) (*po.TeamProject, error) {
	var project po.TeamProject
	err := repo.db().First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByIDs 根据 ID 列表批量查询团队项目
func (repo *TeamProjectRepository) GetByIDs(ids []int) ([]po.TeamProject, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var projects []po.TeamProject
	err := repo.db().Where("id IN ?", ids).Find(&projects).Error
	return projects, err
}

// GetByName 根据项目名查询团队项目
func (repo *TeamProjectRepository) GetByName(name string) (*po.TeamProject, error) {
	var project po.TeamProject
	err := repo.db().Where("project_name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListByIDs 根据 ID 列表批量查询团队项目，返回 id -> TeamProject 的映射
func (repo *TeamProjectRepository) ListByIDs(ids []int) (map[int]po.TeamProject, error) {
	if len(ids) == 0 {
		return map[int]po.TeamProject{}, nil
	}
	var projects []po.TeamProject
	err := repo.db().Where("id IN ?", ids).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]po.TeamProject, len(projects))
	for i := range projects {
		m[projects[i].Id] = projects[i]
	}
	return m, nil
}

// Create 创建团队项目记录
func (repo *TeamProjectRepository) Create(project *po.TeamProject) error {
	return repo.db().Create(project).Error
}

// Delete 删除团队项目记录
func (repo *TeamProjectRepository) Delete(id int) error {
	return repo.db().Delete(&po.TeamProject{}, "id = ?", id).Error
}

// UpdateDescription 更新简介
func (repo *TeamProjectRepository) UpdateDescription(id int, description string) error {
	return repo.db().Model(&po.TeamProject{}).Where("id = ?", id).
		Update("description", description).Error
}

// UpdateAccess 更新访问权限
func (repo *TeamProjectRepository) UpdateAccess(id int, access uint8) error {
	return repo.db().Model(&po.TeamProject{}).Where("id = ?", id).
		Update("access", access).Error
}

// GetUserByID 根据 userId 查询用户信息（用于获取昵称、账号和头像）
func (repo *TeamProjectRepository) GetUserByID(userId string) (*po.User, error) {
	var user po.User
	err := repo.db().Where("user_id = ? AND deleted = 0", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersByIDs 批量查询用户信息（用于 list 时填充 creatorName / creatorAvatar）
func (repo *TeamProjectRepository) GetUsersByIDs(userIds []string) (map[string]*po.User, error) {
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

const teamProjectLockTTL = 30 * time.Minute

func teamProjectLockKey(id int) string {
	return fmt.Sprintf("team_project_lock:%d", id)
}

// LockTeamProject 对团队项目加锁（SetNX），同一时刻只能有一个 Agent 操作该项目
// 返回 true 表示加锁成功
func (repo *TeamProjectRepository) LockTeamProject(id int, sessionId string) (bool, error) {
	ctx := context.Background()
	ok, err := repo.Redis().SetNX(ctx, teamProjectLockKey(id), sessionId, teamProjectLockTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// UnlockTeamProject 解除团队项目锁
func (repo *TeamProjectRepository) UnlockTeamProject(id int) error {
	ctx := context.Background()
	return repo.Redis().Del(ctx, teamProjectLockKey(id)).Err()
}

// IncrementVisitAndUpdateLastActive 访问次数 +1 并更新最近活跃时间
func (repo *TeamProjectRepository) IncrementVisitAndUpdateLastActive(id int) error {
	return repo.db().Model(&po.TeamProject{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"visit_count":      gorm.Expr("visit_count + 1"),
			"last_active_time": time.Now(),
		}).Error
}

// IsTeamProjectLocked 检查团队项目是否被锁定，返回锁状态和占用者 sessionId
func (repo *TeamProjectRepository) IsTeamProjectLocked(id int) (bool, string, error) {
	ctx := context.Background()
	val, err := repo.Redis().Get(ctx, teamProjectLockKey(id)).Result()
	if err != nil {
		return false, "", nil
	}
	return true, val, nil
}

// --- 成员管理 ---

// ListByUser 查询用户可见的团队项目（全员开放 OR 创建者 OR 成员），按更新时间倒序
func (repo *TeamProjectRepository) ListByUser(userId string) ([]po.TeamProject, error) {
	var projects []po.TeamProject
	err := repo.db().
		Where("access = ? OR creator_id = ? OR id IN (?)",
			po.TeamProjectAccessAll,
			userId,
			repo.db().Model(&po.TeamProjectMember{}).Select("project_id").Where("user_id = ?", userId),
		).
		Order("updated DESC").
		Find(&projects).Error
	return projects, err
}

// ListMembers 查询项目的所有成员
func (repo *TeamProjectRepository) ListMembers(projectId int) ([]po.TeamProjectMember, error) {
	var members []po.TeamProjectMember
	err := repo.db().Where("project_id = ?", projectId).Order("created ASC").Find(&members).Error
	return members, err
}

// AddMember 添加项目成员
func (repo *TeamProjectRepository) AddMember(projectId int, userId string) error {
	member := &po.TeamProjectMember{
		ProjectId: projectId,
		UserId:    userId,
	}
	return repo.db().Create(member).Error
}

// RemoveMember 移除项目成员
func (repo *TeamProjectRepository) RemoveMember(projectId int, userId string) error {
	return repo.db().Where("project_id = ? AND user_id = ?", projectId, userId).
		Delete(&po.TeamProjectMember{}).Error
}

// RemoveMembersByProjectId 删除项目的所有成员记录（项目删除时调用）
func (repo *TeamProjectRepository) RemoveMembersByProjectId(projectId int) error {
	return repo.db().Where("project_id = ?", projectId).Delete(&po.TeamProjectMember{}).Error
}

// IsMember 检查用户是否为项目成员
func (repo *TeamProjectRepository) IsMember(projectId int, userId string) (bool, error) {
	var count int64
	err := repo.db().Model(&po.TeamProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectId, userId).
		Count(&count).Error
	return count > 0, err
}

// PaginateActiveUsers 分页查询所有可用用户（成员选择器用）
func (repo *TeamProjectRepository) PaginateActiveUsers(req *vo.TeamProjectUserListReq) ([]vo.TeamProjectUserItem, *PageResult, error) {
	var users []po.User
	pr, err := Paginate(
		repo.db().Model(&po.User{}).Where("deleted = 0 AND status = 1").Order("username ASC"),
		&PageQuery{Page: req.Page, PageSize: req.PageSize},
		&users,
	)
	if err != nil {
		return nil, nil, err
	}
	items := make([]vo.TeamProjectUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, vo.TeamProjectUserItem{
			UserId:   u.UserId,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
		})
	}
	return items, pr, nil
}

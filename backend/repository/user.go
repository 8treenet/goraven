package repository

import (
	"context"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/vo"
	"strconv"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

const (
	LoginFailKey           = "auth:login_fail"
	CaptchaAnswerKeyPrefix = "auth:captcha:"
	LoginFailWindow        = 5 * time.Minute
	CaptchaTTL             = 5 * time.Minute
	CaptchaThreshold       = 30
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *UserRepository {
			return &UserRepository{}
		})
	})
}

// UserRepository 用户仓储
type UserRepository struct {
	freedom.Repository
	Auth *infra.Auth
}

func (repo *UserRepository) FindByUsername(username string) (*po.User, error) {
	var user po.User
	err := repo.db().First(&user, "username = ? AND deleted = 0", username).Error
	return &user, err
}

func (repo *UserRepository) FindByUserId(userId string) (*po.User, error) {
	var user po.User
	err := repo.db().First(&user, "user_id = ? AND deleted = 0", userId).Error
	return &user, err
}

func (repo *UserRepository) FindSuperAdmin() (*po.User, error) {
	var user po.User
	err := repo.db().First(&user, "super_admin = ? AND deleted = 0", po.UserIsSuperAdmin).Error
	return &user, err
}

// PaginateUsers 分页查询用户列表
// 同时 JOIN session 表统计会话数和最后活跃时间
func (repo *UserRepository) PaginateUsers(req *vo.AdminUserListReq) ([]vo.AdminUserItem, *PageResult, error) {
	query := repo.db().Model(&po.User{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("username LIKE ?", "%"+req.Search+"%")
	}
	if req.Role != nil {
		query = query.Where("role = ?", *req.Role)
	}

	var users []po.User
	pr, err := Paginate(query.Order("created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &users)
	if err != nil {
		return nil, nil, err
	}

	items := make([]vo.AdminUserItem, 0, len(users))
	for _, u := range users {
		sessionCount, lastActiveTime := repo.GetUserSessionStats(u.UserId)

		items = append(items, vo.AdminUserItem{
			UserId:          u.UserId,
			Username:        u.Username,
			Nickname:        u.Nickname,
			Email:           u.Email,
			Avatar:          u.Avatar,
			Role:            u.Role,
			Status:          u.Status,
			DailyTokenLimit: u.DailyTokenLimit,
			SessionCount:    int(sessionCount),
			LastActiveTime:  lastActiveTime,
			Created:         u.Created,
		})
	}

	return items, pr, nil
}

// GetUserSessionStats 获取用户会话统计（会话数 + 最后活跃时间）
func (repo *UserRepository) GetUserSessionStats(userId string) (int, *time.Time) {
	var sessionCount int64
	repo.db().Model(&po.Session{}).Where("user_id = ? AND deleted = 0", userId).Count(&sessionCount)
	var lastActiveTime *time.Time
	if sessionCount > 0 {
		var latest po.Session
		if err := repo.db().Model(&po.Session{}).
			Where("user_id = ? AND deleted = 0", userId).
			Select("last_chat_time").
			Order("last_chat_time DESC").
			First(&latest).Error; err == nil {
			lastActiveTime = &latest.LastChatTime
		}
	}
	return int(sessionCount), lastActiveTime
}

// CreateUser 创建用户
func (repo *UserRepository) CreateUser(user *po.User) error {
	return repo.db().Create(user).Error
}

// UpdateUser 更新用户信息（昵称、邮箱、角色、状态、日限额）
func (repo *UserRepository) UpdateUser(user *po.User) error {
	return repo.db().Model(&po.User{}).Where("user_id = ? AND deleted = 0", user.UserId).Updates(map[string]interface{}{
		"nickname":          user.Nickname,
		"email":             user.Email,
		"role":              user.Role,
		"status":            user.Status,
		"daily_token_limit": user.DailyTokenLimit,
	}).Error
}

// UpdatePassword 更新用户密码
func (repo *UserRepository) UpdatePassword(userId, password string) error {
	return repo.db().Model(&po.User{}).Where("user_id = ? AND deleted = 0", userId).Update("password", password).Error
}

// FindByUserIds 根据 userId 列表批量查询用户
func (repo *UserRepository) FindByUserIds(userIds []string) ([]po.User, error) {
	if len(userIds) == 0 {
		return nil, nil
	}
	var users []po.User
	err := repo.db().Where("user_id IN ? AND deleted = 0", userIds).Find(&users).Error
	return users, err
}

// SoftDeleteUser 软删除用户（deleted = 1）
func (repo *UserRepository) SoftDeleteUser(userId string) error {
	return repo.db().Model(&po.User{}).Where("user_id = ? AND deleted = 0", userId).Update("deleted", 1).Error
}

// UpdateProfile 更新用户个人资料（仅更新传入的字段）
func (repo *UserRepository) UpdateProfile(userId string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return repo.db().Model(&po.User{}).Where("user_id = ? AND deleted = 0", userId).Updates(updates).Error
}

func (repo *UserRepository) GenerateToken(userId, clientIP, clientUA string) (string, error) {
	return repo.Auth.GenerateAccessToken(userId, time.Now().AddDate(0, 0, 7), clientIP, clientUA)
}

func (repo *UserRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

// RecordLoginFailure 全局密码错误后累加计数，返回累加后的当前值
// 实现：先读当前值 +1 后写回并刷新 TTL；本地缓存模式与 Redis 模式语义一致
// 全局计数：企业内应用，所有用户名共享同一个计数器
func (repo *UserRepository) RecordLoginFailure() int64 {
	ctx := context.Background()
	current := repo.GetLoginFailureCount()
	current++
	repo.Redis().Set(ctx, LoginFailKey, strconv.FormatInt(current, 10), LoginFailWindow)
	return current
}

// GetLoginFailureCount 读取全局密码错误累计次数，无值返回 0
func (repo *UserRepository) GetLoginFailureCount() int64 {
	v, err := repo.Redis().Get(context.Background(), LoginFailKey).Result()
	if err != nil {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// SaveCaptchaAnswer 下发验证码后保存正确答案（按用户名隔离，一次性，5 分钟过期）
func (repo *UserRepository) SaveCaptchaAnswer(username string, answer int) {
	repo.Redis().Set(context.Background(), CaptchaAnswerKeyPrefix+username, strconv.Itoa(answer), CaptchaTTL)
}

// PopCaptchaAnswer 校验时取出并删除该用户的答案，返回正确答案与是否存在
func (repo *UserRepository) PopCaptchaAnswer(username string) (int, bool) {
	ctx := context.Background()
	key := CaptchaAnswerKeyPrefix + username
	v, err := repo.Redis().Get(ctx, key).Result()
	if err != nil {
		return 0, false
	}
	repo.Redis().Del(ctx, key)
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/vo"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

const (
	// 登录失败计数按「账号」维度隔离。不再使用全局单键：避免一个人刷错密码
	// 就让全站强制验证码 / 无法登录。
	LoginFailUserKeyPrefix = "auth:login_fail:u:"
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

// scopeDigest 把账号归一化成定长十六进制，避免把用户名直接写进 Redis key。
func scopeDigest(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) > 128 {
		s = s[:128]
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:16])
}

// loginFailKey 返回该账号的失败计数 key。
//
// 刻意不把来源 IP 作为共享计数维度：应用层看到的 RemoteAddr 在反向代理后
// 往往恒为网关地址，所有账号共享同一个 IP key 会重新退化成「全站单计数器」，
// 正是本补丁要消除的 DoS 根因。按 IP 的限流应放在可信的网关/反代层。
// clientIP 参数保留给未来的可插拔限流策略，当前不参与判定。
func (repo *UserRepository) loginFailKey(username string, clientIP string) string {
	return LoginFailUserKeyPrefix + scopeDigest(username)
}

// loginFailMu 保证「读-加-写」在进程内原子。
var loginFailMu sync.Mutex

// RecordLoginFailure 累加该账号的失败次数并返回当前值。
//
// 注意：刻意不用 Redis INCR / EXPIRE。freedom 的本地缓存（cache_type=local）
// 只实现了 redis.Cmdable 的部分方法，调用 Incr() 会直接 nil-pointer panic
// （不是返回 error，调用方无法兜住）。统一改用 Get + Set（带 TTL），
// 并用进程级互斥锁保证原子。Redis 模式下多实例并发最多让计数略偏低，
// 只会让防护稍晚触发，不会误伤正常用户。
func (repo *UserRepository) RecordLoginFailure(username, clientIP string) int64 {
	loginFailMu.Lock()
	defer loginFailMu.Unlock()

	ctx := context.Background()
	key := repo.loginFailKey(username, clientIP)
	var current int64
	if v, err := repo.Redis().Get(ctx, key).Result(); err == nil {
		current, _ = strconv.ParseInt(v, 10, 64)
	}
	current++
	if err := repo.Redis().Set(ctx, key, strconv.FormatInt(current, 10), LoginFailWindow).Err(); err != nil {
		return 0
	}
	return current
}

// GetLoginFailureCount 读取该账号的失败次数。
func (repo *UserRepository) GetLoginFailureCount(username, clientIP string) int64 {
	key := repo.loginFailKey(username, clientIP)
	v, err := repo.Redis().Get(context.Background(), key).Result()
	if err != nil {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// ResetLoginFailure 登录成功后清除该账号的计数。
func (repo *UserRepository) ResetLoginFailure(username, clientIP string) {
	repo.Redis().Del(context.Background(), repo.loginFailKey(username, clientIP))
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

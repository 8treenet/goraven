package po

import (
	"time"

	"gorm.io/gorm"
)

// 用户角色常量
const (
	UserRoleUser  = 0 // 普通用户
	UserRoleAdmin = 1 // 管理员
)

// 用户状态常量
const (
	UserStatusDisabled = 0 // 禁用
	UserStatusEnabled  = 1 // 启用
)

// 超级管理员常量
const (
	UserNotSuperAdmin = 0 // 非超级管理员
	UserIsSuperAdmin  = 1 // 超级管理员
)

// User 用户表
// 管理员仅在系统初始化时注册，全局唯一。
// 管理员可通过后台添加其他普通用户，输入账号密码即可。
type User struct {
	UserId     string    `gorm:"primaryKey;column:user_id;type:varchar(64)"`            // 主键唯一ID，与 session.userId 保持一致
	Username   string    `gorm:"uniqueIndex;column:username;type:varchar(64);not null"` // 登录账号，全局唯一
	Password   string    `gorm:"column:password;type:varchar(256);not null"`            // 密码（存储 MD5 哈希值，由后端统一哈希，禁止明文）
	Email      string    `gorm:"column:email;type:varchar(128)"`                        // 邮箱
	Role       uint8     `gorm:"column:role;default:0;not null"`                        // 角色：0普通用户 1管理员
	SuperAdmin uint8     `gorm:"column:super_admin;default:0;not null"`                 // 超级管理员：0否 1是
	Status     uint8     `gorm:"column:status;default:1;not null"`                      // 状态：0禁用 1启用
	Nickname        string    `gorm:"column:nickname;type:varchar(128)"`                     // 显示昵称（为空时前端可 fallback 到 username）
	Avatar          string    `gorm:"column:avatar;type:varchar(512)"`                       // 头像 URL 或 Nickname首字母头像(未设置就Username首字母)
	DailyTokenLimit int       `gorm:"column:daily_token_limit;default:0;not null"`           // 每日 Token 限额（单位：M，0 表示不限制）
	Deleted         uint8     `gorm:"column:deleted;default:0"`                              // 软删除：0正常 1已删除
	Created    time.Time `gorm:"not null;column:created"`
	Updated    time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (u *User) TableName() string {
	return "user"
}

// BeforeCreate 创建时设置时间戳
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	u.Created = now
	u.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (u *User) BeforeSave(tx *gorm.DB) error {
	u.Updated = time.Now()
	return nil
}

// IsAdmin 判断是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsSuperAdmin 判断是否为超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.SuperAdmin == UserIsSuperAdmin
}

// IsEnabled 判断账号是否启用
func (u *User) IsEnabled() bool {
	return u.Status == UserStatusEnabled
}

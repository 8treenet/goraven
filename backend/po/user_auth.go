package po

import (
	"time"

	"gorm.io/gorm"
)

// UserAuth 用户登录鉴权表
// 用于存储用户登录后的 AccessToken，支持多端登录和 Token 过期管理。
type UserAuth struct {
	AuthId      int       `gorm:"primaryKey;column:auth_id;autoIncrement"`               // 主键ID
	UserId      string    `gorm:"index;column:user_id;type:varchar(64);not null"`        // 关联用户ID
	AccessToken string    `gorm:"uniqueIndex;column:access_token;type:varchar(128);not null"` // AccessToken，如 rvn_xxxx...
	ExpiresAt   time.Time `gorm:"column:expires_at"`                                      // 过期时间，零值表示永不过期
	ClientIP    string    `gorm:"column:client_ip;type:varchar(64)"`                     // 登录时客户端IP
	ClientUA    string    `gorm:"column:client_ua;type:varchar(512)"`                    // 登录时User-Agent
	Created     time.Time `gorm:"not null;column:created"`
	Updated     time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (ua *UserAuth) TableName() string {
	return "user_auth"
}

// BeforeCreate 创建时设置时间戳
func (ua *UserAuth) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	ua.Created = now
	ua.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (ua *UserAuth) BeforeSave(tx *gorm.DB) error {
	ua.Updated = time.Now()
	return nil
}

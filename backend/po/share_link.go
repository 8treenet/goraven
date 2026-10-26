package po

import (
	"time"

	"gorm.io/gorm"
)

// ShareLink 对话分享链接表
// 存储用户创建的对话分享链接，支持公开(public)和内部(internal)两种类型
type ShareLink struct {
	ShareId   string    `gorm:"primaryKey;column:share_id;type:varchar(64)"`          // UUID，URL 中的唯一标识
	UserId    string    `gorm:"index;column:user_id;type:varchar(64);not null"`       // 创建者
	SessionId string    `gorm:"index;column:session_id;type:varchar(64);not null"`    // 被分享的会话
	Title     string    `gorm:"column:title;type:varchar(255)"`                       // 分享标题
	ShareType string    `gorm:"column:share_type;type:varchar(16);default:public;not null"` // 分享类型：public(公开) / internal(内部)
	ExpiresAt time.Time `gorm:"column:expires_at"`                                    // 过期时间（零值表示永不过期）
	ViewCount int       `gorm:"column:view_count;default:0"`                          // 浏览次数
	Deleted   uint8     `gorm:"column:deleted;default:0"`                             // 软删除：0正常 1删除
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (s *ShareLink) TableName() string {
	return "share_link"
}

// BeforeCreate 创建时设置时间戳
func (s *ShareLink) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (s *ShareLink) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

// IsExpired 检查是否已过期
func (s *ShareLink) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

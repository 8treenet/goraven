package po

import (
	"time"

	"gorm.io/gorm"
)

type ShareLink struct {
	ShareId		string		`gorm:"primaryKey;column:share_id;type:varchar(64)"`
	UserId		string		`gorm:"index;column:user_id;type:varchar(64);not null"`
	SessionId	string		`gorm:"index;column:session_id;type:varchar(64);not null"`
	Title		string		`gorm:"column:title;type:varchar(255)"`
	ExpiresAt	time.Time	`gorm:"column:expires_at"`
	ViewCount	int		`gorm:"column:view_count;default:0"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (s *ShareLink) TableName() string {
	return "share_link"
}

func (s *ShareLink) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

func (s *ShareLink) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

func (s *ShareLink) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

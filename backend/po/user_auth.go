package po

import (
	"time"

	"gorm.io/gorm"
)

type UserAuth struct {
	AuthId		int		`gorm:"primaryKey;column:auth_id;autoIncrement"`
	UserId		string		`gorm:"index;column:user_id;type:varchar(64);not null"`
	AccessToken	string		`gorm:"uniqueIndex;column:access_token;type:varchar(128);not null"`
	ExpiresAt	time.Time	`gorm:"column:expires_at"`
	ClientIP	string		`gorm:"column:client_ip;type:varchar(64)"`
	ClientUA	string		`gorm:"column:client_ua;type:varchar(512)"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (ua *UserAuth) TableName() string {
	return "user_auth"
}

func (ua *UserAuth) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	ua.Created = now
	ua.Updated = now
	return nil
}

func (ua *UserAuth) BeforeSave(tx *gorm.DB) error {
	ua.Updated = time.Now()
	return nil
}

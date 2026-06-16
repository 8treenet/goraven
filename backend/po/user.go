package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	UserRoleUser	= 0
	UserRoleAdmin	= 1
)

const (
	UserStatusDisabled	= 0
	UserStatusEnabled	= 1
)

const (
	UserNotSuperAdmin	= 0
	UserIsSuperAdmin	= 1
)

type User struct {
	UserId		string		`gorm:"primaryKey;column:user_id;type:varchar(64)"`
	Username	string		`gorm:"uniqueIndex;column:username;type:varchar(64);not null"`
	Password	string		`gorm:"column:password;type:varchar(256);not null"`
	Email		string		`gorm:"column:email;type:varchar(128)"`
	Role		uint8		`gorm:"column:role;default:0;not null"`
	SuperAdmin	uint8		`gorm:"column:super_admin;default:0;not null"`
	Status		uint8		`gorm:"column:status;default:1;not null"`
	Nickname	string		`gorm:"column:nickname;type:varchar(128)"`
	Avatar		string		`gorm:"column:avatar;type:varchar(512)"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	u.Created = now
	u.Updated = now
	return nil
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	u.Updated = time.Now()
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

func (u *User) IsSuperAdmin() bool {
	return u.SuperAdmin == UserIsSuperAdmin
}

func (u *User) IsEnabled() bool {
	return u.Status == UserStatusEnabled
}

package po

import (
	"time"

	"gorm.io/gorm"
)

type UserPersona struct {
	PersonaId	int		`gorm:"primaryKey;column:persona_id;type:int;autoIncrement"`
	UserId		string		`gorm:"uniqueIndex:idx_user_persona;column:user_id;type:varchar(64);not null"`
	Name		string		`gorm:"uniqueIndex:idx_user_persona;column:name;type:varchar(64);not null"`
	Icon		string		`gorm:"column:icon;type:varchar(256)"`
	RoleInfo	string		`gorm:"column:role_info;type:text"`
	CategoryId	int		`gorm:"column:category_id;type:int;default:0"`
	AIModelId	int		`gorm:"column:ai_model_id;default:0"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
}

func (p *UserPersona) TableName() string {
	return "user_persona"
}

func (p *UserPersona) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

func (p *UserPersona) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}

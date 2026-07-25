package po

import (
	"time"

	"gorm.io/gorm"
)

type PersonaTool struct {
	Id        int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`
	PersonaId int       `gorm:"uniqueIndex:idx_persona_type_tool;column:persona_id;type:int;not null"`
	UserId    string    `gorm:"index:idx_user_type_tool;column:user_id;type:varchar(64);not null"`
	ToolType  string    `gorm:"uniqueIndex:idx_persona_type_tool;index:idx_user_type_tool;column:tool_type;type:varchar(16);not null"`
	ToolId    int       `gorm:"uniqueIndex:idx_persona_type_tool;index:idx_user_type_tool;column:tool_id;type:int;not null"`
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

func (PersonaTool) TableName() string {
	return "persona_tool"
}

func (p *PersonaTool) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

func (p *PersonaTool) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}

package po

import (
	"time"

	"gorm.io/gorm"
)

// UserPersona 用户角色配置表（每用户可创建多个角色，每个角色独立配置模型、MCP、技能）
type UserPersona struct {
	PersonaId  int    `gorm:"primaryKey;column:persona_id;type:int;autoIncrement"`                 // 主键ID
	UserId     string `gorm:"uniqueIndex:idx_user_persona;column:user_id;type:varchar(64);not null"` // 用户ID
	Name       string `gorm:"uniqueIndex:idx_user_persona;column:name;type:varchar(64);not null"`  // 角色名称（同一用户下唯一）
	Icon       string `gorm:"column:icon;type:varchar(256)"`                                        // 图标：Lucide 图标名称或 URL
	RoleInfo   string `gorm:"column:role_info;type:text"`                                            // 角色设定文案
	CategoryId int       `gorm:"column:category_id;type:int;default:0"`                                 // 角色分类ID，关联 persona_category 表
	AIModelId  int       `gorm:"column:ai_model_id;default:0"`                                           // 关联模型ID，0表示使用用户默认模型
	Deleted    uint8     `gorm:"column:deleted;default:0"`                                             // 软删除：0正常 1删除
	Created    time.Time `gorm:"not null;column:created"`
	Updated    time.Time `gorm:"not null;column:updated"`
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

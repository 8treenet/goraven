package po

import (
	"time"

	"gorm.io/gorm"
)

// PersonaTool 角色关联工具表（替代 user_persona.mcpIds / skillIds 的 JSON 存储）
// 每条记录表示一个角色绑定的一个 MCP 端点或用户技能
type PersonaTool struct {
	Id        int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`                                                          // 主键ID
	PersonaId int       `gorm:"uniqueIndex:idx_persona_type_tool;column:persona_id;type:int;not null"`                                 // 角色ID，关联 user_persona.personaId
	UserId    string    `gorm:"index:idx_user_type_tool;column:user_id;type:varchar(64);not null"`                                     // 用户ID（冗余，方便按用户级联查询和删除）
	ToolType  string    `gorm:"uniqueIndex:idx_persona_type_tool;index:idx_user_type_tool;column:tool_type;type:varchar(16);not null"` // 类型: mcp / skill
	ToolId    int       `gorm:"uniqueIndex:idx_persona_type_tool;index:idx_user_type_tool;column:tool_id;type:int;not null"`           // MCP端点ID（toolType=mcp）或用户技能ID（toolType=skill）
	Created   time.Time `gorm:"not null;column:created"`
	Updated   time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (PersonaTool) TableName() string {
	return "persona_tool"
}

// BeforeCreate 创建时设置时间戳
func (p *PersonaTool) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (p *PersonaTool) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}

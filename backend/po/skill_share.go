package po

import (
	"time"

	"gorm.io/gorm"
)

// SkillShare 技能共享表
// 用户将自己已安装的技能共享给团队其他用户
// SkillName 全局唯一：同一技能只存在一条共享记录，存在不可重新安装，Owner可覆盖升级，其他用户不可覆盖。
type SkillShare struct {
	ShareId      int       `gorm:"primaryKey;column:share_id;type:int;autoIncrement"`
	OwnerId      string    `gorm:"column:owner_id;type:varchar(64);index;not null"`         // 共享者 user_id
	SkillName    string    `gorm:"column:skill_name;type:varchar(64);uniqueIndex;not null"` // 技能名（全局唯一）
	Description  string    `gorm:"column:description;type:varchar(800)"`                    // 技能描述（共享时从 user_skill 快照）
	Icon         string    `gorm:"column:icon;type:varchar(256)"`                           // 图标（共享时从 user_skill 快照）
	CategoryId   int       `gorm:"column:category_id"`                                      // 分类ID（共享时从 user_skill 快照）
	Note         string    `gorm:"column:note;type:varchar(800)"`                           // 共享附言（区别于技能自身描述）
	InstallCount int       `gorm:"column:install_count;default:0"`                          // 累计被安装次数
	Created      time.Time `gorm:"not null;column:created"`
	Updated      time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (s *SkillShare) TableName() string {
	return "skill_share"
}

// BeforeCreate 创建时设置时间戳
func (s *SkillShare) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (s *SkillShare) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

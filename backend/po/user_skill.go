package po

import (
	"time"

	"gorm.io/gorm"
)

// UserSkillInstallStatus 用户技能安装状态常量
const (
	UserSkillInstallPending  = 0 // 未安装
	UserSkillInstallProgress = 1 // 安装中（SystemAgent 处理中）
	UserSkillInstalled       = 2 // 已安装
	UserSkillInstallFailed   = 3 // 安装失败
)

// UserSkill 用户技能安装表（记录用户安装了哪些市场技能）
type UserSkill struct {
	UserSkillId int    `gorm:"primaryKey;column:user_skill_id;type:int;autoIncrement"`                 // 主键ID
	UserId      string `gorm:"uniqueIndex:idx_user_skill;column:user_id;type:varchar(64);not null"`    // 用户ID
	SkillName   string `gorm:"uniqueIndex:idx_user_skill;column:skill_name;type:varchar(64);not null"` // 目录名（SkillFilter 用此字段匹配）
	Description string `gorm:"column:description;type:varchar(800)"`                                   // 简短描述
	Icon        string `gorm:"column:icon;type:varchar(256)"`                                          // 图标：Lucide 图标名称（如 "folder-git"、"globe"）或 URL

	MarketSkillId int    `gorm:"column:market_skill_id;index"`    // 关联 skill_market.skillId，custom 来源时为 0
	CategoryId    int    `gorm:"column:category_id;index"`        // 分类ID（安装时从 skill_market 快照）
	Source        string `gorm:"column:source;type:varchar(20)"`  // 来源：market / custom / share
	InstallStatus uint8  `gorm:"column:install_status;default:0"` // 安装状态：0未安装 1安装中 2已安装 3失败
	InstallError  string `gorm:"column:install_error;type:text"`  // 安装失败原因
	AlwaysOn      int    `gorm:"column:always_on;default:0"`      // 始终启用：0关闭 1开启

	Created time.Time `gorm:"not null;column:created"`
	Updated time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (us *UserSkill) TableName() string {
	return "user_skill"
}

// BeforeCreate 创建时设置时间戳
func (us *UserSkill) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	us.Created = now
	us.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (us *UserSkill) BeforeSave(tx *gorm.DB) error {
	us.Updated = time.Now()
	return nil
}

// IsInstalled 判断是否安装成功
func (us *UserSkill) IsInstalled() bool {
	return us.InstallStatus == UserSkillInstalled
}

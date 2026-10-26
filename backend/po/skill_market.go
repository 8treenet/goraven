package po

import (
	"time"

	"gorm.io/gorm"
)

// SkillSource 技能来源常量
const (
	SkillSourceClawHub      = "clawhub"       // 从 ClawHub 同步
	SkillSourceCustomUpload = "custom_upload" // 管理员手动上传 zip
	SkillSourceSystem       = "system"        // 系统内置
)

// SkillStatus 技能市场状态常量
const (
	SkillStatusDisabled = 0 // 禁用（下架）
	SkillStatusEnabled  = 1 // 启用（上架）
)

// SkillMarket 技能市场表（管理员维护的 skill 货架）
type SkillMarket struct {
	SkillId     int    `gorm:"primaryKey;column:skill_id;type:int;autoIncrement"`  // 主键ID
	Name        string `gorm:"uniqueIndex;column:name;type:varchar(64);not null"` // 唯一标识名，也是目录名
	Description string `gorm:"column:description;type:varchar(800)"`              // 简短描述
	Icon        string `gorm:"column:icon;type:varchar(256)"`                     // 图标：Lucide 图标名称（如 "folder-git"、"globe"）或 URL

	Source     string `gorm:"column:source;type:varchar(16)"`     // 来源：clawhub / custom_upload
	SourceUrl  string `gorm:"column:source_url;type:varchar(512)"` // 原始来源地址（clawhub 仓库地址或上传文件名）
	CategoryId int    `gorm:"column:category_id;index"`            // 分类ID

	Status         uint8     `gorm:"column:status;default:1"`         // 状态：0禁用 1启用
	SortOrder      int       `gorm:"column:sort_order;default:0"`      // 排序号
	InstalledCount int       `gorm:"column:installed_count;default:0"` // 被用户安装次数
	Remark         string    `gorm:"column:remark;type:varchar(512)"` // 管理员备注
	Deleted        uint8     `gorm:"column:deleted;default:0"`        // 软删除：0正常 1删除
	Created        time.Time `gorm:"not null;column:created"`
	Updated        time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (s *SkillMarket) TableName() string {
	return "skill_market"
}

// BeforeCreate 创建时设置时间戳
func (s *SkillMarket) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (s *SkillMarket) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

// IsEnabled 判断是否上架
func (s *SkillMarket) IsEnabled() bool {
	return s.Status == SkillStatusEnabled && s.Deleted == 0
}

package po

import (
	"time"

	"gorm.io/gorm"
)

// ToolDailyStats 工具/技能日统计表
// 按 (用户, 日期, 类型, 名称) 粒度预聚合工具调用次数
// 每轮对话 SSE end 事件后增量更新，支撑仪表盘技能/MCP 使用排行
type ToolDailyStats struct {
	StatId   int       `gorm:"primaryKey;column:id;autoIncrement"`
	UserId   string    `gorm:"uniqueIndex:idx_user_date_type_name;column:user_id;type:varchar(64);not null"`                         // 用户ID (空字符串表示全局聚合)
	ToolType string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:1;column:tool_type;type:varchar(16);not null"`  // 类型: mcp / skill / tool
	ToolName string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:3;column:tool_name;type:varchar(128);not null"` // mcp名称或工具名称或技能名称
	StatDate string    `gorm:"uniqueIndex:idx_user_date_type_name;index:idx_type_date_name,priority:2;column:stat_date;type:varchar(10);not null"` // 统计日期 (YYYY-MM-DD)
	Count    int       `gorm:"column:count;default:0"`                                                         // 当日调用次数
	Created  time.Time `gorm:"not null;column:created"`
	Updated  time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (s *ToolDailyStats) TableName() string {
	return "tool_daily_stats"
}

// BeforeCreate 创建时设置时间戳
func (s *ToolDailyStats) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.Created = now
	s.Updated = now
	return nil
}

// BeforeSave 更新时间戳
func (s *ToolDailyStats) BeforeSave(tx *gorm.DB) error {
	s.Updated = time.Now()
	return nil
}

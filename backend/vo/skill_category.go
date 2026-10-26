package vo

import "time"

// AdminSkillCategoryListReq 技能分类列表请求
type AdminSkillCategoryListReq struct {
	Search   string `url:"search"`   // 名称模糊搜索
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// AdminSkillCategoryItem 技能分类列表条目
type AdminSkillCategoryItem struct {
	CategoryId int       `json:"categoryId"` // 主键ID
	Name       string    `json:"name"`       // 分类名称
	Icon       string    `json:"icon"`       // 图标：Lucide 图标名称或 URL
	IsDefault  uint8     `json:"isDefault"`  // 是否默认分类
	SkillCount int       `json:"skillCount"` // 该分类下技能数量
	Updated    time.Time `json:"updated"`    // 更新时间
}

// AdminSkillCategoryDetailRsp 技能分类详情响应
type AdminSkillCategoryDetailRsp struct {
	CategoryId int       `json:"categoryId"` // 主键ID
	Name       string    `json:"name"`       // 分类名称
	Icon       string    `json:"icon"`       // 图标：Lucide 图标名称或 URL
	IsDefault  uint8     `json:"isDefault"`  // 是否默认分类
	Created    time.Time `json:"created"`    // 创建时间
	Updated    time.Time `json:"updated"`    // 更新时间
}

// AdminCreateSkillCategoryReq 创建技能分类请求
type AdminCreateSkillCategoryReq struct {
	Name string `json:"name" validate:"required"` // 分类名称
	Icon string `json:"icon"`                     // 图标：Lucide 图标名称或 URL（选填）
}

// AdminUpdateSkillCategoryReq 编辑技能分类请求
type AdminUpdateSkillCategoryReq struct {
	Name *string `json:"name"` // 分类名称
	Icon *string `json:"icon"` // 图标：Lucide 图标名称或 URL
}

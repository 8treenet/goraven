package vo

import "time"

// AdminSystemSkillListReq 管理员系统技能列表请求
type AdminSystemSkillListReq struct {
	Search   string `url:"search"`   // 名称/描述模糊搜索
	Status   *uint8 `url:"status"`  // 状态筛选：0启用 1禁用，nil不筛选
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// AdminSystemSkillItem 管理员系统技能列表条目
type AdminSystemSkillItem struct {
	SkillId     int       `json:"skillId"`     // 主键ID
	Name        string    `json:"name"`        // 唯一标识名，goraven-开头
	Description string    `json:"description"` // 简短描述
	Status      uint8     `json:"status"`      // 状态：0启用 1禁用
	Updated     time.Time `json:"updated"`     // 更新时间
}

// AdminSystemSkillDetailRsp 管理员系统技能详情响应（含完整 content，用于编辑回填）
type AdminSystemSkillDetailRsp struct {
	SkillId     int       `json:"skillId"`     // 主键ID
	Name        string    `json:"name"`        // 唯一标识名，goraven-开头
	Description string    `json:"description"` // 简短描述
	Content     string    `json:"content"`     // 完整 SKILL.md 内容
	Status      uint8     `json:"status"`      // 状态：0启用 1禁用
	Created     time.Time `json:"created"`     // 创建时间
	Updated     time.Time `json:"updated"`     // 更新时间
}

// AdminCreateSystemSkillReq 管理员创建系统技能请求
type AdminCreateSystemSkillReq struct {
	Content string `json:"content" validate:"required"` // 完整 SKILL.md 内容（YAML frontmatter + Markdown body）
}

// AdminUpdateSystemSkillReq 管理员编辑系统技能请求（仅传需要修改的字段）
type AdminUpdateSystemSkillReq struct {
	Content string `json:"content"` // 完整 SKILL.md 内容，为空时不修改
}

// AdminSystemSkillStatusReq 管理员系统技能状态切换请求
type AdminSystemSkillStatusReq struct {
	Status *uint8 `json:"status" validate:"required"` // 状态：0启用 1禁用
}

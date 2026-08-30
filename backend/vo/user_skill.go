package vo

import "time"

// UserSkillListReq 用户技能列表请求
type UserSkillListReq struct {
	Search     string `url:"search"`     // 名称模糊搜索
	CategoryId *int   `url:"categoryId"` // 分类筛选
	Source     string `url:"source"`     // 来源筛选：market / custom
	Status     *uint8 `url:"status"`     // 安装状态筛选
	Page       int    `url:"page"`       // 页码
	PageSize   int    `url:"pageSize"`   // 每页条数
}

// UserSkillItem 用户技能列表条目
type UserSkillItem struct {
	UserSkillId   int       `json:"userSkillId"`   // 用户技能ID
	SkillName     string    `json:"skillName"`     // 技能标识名
	Description   string    `json:"description"`   // 描述
	Icon          string    `json:"icon"`          // 图标
	MarketSkillId int       `json:"marketSkillId"` // 关联市场技能ID
	CategoryId    int       `json:"categoryId"`    // 分类ID
	CategoryName  string    `json:"categoryName"`  // 分类名
	Source        string    `json:"source"`        // 来源：market / custom
	InstallStatus uint8     `json:"installStatus"` // 安装状态：0未安装 1安装中 2已安装 3失败
	InstallError  string    `json:"installError"`  // 安装失败原因
	AlwaysOn      int       `json:"alwaysOn"`      // 始终启用：0关闭 1开启
	Created       time.Time `json:"created"`       // 创建时间
	Updated       time.Time `json:"updated"`       // 更新时间
}

// UserSkillDetailRsp 用户技能详情响应
type UserSkillDetailRsp struct {
	UserSkillId   int       `json:"userSkillId"`   // 用户技能ID
	SkillName     string    `json:"skillName"`     // 技能标识名
	Description   string    `json:"description"`   // 描述
	Icon          string    `json:"icon"`          // 图标
	MarketSkillId int       `json:"marketSkillId"` // 关联市场技能ID
	CategoryId    int       `json:"categoryId"`    // 分类ID
	CategoryName  string    `json:"categoryName"`  // 分类名
	Source        string    `json:"source"`        // 来源：market / custom
	InstallStatus uint8     `json:"installStatus"` // 安装状态
	InstallError  string    `json:"installError"`  // 安装失败原因
	Content       string    `json:"content"`       // SKILL.md 原文内容
	IsShared      bool      `json:"isShared"`      // 是否已共享到团队
	Created       time.Time `json:"created"`       // 创建时间
	Updated       time.Time `json:"updated"`       // 更新时间
}

// UserSkillUpdateReq 编辑用户技能请求
type UserSkillUpdateReq struct {
	Icon       *string `json:"icon"`       // 图标：Lucide 图标名称或 URL
	CategoryId *int    `json:"categoryId"` // 分类ID
}

// UserSkillToggleAlwaysOnReq 切换始终启用请求
type UserSkillToggleAlwaysOnReq struct {
	AlwaysOn int `json:"alwaysOn" validate:"oneof=0 1"` // 始终启用：0关闭 1开启
}

// UserSkillInstallReq 安装技能请求
type UserSkillInstallReq struct {
	SkillId int `json:"skillId" validate:"required"` // skill_market.skillId
}

// UserSkillInstallRsp 安装技能响应
type UserSkillInstallRsp struct {
	UserSkillId int `json:"userSkillId"` // 创建的用户技能ID
}

// UserSkillRefreshRsp 刷新用户技能响应
type UserSkillRefreshRsp struct {
	Added   int `json:"added"`   // 新增技能数
	Removed int `json:"removed"` // 移除孤立技能数
}

// UserSkillStatusRsp 用户技能当前状态响应（用于安装进度轮询）
type UserSkillStatusRsp struct {
	UserSkillId   int    `json:"userSkillId"`   // 用户技能ID
	InstallStatus uint8  `json:"installStatus"` // 安装状态：0未安装 1安装中 2已安装 3失败
	InstallError  string `json:"installError"`  // 安装失败原因
}

// UserMarketSkillListReq 用户市场技能列表请求
type UserMarketSkillListReq struct {
	Search     string `url:"search"`     // 名称模糊搜索
	CategoryId *int   `url:"categoryId"` // 分类筛选
	Source     string `url:"source"`     // 来源筛选：clawhub / custom_upload
	Page       int    `url:"page"`       // 页码
	PageSize   int    `url:"pageSize"`   // 每页条数
}

// UserMarketSkillItem 用户市场技能列表条目
type UserMarketSkillItem struct {
	SkillId        int       `json:"skillId"`        // 市场技能ID
	Name           string    `json:"name"`           // 唯一标识名
	Description    string    `json:"description"`    // 描述
	Icon           string    `json:"icon"`           // 图标
	Source         string    `json:"source"`         // 来源
	CategoryId     int       `json:"categoryId"`     // 分类ID
	CategoryName   string    `json:"categoryName"`   // 分类名
	InstalledCount int       `json:"installedCount"` // 被安装次数
	UserInstalled  bool      `json:"userInstalled"`  // 当前用户是否已安装
	Updated        time.Time `json:"updated"`        // 更新时间
}

// SkillCategoryItem 技能分类条目（用户视角）
type SkillCategoryItem struct {
	CategoryId int    `json:"categoryId"` // 分类ID
	Name       string `json:"name"`       // 分类名称
	Icon       string `json:"icon"`       // 分类图标
}

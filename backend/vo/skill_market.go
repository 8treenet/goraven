package vo

import "time"

// AdminMarketSkillListReq 管理员市场技能列表请求
type AdminMarketSkillListReq struct {
	Search   string `url:"search"`   // 名称模糊搜索
	Source   string `url:"source"`   // 来源筛选：clawhub/custom_upload，空不筛选
	Status   *uint8 `url:"status"`   // 状态筛选：0下架 1上架，nil不筛选
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// AdminMarketSkillItem 管理员市场技能列表条目
type AdminMarketSkillItem struct {
	SkillId        int       `json:"skillId"`        // 主键ID
	Name           string    `json:"name"`           // 唯一标识名（目录名）
	Description    string    `json:"description"`    // 简短描述
	Icon           string    `json:"icon"`           // 图标：Lucide 图标名称或 URL
	Source         string    `json:"source"`         // 来源：clawhub / custom_upload
	CategoryId     int       `json:"categoryId"`     // 分类ID
	CategoryName   string    `json:"categoryName"`   // 分类名称
	CategoryIcon   string    `json:"categoryIcon"`   // 分类图标
	InstalledCount int       `json:"installedCount"` // 被用户安装次数
	Status         uint8     `json:"status"`         // 状态：0下架 1上架
	SortOrder      int       `json:"sortOrder"`      // 排序号
	Updated        time.Time `json:"updated"`        // 更新时间
}

// AdminMarketSkillDetailRsp 管理员市场技能详情响应
type AdminMarketSkillDetailRsp struct {
	SkillId        int       `json:"skillId"`        // 主键ID
	Name           string    `json:"name"`           // 唯一标识名（目录名）
	Description    string    `json:"description"`    // 简短描述
	Icon           string    `json:"icon"`           // 图标：Lucide 图标名称或 URL
	Source         string    `json:"source"`         // 来源：clawhub / custom_upload
	SourceUrl      string    `json:"sourceUrl"`      // 原始来源地址
	CategoryId     int       `json:"categoryId"`     // 分类ID
	CategoryName   string    `json:"categoryName"`   // 分类名称
	CategoryIcon   string    `json:"categoryIcon"`   // 分类图标
	Status         uint8     `json:"status"`         // 状态：0下架 1上架
	SortOrder      int       `json:"sortOrder"`      // 排序号
	InstalledCount int       `json:"installedCount"` // 被用户安装次数
	Remark         string    `json:"remark"`         // 管理员备注
	Content        string    `json:"content"`        // SKILL.md 完整内容
	Created        time.Time `json:"created"`        // 创建时间
	Updated        time.Time `json:"updated"`        // 更新时间
}

// AdminUpdateMarketSkillReq 管理员编辑市场技能请求（仅传需要修改的字段）
type AdminUpdateMarketSkillReq struct {
	Icon        *string `json:"icon"`        // 图标：Lucide 图标名称或 URL
	CategoryId  *int    `json:"categoryId"`  // 分类ID
	SortOrder   *int    `json:"sortOrder"`   // 排序号
	Remark      *string `json:"remark"`      // 管理员备注
}

// AdminMarketSkillStatusReq 管理员市场技能状态切换请求
type AdminMarketSkillStatusReq struct {
	Status *uint8 `json:"status" validate:"required"` // 状态：0下架 1上架
}

// AdminPublishMarketSkillReq 管理员发布市场技能请求（从已上传 zip）
type AdminPublishMarketSkillReq struct {
	UploadId   string `json:"uploadId" validate:"required"`   // 已合并完成的 HFS 上传任务 ID
	Icon       string `json:"icon"`                           // 图标：Lucide 图标名称或 URL（选填）
	CategoryId int    `json:"categoryId" validate:"required"` // 分类ID（必填）
}

// AdminImportClawHubSkillReq 管理员从 ClawHub 导入技能请求
type AdminImportClawHubSkillReq struct {
	Slug       string `json:"slug" validate:"required"`       // ClawHub 技能唯一标识
	Icon       string `json:"icon"`                           // 图标：Lucide 图标名称或 URL（选填）
	CategoryId int    `json:"categoryId" validate:"required"` // 分类ID（必填）
}

// AdminMarketSkillUserItem 安装了该市场技能的用户条目
type AdminMarketSkillUserItem struct {
	UserId        string    `json:"userId"`        // 用户ID
	InstallStatus uint8     `json:"installStatus"` // 安装状态：0未安装 1安装中 2已安装 3失败
	Created       time.Time `json:"created"`       // 安装时间
}

// AdminDeleteMarketSkillReq 管理员删除市场技能请求
type AdminDeleteMarketSkillReq struct {
	Cascade bool `url:"cascade"` // 是否级联删除用户已安装的该技能记录
}

// AdminMarketSkillUserListReq 管理员市场技能用户列表请求
type AdminMarketSkillUserListReq struct {
	Page     int `url:"page"`     // 页码
	PageSize int `url:"pageSize"` // 每页条数
}

// UserAvailableSkillItem 用户可选技能列表条目（仅用户已安装技能，系统技能全局默认使用不在此展示）
type UserAvailableSkillItem struct {
	UserSkillId  int    `json:"userSkillId"`  // 用户技能ID
	SkillName    string `json:"skillName"`    // 技能标识名（目录名）
	Description  string `json:"description"`  // 简短描述
	Icon         string `json:"icon"`         // 图标
	Source       string `json:"source"`       // 来源：market / custom
	CategoryId   int    `json:"categoryId"`   // 分类ID
	CategoryName string `json:"categoryName"` // 分类名称
}

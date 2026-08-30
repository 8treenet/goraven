package vo

import "time"

// ════════════════════════════════════════════════════════════════════════════
// 角色模板
// ════════════════════════════════════════════════════════════════════════════

// AdminPersonaTemplateListReq 角色模板列表请求
type AdminPersonaTemplateListReq struct {
	Search     string `url:"search"`     // 名称模糊搜索
	CategoryId *int   `url:"categoryId"` // 分类筛选，nil 不筛选
	Page       int    `url:"page"`       // 页码
	PageSize   int    `url:"pageSize"`   // 每页条数
}

// AdminPersonaTemplateItem 角色模板列表条目
type AdminPersonaTemplateItem struct {
	TemplateId  int       `json:"templateId"`  // 主键ID
	Name        string    `json:"name"`        // 模板名称（展示名称）
	Icon        string    `json:"icon"`        // 图标：Lucide 图标名称或 URL
	Description string    `json:"description"` // 模板描述
	RoleInfo    string    `json:"roleInfo"`    // 系统提示词摘要（截断预览）
	CategoryId  int       `json:"categoryId"`  // 分类ID
	UsageCount  int       `json:"usageCount"`  // 使用次数
	SortOrder   int       `json:"sortOrder"`   // 排序号
	Updated     time.Time `json:"updated"`     // 更新时间

	// 关联分类信息（service 层组装）
	CategoryName string `json:"categoryName,omitempty"` // 分类名称
	CategoryIcon string `json:"categoryIcon,omitempty"` // 分类图标
}

// AdminPersonaTemplateDetailRsp 角色模板详情响应（含完整 roleInfo，用于编辑回填）
type AdminPersonaTemplateDetailRsp struct {
	TemplateId         int       `json:"templateId"`         // 主键ID
	Name               string    `json:"name"`               // 模板名称（展示名称）
	Icon               string    `json:"icon"`               // 图标：Lucide 图标名称或 URL
	Description        string    `json:"description"`        // 模板描述
	RoleInfo           string    `json:"roleInfo"`           // 系统提示词（完整内容）
	CategoryId         int       `json:"categoryId"`         // 分类ID
	UsageCount         int       `json:"usageCount"`         // 使用次数
	SortOrder          int       `json:"sortOrder"`          // 排序号
	Created            time.Time `json:"created"`            // 创建时间
	Updated            time.Time `json:"updated"`            // 更新时间

	// 关联分类信息（service 层组装）
	CategoryName string `json:"categoryName,omitempty"` // 分类名称
	CategoryIcon string `json:"categoryIcon,omitempty"` // 分类图标
}

// AdminCreatePersonaTemplateReq 创建角色模板请求
type AdminCreatePersonaTemplateReq struct {
	Name               string `json:"name" validate:"required"`         // 模板名称（展示名称）
	Icon               string `json:"icon"`                              // 图标：Lucide 图标名称或 URL
	Description        string `json:"description"`                       // 模板描述（选填）
	RoleInfo           string `json:"roleInfo" validate:"required"`      // 系统提示词（核心内容）
	CategoryId         int    `json:"categoryId" validate:"required"`    // 分类ID
	SortOrder          int    `json:"sortOrder"`                         // 排序号
}

// AdminUpdatePersonaTemplateReq 编辑角色模板请求（仅传需要修改的字段）
type AdminUpdatePersonaTemplateReq struct {
	Name               *string `json:"name"`                // 模板名称
	Icon               *string `json:"icon"`                // 图标
	Description        *string `json:"description"`         // 模板描述
	RoleInfo           *string `json:"roleInfo"`            // 系统提示词
	CategoryId         *int    `json:"categoryId"`          // 分类ID
	SortOrder          *int    `json:"sortOrder"`           // 排序号
}

// ════════════════════════════════════════════════════════════════════════════
// 角色分类
// ════════════════════════════════════════════════════════════════════════════

// AdminPersonaCategoryListReq 角色分类列表请求
type AdminPersonaCategoryListReq struct {
	Search   string `url:"search"`   // 名称模糊搜索
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// AdminPersonaCategoryItem 角色分类列表条目
type AdminPersonaCategoryItem struct {
	CategoryId    int       `json:"categoryId"`    // 主键ID
	Name          string    `json:"name"`          // 分类名称
	Icon          string    `json:"icon"`          // 图标：Lucide 图标名称或 URL
	IsDefault     uint8     `json:"isDefault"`     // 是否默认分类
	TemplateCount int       `json:"templateCount"` // 该分类下模板数量
	Updated       time.Time `json:"updated"`       // 更新时间
}

// AdminPersonaCategoryDetailRsp 角色分类详情响应
type AdminPersonaCategoryDetailRsp struct {
	CategoryId int       `json:"categoryId"` // 主键ID
	Name       string    `json:"name"`       // 分类名称
	Icon       string    `json:"icon"`       // 图标：Lucide 图标名称或 URL
	IsDefault  uint8     `json:"isDefault"`  // 是否默认分类
	Created    time.Time `json:"created"`    // 创建时间
	Updated    time.Time `json:"updated"`    // 更新时间
}

// AdminCreatePersonaCategoryReq 创建角色分类请求
type AdminCreatePersonaCategoryReq struct {
	Name string `json:"name" validate:"required"` // 分类名称
	Icon string `json:"icon"`                     // 图标：Lucide 图标名称或 URL（选填）
}

// AdminUpdatePersonaCategoryReq 编辑角色分类请求
type AdminUpdatePersonaCategoryReq struct {
	Name *string `json:"name"` // 分类名称
	Icon *string `json:"icon"` // 图标：Lucide 图标名称或 URL
}

// ════════════════════════════════════════════════════════════════════════════
// 用户角色
// ════════════════════════════════════════════════════════════════════════════

// UserPersonaSidebarItem 侧边栏角色列表条目（轻量，仅 id/name/icon）
type UserPersonaSidebarItem struct {
	PersonaId int    `json:"personaId"` // 主键ID
	Name      string `json:"name"`      // 角色名称
	Icon      string `json:"icon"`      // 图标
}

// UserPersonaSimpleItem 角色简要信息（聊天页选择器用，仅含基础字段）
type UserPersonaSimpleItem struct {
	PersonaId int    `json:"personaId"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
}

// UserPersonaListItem 角色列表条目（详情页列表用，含分类、模型、MCP、技能信息）
type UserPersonaListItem struct {
	PersonaId    int      `json:"personaId"`
	Name         string   `json:"name"`
	Icon         string   `json:"icon"`
	RoleInfo     string   `json:"roleInfo"`
	CategoryId   int      `json:"categoryId"`
	CategoryName string   `json:"categoryName,omitempty"`
	CategoryIcon string   `json:"categoryIcon,omitempty"`
	ModelName    string   `json:"modelName,omitempty"`
	McpIds       []int    `json:"mcpIds"`
	SkillIds     []int    `json:"skillIds"`
	McpNames     []string `json:"mcpNames"`
	SkillNames   []string `json:"skillNames"`
}

// UserPersonaDetailRsp 用户角色详情响应（编辑页回填，含完整配置）
type UserPersonaDetailRsp struct {
	PersonaId   int       `json:"personaId"`   // 主键ID
	Name        string    `json:"name"`        // 角色名称
	Icon        string    `json:"icon"`        // 图标
	RoleInfo    string    `json:"roleInfo"`    // 角色设定文案
	CategoryId  int       `json:"categoryId"`  // 分类ID
	McpIds      []int     `json:"mcpIds"`      // MCP配置ID列表
	SkillIds    []int     `json:"skillIds"`    // 技能ID列表
	AIModelId   int       `json:"aiModelId"`   // 关联模型ID，0表示使用默认模型
	Created     time.Time `json:"created"`     // 创建时间
	Updated     time.Time `json:"updated"`     // 更新时间

	// 关联信息（service 层组装）
	CategoryName string `json:"categoryName,omitempty"` // 分类名称
	CategoryIcon string `json:"categoryIcon,omitempty"` // 分类图标
	ModelName    string `json:"modelName,omitempty"`    // 模型名称（providerDisplayName - modelName）
	ModelIcon    string `json:"modelIcon,omitempty"`    // 模型图标URL
}

// CreateUserPersonaReq 创建用户角色请求
type CreateUserPersonaReq struct {
	Name       string `json:"name" validate:"required"`     // 角色名称（同一用户下唯一，2-50字）
	Icon       string `json:"icon"`                         // 图标：Lucide 图标名称或 URL
	RoleInfo   string `json:"roleInfo" validate:"required"` // 角色设定文案
	CategoryId int    `json:"categoryId" validate:"required"` // 分类ID
	McpIds     []int  `json:"mcpIds"`                       // MCP配置ID列表
	SkillIds   []int  `json:"skillIds"`                     // 技能ID列表
	AIModelId  int    `json:"aiModelId"`                    // 关联模型ID，0表示使用默认模型
	TemplateId *int   `json:"templateId"`                   // 关联模板ID（可选，选择模板创建时传入，后端递增 usageCount）
}

// UpdateUserPersonaReq 编辑用户角色请求（仅传需要修改的字段）
type UpdateUserPersonaReq struct {
	Name       *string `json:"name"`       // 角色名称
	Icon       *string `json:"icon"`       // 图标
	RoleInfo   *string `json:"roleInfo"`   // 角色设定文案
	CategoryId *int    `json:"categoryId"` // 分类ID
	McpIds     *[]int  `json:"mcpIds"`     // MCP配置ID列表
	SkillIds   *[]int  `json:"skillIds"`   // 技能ID列表
	AIModelId  *int    `json:"aiModelId"`  // 关联模型ID
}

// UserPersonaTemplateItem 用户可选角色模板条目
type UserPersonaTemplateItem struct {
	TemplateId  int    `json:"templateId"`  // 主键ID
	Name        string `json:"name"`        // 模板名称
	Icon        string `json:"icon"`        // 图标
	Description string `json:"description"` // 模板描述
	CategoryId  int    `json:"categoryId"`  // 分类ID

	// 关联分类信息（service 层组装）
	CategoryName string `json:"categoryName,omitempty"` // 分类名称
	CategoryIcon string `json:"categoryIcon,omitempty"` // 分类图标
}

// UserPersonaTemplateDetailRsp 用户角色模板详情响应（含完整 roleInfo，用于预填）
type UserPersonaTemplateDetailRsp struct {
	TemplateId  int    `json:"templateId"`  // 主键ID
	Name        string `json:"name"`        // 模板名称
	Icon        string `json:"icon"`        // 图标
	Description string `json:"description"` // 模板描述
	RoleInfo    string `json:"roleInfo"`    // 系统提示词（完整内容，用于预填角色设定）
	CategoryId  int    `json:"categoryId"`  // 分类ID

	// 关联分类信息（service 层组装）
	CategoryName string `json:"categoryName,omitempty"` // 分类名称
	CategoryIcon string `json:"categoryIcon,omitempty"` // 分类图标
}

// UserPersonaCategoryItem 用户可选角色分类条目
type UserPersonaCategoryItem struct {
	CategoryId int    `json:"categoryId"` // 主键ID
	Name       string `json:"name"`       // 分类名称
	Icon       string `json:"icon"`       // 图标
}

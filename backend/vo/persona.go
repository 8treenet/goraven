package vo

import "time"

type AdminPersonaTemplateListReq struct {
	Search		string	`url:"search"`
	CategoryId	*int	`url:"categoryId"`
	Page		int	`url:"page"`
	PageSize	int	`url:"pageSize"`
}

type AdminPersonaTemplateItem struct {
	TemplateId	int		`json:"templateId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	Description	string		`json:"description"`
	RoleInfo	string		`json:"roleInfo"`
	CategoryId	int		`json:"categoryId"`
	UsageCount	int		`json:"usageCount"`
	SortOrder	int		`json:"sortOrder"`
	Updated		time.Time	`json:"updated"`

	CategoryName	string	`json:"categoryName,omitempty"`
	CategoryIcon	string	`json:"categoryIcon,omitempty"`
}

type AdminPersonaTemplateDetailRsp struct {
	TemplateId	int		`json:"templateId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	Description	string		`json:"description"`
	RoleInfo	string		`json:"roleInfo"`
	CategoryId	int		`json:"categoryId"`
	UsageCount	int		`json:"usageCount"`
	SortOrder	int		`json:"sortOrder"`
	Created		time.Time	`json:"created"`
	Updated		time.Time	`json:"updated"`

	CategoryName	string	`json:"categoryName,omitempty"`
	CategoryIcon	string	`json:"categoryIcon,omitempty"`
}

type AdminCreatePersonaTemplateReq struct {
	Name		string	`json:"name" validate:"required"`
	Icon		string	`json:"icon"`
	Description	string	`json:"description"`
	RoleInfo	string	`json:"roleInfo" validate:"required"`
	CategoryId	int	`json:"categoryId" validate:"required"`
	SortOrder	int	`json:"sortOrder"`
}

type AdminUpdatePersonaTemplateReq struct {
	Name		*string	`json:"name"`
	Icon		*string	`json:"icon"`
	Description	*string	`json:"description"`
	RoleInfo	*string	`json:"roleInfo"`
	CategoryId	*int	`json:"categoryId"`
	SortOrder	*int	`json:"sortOrder"`
}

type AdminPersonaCategoryListReq struct {
	Search		string	`url:"search"`
	Page		int	`url:"page"`
	PageSize	int	`url:"pageSize"`
}

type AdminPersonaCategoryItem struct {
	CategoryId	int		`json:"categoryId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	IsDefault	uint8		`json:"isDefault"`
	TemplateCount	int		`json:"templateCount"`
	Updated		time.Time	`json:"updated"`
}

type AdminPersonaCategoryDetailRsp struct {
	CategoryId	int		`json:"categoryId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	IsDefault	uint8		`json:"isDefault"`
	Created		time.Time	`json:"created"`
	Updated		time.Time	`json:"updated"`
}

type AdminCreatePersonaCategoryReq struct {
	Name	string	`json:"name" validate:"required"`
	Icon	string	`json:"icon"`
}

type AdminUpdatePersonaCategoryReq struct {
	Name	*string	`json:"name"`
	Icon	*string	`json:"icon"`
}

type UserPersonaSidebarItem struct {
	PersonaId	int	`json:"personaId"`
	Name		string	`json:"name"`
	Icon		string	`json:"icon"`
}

type UserPersonaSimpleItem struct {
	PersonaId	int	`json:"personaId"`
	Name		string	`json:"name"`
	Icon		string	`json:"icon"`
}

type UserPersonaListItem struct {
	PersonaId	int		`json:"personaId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	RoleInfo	string		`json:"roleInfo"`
	CategoryId	int		`json:"categoryId"`
	CategoryName	string		`json:"categoryName,omitempty"`
	CategoryIcon	string		`json:"categoryIcon,omitempty"`
	ModelName	string		`json:"modelName,omitempty"`
	McpIds		[]int		`json:"mcpIds"`
	SkillIds	[]int		`json:"skillIds"`
	McpNames	[]string	`json:"mcpNames"`
	SkillNames	[]string	`json:"skillNames"`
}

type UserPersonaDetailRsp struct {
	PersonaId	int		`json:"personaId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	RoleInfo	string		`json:"roleInfo"`
	CategoryId	int		`json:"categoryId"`
	McpIds		[]int		`json:"mcpIds"`
	SkillIds	[]int		`json:"skillIds"`
	AIModelId	int		`json:"aiModelId"`
	Created		time.Time	`json:"created"`
	Updated		time.Time	`json:"updated"`

	CategoryName	string	`json:"categoryName,omitempty"`
	CategoryIcon	string	`json:"categoryIcon,omitempty"`
	ModelName	string	`json:"modelName,omitempty"`
}

type CreateUserPersonaReq struct {
	Name		string	`json:"name" validate:"required"`
	Icon		string	`json:"icon"`
	RoleInfo	string	`json:"roleInfo" validate:"required"`
	CategoryId	int	`json:"categoryId" validate:"required"`
	McpIds		[]int	`json:"mcpIds"`
	SkillIds	[]int	`json:"skillIds"`
	AIModelId	int	`json:"aiModelId"`
	TemplateId	*int	`json:"templateId"`
}

type UpdateUserPersonaReq struct {
	Name		*string	`json:"name"`
	Icon		*string	`json:"icon"`
	RoleInfo	*string	`json:"roleInfo"`
	CategoryId	*int	`json:"categoryId"`
	McpIds		*[]int	`json:"mcpIds"`
	SkillIds	*[]int	`json:"skillIds"`
	AIModelId	*int	`json:"aiModelId"`
}

type UserPersonaTemplateItem struct {
	TemplateId	int	`json:"templateId"`
	Name		string	`json:"name"`
	Icon		string	`json:"icon"`
	Description	string	`json:"description"`
	CategoryId	int	`json:"categoryId"`

	CategoryName	string	`json:"categoryName,omitempty"`
	CategoryIcon	string	`json:"categoryIcon,omitempty"`
}

type UserPersonaTemplateDetailRsp struct {
	TemplateId	int	`json:"templateId"`
	Name		string	`json:"name"`
	Icon		string	`json:"icon"`
	Description	string	`json:"description"`
	RoleInfo	string	`json:"roleInfo"`
	CategoryId	int	`json:"categoryId"`

	CategoryName	string	`json:"categoryName,omitempty"`
	CategoryIcon	string	`json:"categoryIcon,omitempty"`
}

type UserPersonaCategoryItem struct {
	CategoryId	int	`json:"categoryId"`
	Name		string	`json:"name"`
	Icon		string	`json:"icon"`
}

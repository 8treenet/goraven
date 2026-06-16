package vo

import "time"

type AdminSkillCategoryListReq struct {
	Search		string	`url:"search"`
	Page		int	`url:"page"`
	PageSize	int	`url:"pageSize"`
}

type AdminSkillCategoryItem struct {
	CategoryId	int		`json:"categoryId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	IsDefault	uint8		`json:"isDefault"`
	SkillCount	int		`json:"skillCount"`
	Updated		time.Time	`json:"updated"`
}

type AdminSkillCategoryDetailRsp struct {
	CategoryId	int		`json:"categoryId"`
	Name		string		`json:"name"`
	Icon		string		`json:"icon"`
	IsDefault	uint8		`json:"isDefault"`
	Created		time.Time	`json:"created"`
	Updated		time.Time	`json:"updated"`
}

type AdminCreateSkillCategoryReq struct {
	Name	string	`json:"name" validate:"required"`
	Icon	string	`json:"icon"`
}

type AdminUpdateSkillCategoryReq struct {
	Name	*string	`json:"name"`
	Icon	*string	`json:"icon"`
}

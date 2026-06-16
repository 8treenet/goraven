package vo

import "time"

type AdminSystemSkillListReq struct {
	Search		string	`url:"search"`
	Status		*uint8	`url:"status"`
	Page		int	`url:"page"`
	PageSize	int	`url:"pageSize"`
}

type AdminSystemSkillItem struct {
	SkillId		int		`json:"skillId"`
	Name		string		`json:"name"`
	Description	string		`json:"description"`
	Status		uint8		`json:"status"`
	Updated		time.Time	`json:"updated"`
}

type AdminSystemSkillDetailRsp struct {
	SkillId		int		`json:"skillId"`
	Name		string		`json:"name"`
	Description	string		`json:"description"`
	Content		string		`json:"content"`
	Status		uint8		`json:"status"`
	Created		time.Time	`json:"created"`
	Updated		time.Time	`json:"updated"`
}

type AdminCreateSystemSkillReq struct {
	Content string `json:"content" validate:"required"`
}

type AdminUpdateSystemSkillReq struct {
	Content string `json:"content"`
}

type AdminSystemSkillStatusReq struct {
	Status *uint8 `json:"status" validate:"required"`
}

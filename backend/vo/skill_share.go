package vo

// SkillShareListReq 团队技能列表请求
type SkillShareListReq struct {
	Page     int    `url:"page"`
	PageSize int    `url:"pageSize"`
	Search   string `url:"search"`
}

// SkillShareItem 团队技能列表项
type SkillShareItem struct {
	ShareId      int    `json:"shareId"`
	OwnerId      string `json:"ownerId"`
	OwnerName    string `json:"ownerName"`
	SkillName    string `json:"skillName"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	CategoryId   int    `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Note         string `json:"note"`
	InstallCount int    `json:"installCount"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	CanDelete    bool   `json:"canDelete"`
}

// ShareSkillReq 共享技能请求
type ShareSkillReq struct {
	UserSkillId int    `json:"userSkillId" validate:"required"`
	Note        string `json:"note"`
}

// ShareSkillRsp 共享技能响应
type ShareSkillRsp struct {
	ShareId int `json:"shareId"`
}

// SkillShareDetailRsp 团队技能详情（含 SKILL.md 内容）
type SkillShareDetailRsp struct {
	ShareId      int    `json:"shareId"`
	OwnerId      string `json:"ownerId"`
	OwnerName    string `json:"ownerName"`
	SkillName    string `json:"skillName"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	CategoryId   int    `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Note         string `json:"note"`
	Content      string `json:"content"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

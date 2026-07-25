package vo

type SkillShareListReq struct {
	Page     int    `url:"page"`
	PageSize int    `url:"pageSize"`
	Search   string `url:"search"`
}

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

type ShareSkillReq struct {
	UserSkillId int    `json:"userSkillId" validate:"required"`
	Note        string `json:"note"`
}

type ShareSkillRsp struct {
	ShareId int `json:"shareId"`
}

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

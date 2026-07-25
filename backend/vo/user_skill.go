package vo

import "time"

type UserSkillListReq struct {
	Search     string `url:"search"`
	CategoryId *int   `url:"categoryId"`
	Source     string `url:"source"`
	Status     *uint8 `url:"status"`
	Page       int    `url:"page"`
	PageSize   int    `url:"pageSize"`
}

type UserSkillItem struct {
	UserSkillId   int       `json:"userSkillId"`
	SkillName     string    `json:"skillName"`
	Description   string    `json:"description"`
	Icon          string    `json:"icon"`
	MarketSkillId int       `json:"marketSkillId"`
	CategoryId    int       `json:"categoryId"`
	CategoryName  string    `json:"categoryName"`
	Source        string    `json:"source"`
	InstallStatus uint8     `json:"installStatus"`
	InstallError  string    `json:"installError"`
	AlwaysOn      int       `json:"alwaysOn"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

type UserSkillDetailRsp struct {
	UserSkillId   int       `json:"userSkillId"`
	SkillName     string    `json:"skillName"`
	Description   string    `json:"description"`
	Icon          string    `json:"icon"`
	MarketSkillId int       `json:"marketSkillId"`
	CategoryId    int       `json:"categoryId"`
	CategoryName  string    `json:"categoryName"`
	Source        string    `json:"source"`
	InstallStatus uint8     `json:"installStatus"`
	InstallError  string    `json:"installError"`
	Content       string    `json:"content"`
	IsShared      bool      `json:"isShared"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

type UserSkillUpdateReq struct {
	Icon       *string `json:"icon"`
	CategoryId *int    `json:"categoryId"`
}

type UserSkillToggleAlwaysOnReq struct {
	AlwaysOn int `json:"alwaysOn" validate:"oneof=0 1"`
}

type UserSkillInstallReq struct {
	SkillId int `json:"skillId" validate:"required"`
}

type UserSkillInstallRsp struct {
	UserSkillId int `json:"userSkillId"`
}

type UserSkillRefreshRsp struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
}

type UserSkillStatusRsp struct {
	UserSkillId   int    `json:"userSkillId"`
	InstallStatus uint8  `json:"installStatus"`
	InstallError  string `json:"installError"`
}

type UserMarketSkillListReq struct {
	Search     string `url:"search"`
	CategoryId *int   `url:"categoryId"`
	Source     string `url:"source"`
	Page       int    `url:"page"`
	PageSize   int    `url:"pageSize"`
}

type UserMarketSkillItem struct {
	SkillId        int       `json:"skillId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Icon           string    `json:"icon"`
	Source         string    `json:"source"`
	CategoryId     int       `json:"categoryId"`
	CategoryName   string    `json:"categoryName"`
	InstalledCount int       `json:"installedCount"`
	UserInstalled  bool      `json:"userInstalled"`
	Updated        time.Time `json:"updated"`
}

type SkillCategoryItem struct {
	CategoryId int    `json:"categoryId"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
}

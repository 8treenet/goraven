package vo

import "time"

type AdminMarketSkillListReq struct {
	Search   string `url:"search"`
	Source   string `url:"source"`
	Status   *uint8 `url:"status"`
	Page     int    `url:"page"`
	PageSize int    `url:"pageSize"`
}

type AdminMarketSkillItem struct {
	SkillId        int       `json:"skillId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Icon           string    `json:"icon"`
	Source         string    `json:"source"`
	CategoryId     int       `json:"categoryId"`
	CategoryName   string    `json:"categoryName"`
	CategoryIcon   string    `json:"categoryIcon"`
	InstalledCount int       `json:"installedCount"`
	Status         uint8     `json:"status"`
	SortOrder      int       `json:"sortOrder"`
	Updated        time.Time `json:"updated"`
}

type AdminMarketSkillDetailRsp struct {
	SkillId        int       `json:"skillId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Icon           string    `json:"icon"`
	Source         string    `json:"source"`
	SourceUrl      string    `json:"sourceUrl"`
	CategoryId     int       `json:"categoryId"`
	CategoryName   string    `json:"categoryName"`
	CategoryIcon   string    `json:"categoryIcon"`
	Status         uint8     `json:"status"`
	SortOrder      int       `json:"sortOrder"`
	InstalledCount int       `json:"installedCount"`
	Remark         string    `json:"remark"`
	Content        string    `json:"content"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}

type AdminUpdateMarketSkillReq struct {
	Icon       *string `json:"icon"`
	CategoryId *int    `json:"categoryId"`
	SortOrder  *int    `json:"sortOrder"`
	Remark     *string `json:"remark"`
}

type AdminMarketSkillStatusReq struct {
	Status *uint8 `json:"status" validate:"required"`
}

type AdminPublishMarketSkillReq struct {
	UploadId   string `json:"uploadId" validate:"required"`
	Icon       string `json:"icon"`
	CategoryId int    `json:"categoryId" validate:"required"`
}

type AdminImportClawHubSkillReq struct {
	Slug       string `json:"slug" validate:"required"`
	Icon       string `json:"icon"`
	CategoryId int    `json:"categoryId" validate:"required"`
}

type AdminMarketSkillUserItem struct {
	UserId        string    `json:"userId"`
	InstallStatus uint8     `json:"installStatus"`
	Created       time.Time `json:"created"`
}

type AdminDeleteMarketSkillReq struct {
	Cascade bool `url:"cascade"`
}

type AdminMarketSkillUserListReq struct {
	Page     int `url:"page"`
	PageSize int `url:"pageSize"`
}

type UserAvailableSkillItem struct {
	UserSkillId  int    `json:"userSkillId"`
	SkillName    string `json:"skillName"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Source       string `json:"source"`
	CategoryId   int    `json:"categoryId"`
	CategoryName string `json:"categoryName"`
}

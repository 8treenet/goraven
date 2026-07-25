package vo

import "time"

type TeamProjectShareReq struct {
	ProjectName string `json:"projectName" validate:"required"`
	Description string `json:"description"`
}

type TeamProjectShareRsp struct {
	SharedId int `json:"sharedId"`
}

type TeamProjectUpdateReq struct {
	Description string `json:"description"`
}

type TeamProjectItem struct {
	Id          int       `json:"id"`
	OwnerId     string    `json:"ownerId"`
	OwnerName   string    `json:"ownerName"`
	OwnerAvatar string    `json:"ownerAvatar"`
	ProjectName string    `json:"projectName"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
	IsOwner     bool      `json:"isOwner"`
}

type TeamProjectListRsp struct {
	Items []TeamProjectItem `json:"items"`
}

type AdminSharedProjectItem struct {
	Id           int        `json:"id"`
	OwnerId      string     `json:"ownerId"`
	OwnerName    string     `json:"ownerName"`
	OwnerAvatar  string     `json:"ownerAvatar"`
	ProjectName  string     `json:"projectName"`
	Description  string     `json:"description"`
	VisitCount   int        `json:"visitCount"`
	LastActiveAt *time.Time `json:"lastActiveAt"`
	Locked       bool       `json:"locked"`
	LockedBy     string     `json:"lockedBy"`
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
}

type AdminSharedProjectListReq struct {
	Search   string `url:"search"`
	Page     int    `url:"page"`
	PageSize int    `url:"pageSize"`
}

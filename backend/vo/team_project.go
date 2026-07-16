package vo

import "time"

// --- 共享管理 ---

// TeamProjectShareReq 共享项目请求 POST /api/teamProject/share
type TeamProjectShareReq struct {
	ProjectName string `json:"projectName" validate:"required"` // projects/ 下的子目录名
	Description string `json:"description"`                     // 项目简介，可空
}

// TeamProjectShareRsp 共享项目响应
type TeamProjectShareRsp struct {
	SharedId int `json:"sharedId"`
}

// TeamProjectUpdateReq 更新简介请求 PUT /api/teamProject/:id
type TeamProjectUpdateReq struct {
	Description string `json:"description"` // 新的项目简介
}

// TeamProjectItem 团队项目列表项
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

// TeamProjectListRsp 团队项目列表响应
type TeamProjectListRsp struct {
	Items []TeamProjectItem `json:"items"`
}

// --- 管理端 ---

// AdminSharedProjectItem 管理端团队项目列表项
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

// AdminSharedProjectListReq 管理端团队项目列表请求（分页+搜索）
type AdminSharedProjectListReq struct {
	Search   string `url:"search"`   // 项目名/owner 模糊搜索
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

type TempAccessReq struct {
	Path string `json:"path" validate:"required"` // 用户空间相对路径
	Type string `json:"type" validate:"required"` // "file" 或 "dir"
}

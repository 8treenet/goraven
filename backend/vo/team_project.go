package vo

import "time"

// --- 团队项目管理 ---

// TeamProjectCreateReq 创建团队项目请求 POST /api/teamProject
type TeamProjectCreateReq struct {
	ProjectName string `json:"projectName" validate:"required"` // 项目目录名
	Description string `json:"description"`                     // 项目简介，可空
}

// TeamProjectCreateRsp 创建团队项目响应
type TeamProjectCreateRsp struct {
	Id int `json:"id"`
}

// TeamProjectUpdateReq 更新简介请求 PUT /api/teamProject/:id
type TeamProjectUpdateReq struct {
	Description string `json:"description"` // 新的项目简介
}

// TeamProjectItem 团队项目列表项
type TeamProjectItem struct {
	Id            int       `json:"id"`
	CreatorId     string    `json:"creatorId"`
	CreatorName   string    `json:"creatorName"`
	CreatorAvatar string    `json:"creatorAvatar"`
	ProjectName   string    `json:"projectName"`
	Description   string    `json:"description"`
	Access        uint8     `json:"access"`
	UpdatedAt     time.Time `json:"updatedAt"`
	IsCreator     bool      `json:"isCreator"`
}

// TeamProjectListRsp 团队项目列表响应
type TeamProjectListRsp struct {
	Items []TeamProjectItem `json:"items"`
}

// --- 管理端 ---

// AdminTeamProjectItem 管理端团队项目列表项
type AdminTeamProjectItem struct {
	Id            int        `json:"id"`
	CreatorId     string     `json:"creatorId"`
	CreatorName   string     `json:"creatorName"`
	CreatorAvatar string     `json:"creatorAvatar"`
	ProjectName   string     `json:"projectName"`
	Description   string     `json:"description"`
	Access        uint8      `json:"access"`
	VisitCount    int        `json:"visitCount"`
	LastActiveAt  *time.Time `json:"lastActiveAt"`
	Locked        bool       `json:"locked"`
	LockedBy      string     `json:"lockedBy"`
	Created       time.Time  `json:"created"`
	Updated       time.Time  `json:"updated"`
}

// AdminTeamProjectListReq 管理端团队项目列表请求（分页+搜索）
type AdminTeamProjectListReq struct {
	Search   string `url:"search"`   // 项目名模糊搜索
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// --- 成员管理 ---

// TeamProjectUserListReq 用户列表请求（成员选择器） GET /api/teamProject/users
type TeamProjectUserListReq struct {
	Page     int `url:"page"`     // 页码
	PageSize int `url:"pageSize"` // 每页条数
}

// TeamProjectUserItem 用户列表项（成员选择器）
type TeamProjectUserItem struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// TeamProjectMembersRsp 成员列表响应
type TeamProjectMembersRsp struct {
	CreatorId string   `json:"creatorId"` // 创建者 userId
	MemberIds []string `json:"memberIds"` // 成员 userId 列表（不含创建者）
}

// TeamProjectMemberUpdateReq 编辑成员请求 PUT /api/teamProject/:id/members
type TeamProjectMemberUpdateReq struct {
	AddUserIds    []string `json:"addUserIds"`    // 要添加的用户 ID 列表
	RemoveUserIds []string `json:"removeUserIds"` // 要移除的用户 ID 列表
}

// TeamProjectAccessUpdateReq 设置访问权限请求 PUT /api/teamProject/:id/access
type TeamProjectAccessUpdateReq struct {
	Access uint8 `json:"access"` // 访问权限：0全员开放 1仅成员可见
}

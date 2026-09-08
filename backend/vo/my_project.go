package vo

import "time"

// --- 个人项目（user_project）---

// MyProjectCreateReq 创建个人项目请求 POST /api/myProject/create
type MyProjectCreateReq struct {
	ProjectName string `json:"projectName" validate:"required"` // 项目目录名
	Description string `json:"description"`                     // 项目简介，可空
}

// MyProjectCreateRsp 创建个人项目响应
type MyProjectCreateRsp struct {
	Id int `json:"id"`
}

// MyProjectUpdateReq 更新个人项目请求 PUT /api/myProject/:id
// description 与 projectName 至少传一个；projectName 变更会同步改名目录并联动 session/automation 引用
type MyProjectUpdateReq struct {
	ProjectName string `json:"projectName"` // 新项目名，空表示不改名
	Description string `json:"description"` // 新简介（传该字段即更新，可为空串）
}

// MyProjectItem 个人项目列表项
type MyProjectItem struct {
	Id          int       `json:"id"`
	ProjectName string    `json:"projectName"`
	Description string    `json:"description"`
	GitUrl      string    `json:"gitUrl"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Created     time.Time `json:"created"`
}

// MyProjectListRsp 个人项目列表响应
type MyProjectListRsp struct {
	Items []MyProjectItem `json:"items"`
}

package vo

import "time"

// UserLoginReq 用户登录请求
// 前端传明文密码，后端统一做 bcrypt 哈希
type UserLoginReq struct {
	Username      string `json:"username" validate:"required"` // 用户名
	Password      string `json:"password" validate:"required"` // 明文密码
	CaptchaAnswer int    `json:"captchaAnswer"`                // 算术验证码答案；后端开启防护时必填
}

// UserLoginRsp 用户登录响应
type UserLoginRsp struct {
	AccessToken string `json:"accessToken"` // 访问令牌
}

// CaptchaReq 拉取登录验证码请求
type CaptchaReq struct {
	Username string `url:"username" validate:"required"` // 用户名
}

// CaptchaRsp 拉取登录验证码响应
// Required=false 时无需展示验证码，登录可不带 CaptchaAnswer
type CaptchaRsp struct {
	Required bool   `json:"required"`         // 是否需要验证码
	Image1   string `json:"image1,omitempty"` // 第一个加数图片（data:image/png;base64,...）
	Image2   string `json:"image2,omitempty"` // 第二个加数图片（data:image/png;base64,...）
}

// UserInfoRsp 当前用户信息响应
type UserInfoRsp struct {
	UserId   string    `json:"userId"`   // 用户 ID
	Username string    `json:"username"` // 用户名
	Email    string    `json:"email"`    // 邮箱
	Role     uint8     `json:"role"`     // 角色 0普通用户 1管理员
	Status   uint8     `json:"status"`   // 状态 0禁用 1启用
	Nickname string    `json:"nickname"` // 昵称
	Avatar   string    `json:"avatar"`   // 头像 URL
	Created  time.Time `json:"created"`  // 注册时间
}

// AdminUserListReq 管理员用户列表请求
type AdminUserListReq struct {
	Search   string `url:"search"`   // 用户名模糊搜索
	Role     *uint8 `url:"role"`     // 角色筛选（nil 不筛选）
	Page     int    `url:"page"`     // 页码
	PageSize int    `url:"pageSize"` // 每页条数
}

// AdminUserItem 管理员用户列表条目
type AdminUserItem struct {
	UserId          string     `json:"userId"`          // 用户 ID
	Username        string     `json:"username"`        // 用户名
	Nickname        string     `json:"nickname"`        // 昵称
	Email           string     `json:"email"`           // 邮箱
	Avatar          string     `json:"avatar"`          // 头像
	Role            uint8      `json:"role"`            // 角色 0普通用户 1管理员
	Status          uint8      `json:"status"`          // 状态 0禁用 1启用
	DailyTokenLimit int        `json:"dailyTokenLimit"` // 每日 Token 限额（单位 M，0=不限制）
	SessionCount    int        `json:"sessionCount"`    // 会话数
	LastActiveTime  *time.Time `json:"lastActiveTime"`  // 最后活跃时间（无会话时为 null）
	Created         time.Time  `json:"created"`         // 创建时间
}

// AdminCreateUserReq 管理员创建用户请求
type AdminCreateUserReq struct {
	Username string `json:"username" validate:"required,min=8,max=16"` // 用户名 8-16 字符
	Password string `json:"password" validate:"required"`              // 明文密码（后端做 bcrypt 哈希）
	Nickname string `json:"nickname"`                                   // 昵称
	Role     uint8  `json:"role"`                                       // 角色 0普通用户 1管理员
}

// AdminUpdateUserReq 管理员编辑用户请求
type AdminUpdateUserReq struct {
	Nickname        string `json:"nickname"`        // 昵称
	Email           string `json:"email"`           // 邮箱
	Role            *uint8 `json:"role"`            // 角色 0普通用户 1管理员（nil 不修改）
	Status          *uint8 `json:"status"`          // 状态 0禁用 1启用（nil 不修改）
	DailyTokenLimit *int   `json:"dailyTokenLimit"` // 每日 Token 限额（单位 M，0=不限制，nil 不修改）
}

// AdminBatchUserReq 管理员批量查询用户请求
type AdminBatchUserReq struct {
	UserIds []string `json:"userIds" validate:"required"` // 用户 ID 列表
}

// AdminUserDetailRsp 管理员用户详情响应
type AdminUserDetailRsp struct {
	UserId          string     `json:"userId"`          // 用户 ID
	Username        string     `json:"username"`        // 用户名
	Nickname        string     `json:"nickname"`        // 昵称
	Email           string     `json:"email"`           // 邮箱
	Avatar          string     `json:"avatar"`          // 头像
	Role            uint8      `json:"role"`            // 角色 0普通用户 1管理员
	Status          uint8      `json:"status"`          // 状态 0禁用 1启用
	DailyTokenLimit int        `json:"dailyTokenLimit"` // 每日 Token 限额（单位 M，0=不限制）
	SessionCount    int        `json:"sessionCount"`    // 会话数
	LastActiveTime  *time.Time `json:"lastActiveTime"`  // 最后活跃时间（无会话时为 null）
	Created         time.Time  `json:"created"`         // 创建时间
	Updated         time.Time  `json:"updated"`         // 更新时间
}

// AdminResetPasswordReq 管理员重置密码请求
type AdminResetPasswordReq struct {
	Password string `json:"password" validate:"required"` // 明文新密码（后端做 bcrypt 哈希）
}

// UserProfileReq 个人资料修改请求
// 仅传需要修改的字段，nil 表示不修改该字段
type UserProfileReq struct {
	Nickname *string `json:"nickname"` // 昵称（最多 20 字符，不能为空）
	Email    *string `json:"email"`    // 邮箱（空字符串表示删除邮箱）
	Avatar   *string `json:"avatar"`   // 头像路径（文件上传后返回的路径）
}

// UserPasswordReq 用户修改密码请求
// 前端传明文密码，后端统一做 bcrypt 哈希
type UserPasswordReq struct {
	CurrentPassword string `json:"currentPassword" validate:"required"` // 当前明文密码
	NewPassword     string `json:"newPassword" validate:"required"`     // 新明文密码
}

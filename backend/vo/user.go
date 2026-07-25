package vo

import "time"

type UserLoginReq struct {
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password" validate:"required"`
	CaptchaAnswer int    `json:"captchaAnswer"`
}

type UserLoginRsp struct {
	AccessToken string `json:"accessToken"`
}

type CaptchaReq struct {
	Username string `url:"username" validate:"required"`
}

type CaptchaRsp struct {
	Required bool   `json:"required"`
	Image1   string `json:"image1,omitempty"`
	Image2   string `json:"image2,omitempty"`
}

type UserInfoRsp struct {
	UserId   string    `json:"userId"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     uint8     `json:"role"`
	Status   uint8     `json:"status"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Created  time.Time `json:"created"`
}

type AdminUserListReq struct {
	Search   string `url:"search"`
	Role     *uint8 `url:"role"`
	Page     int    `url:"page"`
	PageSize int    `url:"pageSize"`
}

type AdminUserItem struct {
	UserId          string     `json:"userId"`
	Username        string     `json:"username"`
	Nickname        string     `json:"nickname"`
	Email           string     `json:"email"`
	Avatar          string     `json:"avatar"`
	Role            uint8      `json:"role"`
	Status          uint8      `json:"status"`
	DailyTokenLimit int        `json:"dailyTokenLimit"`
	SessionCount    int        `json:"sessionCount"`
	LastActiveTime  *time.Time `json:"lastActiveTime"`
	Created         time.Time  `json:"created"`
}

type AdminCreateUserReq struct {
	Username string `json:"username" validate:"required,min=8,max=16"`
	Password string `json:"password" validate:"required"`
	Nickname string `json:"nickname"`
	Role     uint8  `json:"role"`
}

type AdminUpdateUserReq struct {
	Nickname        string `json:"nickname"`
	Email           string `json:"email"`
	Role            *uint8 `json:"role"`
	Status          *uint8 `json:"status"`
	DailyTokenLimit *int   `json:"dailyTokenLimit"`
}

type AdminBatchUserReq struct {
	UserIds []string `json:"userIds" validate:"required"`
}

type AdminUserDetailRsp struct {
	UserId          string     `json:"userId"`
	Username        string     `json:"username"`
	Nickname        string     `json:"nickname"`
	Email           string     `json:"email"`
	Avatar          string     `json:"avatar"`
	Role            uint8      `json:"role"`
	Status          uint8      `json:"status"`
	DailyTokenLimit int        `json:"dailyTokenLimit"`
	SessionCount    int        `json:"sessionCount"`
	LastActiveTime  *time.Time `json:"lastActiveTime"`
	Created         time.Time  `json:"created"`
	Updated         time.Time  `json:"updated"`
}

type AdminResetPasswordReq struct {
	Password string `json:"password" validate:"required"`
}

type UserProfileReq struct {
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Avatar   *string `json:"avatar"`
}

type UserPasswordReq struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
}

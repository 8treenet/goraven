package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/user", &UserController{}, infra.NewAuth(true, "/login", "/captcha"))
	})
}

type UserController struct {
	UserSev *service.UserService
	Request *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/user
func (controller *UserController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/login", "PostLogin")
	b.Handle("GET", "/captcha", "GetCaptcha")
	b.Handle("GET", "/", "User")
	b.Handle("PUT", "/profile", "PutProfile")
	b.Handle("PUT", "/password", "PutPassword")
	b.Handle("POST", "/logout", "PostLogout")
}

// PostLogin 用户登录，路由 POST /api/user/login
func (controller *UserController) PostLogin() freedom.Result {
	var req vo.UserLoginReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if len(req.Username) < 8 {
		return &infra.JSONResponse{Error: errs.ErrUsernameTooShort}
	}

	rsp, err := controller.UserSev.Login(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// GetCaptcha 拉取登录算术验证码，路由 GET /api/user/captcha?username=xxx
// 全局错误计数达阈值时返回两张数字图片，答案按用户名存储隔离
// 返回 Required=false 表示当前未触发防护，前端无需展示验证码
func (controller *UserController) GetCaptcha() freedom.Result {
	var req vo.CaptchaReq
	if err := controller.Request.ReadQuery(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	rsp, err := controller.UserSev.GetCaptcha(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// Get 获取当前用户信息，路由 GET /api/user
func (controller *UserController) User() freedom.Result {
	rsp, err := controller.UserSev.GetUserInfo()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

// PutProfile 修改个人资料 PUT /api/user/profile
func (controller *UserController) PutProfile() freedom.Result {
	var req vo.UserProfileReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.UserSev.UpdateProfile(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// PutPassword 修改密码 PUT /api/user/password
func (controller *UserController) PutPassword() freedom.Result {
	var req vo.UserPasswordReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	if err := controller.UserSev.ChangePassword(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// PostLogout 退出登录 POST /api/user/logout
func (controller *UserController) PostLogout() freedom.Result {
	if err := controller.UserSev.Logout(); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

package controller

import (
	"raven/backend/infra"
	"raven/backend/service"
	"raven/backend/vo"
	"raven/backend/vo/errs"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/user", &UserController{}, infra.NewAuth(true, "/login", "/captcha"))
	})
}

type UserController struct {
	UserSev	*service.UserService
	Request	*infra.Request
}

func (controller *UserController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/login", "PostLogin")
	b.Handle("GET", "/captcha", "GetCaptcha")
	b.Handle("GET", "/", "User")
	b.Handle("PUT", "/profile", "PutProfile")
	b.Handle("PUT", "/password", "PutPassword")
	b.Handle("POST", "/logout", "PostLogout")
}

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

func (controller *UserController) User() freedom.Result {
	rsp, err := controller.UserSev.GetUserInfo()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

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

func (controller *UserController) PostLogout() freedom.Result {
	if err := controller.UserSev.Logout(); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

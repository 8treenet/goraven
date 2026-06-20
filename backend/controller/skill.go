package controller

import (
	"raven/backend/infra"
	"raven/backend/service"
	"raven/backend/vo"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/skills", &SkillController{}, infra.NewAuth(true))
	})
}

type SkillController struct {
	SkillSev *service.SkillService
	Request  *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/skills
func (controller *SkillController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/simpleSkills", "GetSimpleSkills")
	b.Handle("GET", "/simpleSkills/byIds", "GetSimpleSkillsByIDs")
	b.Handle("GET", "/market", "ListMarketSkills")
	b.Handle("GET", "/market/{id:int}", "GetMarketSkillDetail")
	b.Handle("GET", "/user", "ListUserSkills")
	b.Handle("GET", "/user/{id:int}", "GetUserSkillDetail")
	b.Handle("GET", "/user/{id:int}/status", "GetUserSkillStatus")
	b.Handle("PUT", "/user/{id:int}", "UpdateUserSkill")
	b.Handle("PUT", "/user/{id:int}/alwaysOn", "ToggleAlwaysOn")
	b.Handle("DELETE", "/user/{id:int}", "DeleteUserSkill")
	b.Handle("POST", "/user/refresh", "RefreshUserSkills")
	b.Handle("GET", "/categories", "ListCategories")
	b.Handle("POST", "/install", "InstallSkill")
	b.Handle("PUT", "/user/{id:int}/retry", "RetryInstallSkill")
	b.Handle("POST", "/shares", "ShareSkill")
	b.Handle("GET", "/shares", "ListSkillShares")
	b.Handle("GET", "/shares/{id:int}", "GetSkillShareDetail")
	b.Handle("PUT", "/shares/{id:int}", "UpdateSharedSkill")
	b.Handle("DELETE", "/shares/{id:int}", "DeleteSkillShare")
	b.Handle("POST", "/shares/{id:int}/install", "InstallSharedSkill")
}

func (controller *SkillController) GetSimpleSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.ListAvailableSkills(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) GetSimpleSkillsByIDs() freedom.Result {
	userId := controller.Request.GetUserId()
	var query struct {
		IDs []int `url:"ids" validate:"required"`
	}
	if err := controller.Request.ReadQuery(&query, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.ListAvailableSkillsByIDs(userId, query.IDs)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) ListMarketSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.UserMarketSkillListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.ListMarketSkillsForUser(userId, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) GetMarketSkillDetail(id int) freedom.Result {
	rsp, err := controller.SkillSev.GetMarketSkillDetailForUser(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) ListUserSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.UserSkillListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.ListUserSkills(userId, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) GetUserSkillDetail(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.GetUserSkillDetail(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) GetUserSkillStatus(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.GetUserSkillStatus(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) InstallSkill() freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.UserSkillInstallReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.InstallSkill(userId, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) RetryInstallSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.RetryInstallSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) UpdateUserSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.UserSkillUpdateReq
	if err := controller.Request.ReadJSON(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.SkillSev.UpdateUserSkill(id, userId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) ToggleAlwaysOn(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.UserSkillToggleAlwaysOnReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.SkillSev.ToggleAlwaysOn(id, userId, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) DeleteUserSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.DeleteUserSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) RefreshUserSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.RefreshUserSkills(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) ListCategories() freedom.Result {
	rsp, err := controller.SkillSev.ListSkillCategoriesForUser()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) ShareSkill() freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.ShareSkillReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.ShareSkill(userId, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) ListSkillShares() freedom.Result {
	userId := controller.Request.GetUserId()
	var req vo.SkillShareListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.SkillSev.ListSkillShares(userId, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) GetSkillShareDetail(id int) freedom.Result {
	rsp, err := controller.SkillSev.GetSkillShareDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *SkillController) UpdateSharedSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.UpdateSharedSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) DeleteSkillShare(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.DeleteSkillShare(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

func (controller *SkillController) InstallSharedSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.InstallSharedSkill(userId, id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/skills", &SkillController{}, infra.NewAuth(true))
	})
}

// SkillController 用户可用的技能相关接口
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

// GetSimpleSkills 获取用户可选技能精简列表 GET /api/skills/simpleSkills
func (controller *SkillController) GetSimpleSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.ListAvailableSkills(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetSimpleSkillsByIDs 根据指定的 userSkillId 列表获取技能精简列表 GET /api/skills/simpleSkills/byIds?ids=1&ids=2
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

// ListMarketSkills 获取市场技能列表（用户视角） GET /api/skills/market
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

// GetMarketSkillDetail 获取市场技能详情 GET /api/skills/market/:id
func (controller *SkillController) GetMarketSkillDetail(id int) freedom.Result {
	rsp, err := controller.SkillSev.GetMarketSkillDetailForUser(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ListUserSkills 获取用户已安装技能列表 GET /api/skills/user
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

// GetUserSkillDetail 获取用户技能详情 GET /api/skills/user/:id
func (controller *SkillController) GetUserSkillDetail(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.GetUserSkillDetail(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetUserSkillStatus 获取用户技能当前安装状态 GET /api/skills/user/:id/status
// 轻量接口，前端轮询安装进度时使用
func (controller *SkillController) GetUserSkillStatus(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.GetUserSkillStatus(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// InstallSkill 安装市场技能 POST /api/skills/install
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

// RetryInstallSkill 重试安装失败技能 PUT /api/skills/user/:id/retry
func (controller *SkillController) RetryInstallSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.RetryInstallSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

// UpdateUserSkill 编辑用户技能 PUT /api/skills/user/:id
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

// ToggleAlwaysOn 切换始终启用 PUT /api/skills/user/:id/alwaysOn
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

// DeleteUserSkill 删除用户技能 DELETE /api/skills/user/:id
func (controller *SkillController) DeleteUserSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.DeleteUserSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

// RefreshUserSkills 刷新同步用户创建的技能 POST /api/skills/user/refresh
func (controller *SkillController) RefreshUserSkills() freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.RefreshUserSkills(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ListCategories 获取技能分类列表 GET /api/skills/categories
func (controller *SkillController) ListCategories() freedom.Result {
	rsp, err := controller.SkillSev.ListSkillCategoriesForUser()
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ShareSkill 共享技能到团队 POST /api/skills/shares
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

// ListSkillShares 团队技能列表 GET /api/skills/shares
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

// GetSkillShareDetail 团队技能详情 GET /api/skills/shares/:id
func (controller *SkillController) GetSkillShareDetail(id int) freedom.Result {
	rsp, err := controller.SkillSev.GetSkillShareDetail(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// UpdateSharedSkill 更新共享技能 PUT /api/skills/shares/:id
func (controller *SkillController) UpdateSharedSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.UpdateSharedSkill(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

// DeleteSkillShare 删除共享技能 DELETE /api/skills/shares/:id
func (controller *SkillController) DeleteSkillShare(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.SkillSev.DeleteSkillShare(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{}
}

// InstallSharedSkill 安装团队共享技能 POST /api/skills/shares/:id/install
func (controller *SkillController) InstallSharedSkill(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	rsp, err := controller.SkillSev.InstallSharedSkill(userId, id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

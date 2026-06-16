package controller

import (
	"raven/backend/infra"
	"raven/backend/service"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/dashboard", &DashboardController{}, infra.NewAuth(true))
	})
}

type DashboardController struct {
	DashboardSev	*service.DashboardService
	Request		*infra.Request
}

func (controller *DashboardController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "Dashboard")
	b.Handle("GET", "/tokenTrend", "GetTokenTrend")
}

func (controller *DashboardController) Dashboard() freedom.Result {

	userId := controller.Request.GetUserId()
	refresh := controller.Request.Worker().IrisContext().URLParam("refresh") == "true"
	rsp, err := controller.DashboardSev.GetUserDashboard(userId, refresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *DashboardController) GetTokenTrend() freedom.Result {
	var req struct {
		Days	int	`url:"days"`
		Refresh	string	`url:"refresh"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	userId := controller.Request.GetUserId()
	refresh := req.Refresh == "true"

	rsp, err := controller.DashboardSev.GetUserTokenTrend(userId, req.Days, refresh)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

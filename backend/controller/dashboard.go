package controller

import (
	"goraven/backend/infra"
	"goraven/backend/repository/seed/mock"
	"goraven/backend/service"
	"goraven/config"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/dashboard", &DashboardController{}, infra.NewAuth(true))
	})
}

type DashboardController struct {
	DashboardSev *service.DashboardService
	Request      *infra.Request
}

func (controller *DashboardController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "Dashboard")
	b.Handle("GET", "/tokenTrend", "GetTokenTrend")
	b.Handle("GET", "/modelUsage", "GetModelUsage")
}

func (controller *DashboardController) Dashboard() freedom.Result {
	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildUserDashboard()}
	}

	userId := controller.Request.GetUserId()
	rsp, err := controller.DashboardSev.GetUserDashboard(userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *DashboardController) GetTokenTrend() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildUserTokenTrend(req.Days)}
	}

	userId := controller.Request.GetUserId()

	rsp, err := controller.DashboardSev.GetUserTokenTrend(userId, req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

func (controller *DashboardController) GetModelUsage() freedom.Result {
	var req struct {
		Days int `url:"days"`
	}
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if req.Days <= 0 {
		req.Days = 30
	}

	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildUserModelUsage(req.Days)}
	}

	userId := controller.Request.GetUserId()

	rsp, err := controller.DashboardSev.GetUserModelUsage(userId, req.Days)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: rsp}
}

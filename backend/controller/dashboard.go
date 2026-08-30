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

// DashboardController 用户仪表盘控制器
// 路由前缀 /api/dashboard，所有接口需登录
type DashboardController struct {
	DashboardSev *service.DashboardService // 仪表盘服务
	Request      *infra.Request            // 请求工具
}

// BeforeActivation 绑定用户仪表盘路由
func (controller *DashboardController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "Dashboard")
	b.Handle("GET", "/tokenTrend", "GetTokenTrend")
	b.Handle("GET", "/modelUsage", "GetModelUsage")
}

// Get 用户仪表盘聚合数据 GET /api/dashboard
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

// GetTokenTrend 用户个人 Token 趋势 GET /api/dashboard/tokenTrend?days=30
func (controller *DashboardController) GetTokenTrend() freedom.Result {
	var req struct {
		Days int `url:"days"` // 天数，默认 30
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

// GetModelUsage 用户个人模型使用分布 GET /api/dashboard/modelUsage?days=30
func (controller *DashboardController) GetModelUsage() freedom.Result {
	var req struct {
		Days int `url:"days"` // 天数，默认 30
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

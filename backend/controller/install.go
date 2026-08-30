package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/config"

	"github.com/8treenet/freedom"
)

func init() {
	if config.Get().System.Initialized {
		return
	}
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/install", &InstallController{})
	})
}

// InstallController 系统安装控制器，仅在系统未初始化时注册
type InstallController struct {
	InstallSev *service.InstallService
	Request    *infra.Request
}

// BeforeActivation .
func (controller *InstallController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/check-db", "CheckDB")
	b.Handle("POST", "/check-redis", "CheckRedis")
	b.Handle("POST", "/init", "DoInit")
}

// CheckDB 测试数据库连接，路由 POST /api/install/check-db
func (controller *InstallController) CheckDB() freedom.Result {
	var req vo.InstallDBCheckReq
	if err := 	controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := req.Check(); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := 	controller.InstallSev.CheckDB(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: "ok"}
}

// CheckRedis 测试 Redis 连接，路由 POST /api/install/check-redis
func (controller *InstallController) CheckRedis() freedom.Result {
	var req vo.InstallRedisCheckReq
	if err := 	controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := 	controller.InstallSev.CheckRedis(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: "ok"}
}

// DoInit 执行系统初始化，路由 POST /api/install/init
func (controller *InstallController) DoInit() freedom.Result {
	var req vo.InstallInitReq
	if err := 	controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := req.Check(); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := 	controller.InstallSev.Init(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

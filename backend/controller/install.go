package controller

import (
	"raven/backend/infra"
	"raven/backend/service"
	"raven/backend/vo"
	"raven/config"

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

type InstallController struct {
	InstallSev	*service.InstallService
	Request		*infra.Request
}

func (controller *InstallController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("POST", "/check-db", "CheckDB")
	b.Handle("POST", "/check-redis", "CheckRedis")
	b.Handle("POST", "/init", "DoInit")
}

func (controller *InstallController) CheckDB() freedom.Result {
	var req vo.InstallDBCheckReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := req.Check(); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.InstallSev.CheckDB(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: "ok"}
}

func (controller *InstallController) CheckRedis() freedom.Result {
	var req vo.InstallRedisCheckReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.InstallSev.CheckRedis(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: "ok"}
}

func (controller *InstallController) DoInit() freedom.Result {
	var req vo.InstallInitReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := req.Check(); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.InstallSev.Init(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

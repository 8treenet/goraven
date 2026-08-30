package controller

import (
	"goraven/backend/infra"
	"goraven/backend/repository/seed/mock"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/automationTasks", &AutomationController{}, infra.NewAuth(true))
	})
}

// AutomationController 自动化任务控制器
// GET /api/automationTasks                                        任务分页列表（不含需求描述）
// GET /api/automationTasks/{id}                                   任务详情（含需求描述）
// GET /api/automationTasks/{id}/executions                        执行记录分页
// GET /api/automationTasks/{id}/executions/{executionId}/messages 执行问答对
// DELETE /api/automationTasks/{id}                                删除任务（软删除）
// PUT /api/automationTasks/{id}/status                            启用/停用任务
// PUT /api/automationTasks/{id}/requirement                       修改任务需求描述
// POST /api/automationTasks/{id}/execute                          立即执行任务
type AutomationController struct {
	AutomationSev *service.AutomationService
	ChatService   *service.ChatService
	Request       *infra.Request
}

// BeforeActivation 绑定路由前缀 /api/automationTasks
func (controller *AutomationController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/", "ListTasks")
	b.Handle("GET", "/{id:int}", "GetTask")
	b.Handle("GET", "/{id:int}/executions", "ListExecutions")
	b.Handle("GET", "/{id:int}/executions/{executionId:int}/messages", "GetExecutionMessages")
	b.Handle("DELETE", "/{id:int}", "DeleteTask")
	b.Handle("PUT", "/{id:int}/status", "UpdateTaskStatus")
	b.Handle("PUT", "/{id:int}/requirement", "UpdateTaskRequirement")
	b.Handle("POST", "/{id:int}/execute", "ExecuteTask")
}

// ListTasks 获取自动化任务分页列表 GET /api/automationTasks?page=1&pageSize=20&status=1
func (controller *AutomationController) ListTasks() freedom.Result {
	req := &vo.AutomationTaskListReq{}
	if err := controller.Request.ReadQuery(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAutomationTasks(req.Page, req.PageSize, req.Status)}
	}
	userId := controller.Request.GetUserId()
	rsp, err := controller.AutomationSev.ListTasks(userId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetTask 获取自动化任务详情 GET /api/automationTasks/:id
func (controller *AutomationController) GetTask(id int) freedom.Result {
	if config.Get().Behavior.PreviewUser != "" {
		if rsp := mock.BuildAutomationTask(id); rsp != nil {
			return &infra.JSONResponse{Object: rsp}
		}
		return &infra.JSONResponse{Error: errs.ErrAutomationTaskNotFound}
	}
	userId := controller.Request.GetUserId()
	rsp, err := controller.AutomationSev.GetTask(id, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ListExecutions 获取任务的执行记录分页 GET /api/automationTasks/:id/executions?page=1&pageSize=20
func (controller *AutomationController) ListExecutions(id int) freedom.Result {
	req := &vo.AutomationExecutionListReq{}
	if err := controller.Request.ReadQuery(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAutomationExecutions(id, req.Page, req.PageSize)}
	}
	userId := controller.Request.GetUserId()
	rsp, err := controller.AutomationSev.ListExecutions(id, userId, req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GetExecutionMessages 获取执行记录的问答对 GET /api/automationTasks/:id/executions/:executionId/messages
func (controller *AutomationController) GetExecutionMessages(id int, executionId int) freedom.Result {
	if config.Get().Behavior.PreviewUser != "" {
		return &infra.JSONResponse{Object: mock.BuildAutomationQA(id, executionId)}
	}
	userId := controller.Request.GetUserId()
	rsp, err := controller.AutomationSev.GetExecutionQA(id, executionId, userId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// DeleteTask 删除自动化任务（软删除） DELETE /api/automationTasks/:id
func (controller *AutomationController) DeleteTask(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.AutomationSev.DeleteTask(id, userId); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateTaskStatus 启用/停用任务 PUT /api/automationTasks/:id/status
func (controller *AutomationController) UpdateTaskStatus(id int) freedom.Result {
	req := &vo.AutomationTaskStatusReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	userId := controller.Request.GetUserId()
	if err := controller.AutomationSev.UpdateTaskStatus(id, userId, req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateTaskRequirement 修改任务需求描述 PUT /api/automationTasks/:id/requirement
func (controller *AutomationController) UpdateTaskRequirement(id int) freedom.Result {
	req := &vo.AutomationTaskRequirementReq{}
	if err := controller.Request.ReadJSON(req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	userId := controller.Request.GetUserId()
	if err := controller.AutomationSev.UpdateTaskRequirement(id, userId, req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ExecuteTask 立即执行任务 POST /api/automationTasks/:id/execute
func (controller *AutomationController) ExecuteTask(id int) freedom.Result {
	userId := controller.Request.GetUserId()
	if err := controller.AutomationSev.ExecuteTask(id, userId, controller.ChatService); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

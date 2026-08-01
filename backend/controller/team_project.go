package controller

import (
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/teamProject", &TeamProjectController{}, infra.NewAuth(true))
	})
}

// TeamProjectController 团队项目控制器
type TeamProjectController struct {
	TPSev   *service.TeamProjectService
	Worker  freedom.Worker
	Request *infra.Request
}

// BeforeActivation 注册路由
func (controller *TeamProjectController) BeforeActivation(b freedom.BeforeActivation) {
	// 项目管理
	b.Handle("GET", "/list", "List")
	b.Handle("GET", "/{id:int}", "Get")
	b.Handle("POST", "/create", "Create")
	b.Handle("DELETE", "/{id:int}", "DeleteProject")
	b.Handle("PUT", "/{id:int}", "UpdateDescription")
	// 文件操作
	b.Handle("GET", "/{id:int}/list", "ListFiles")
	b.Handle("POST", "/{id:int}/upload", "Upload")
	b.Handle("DELETE", "/{id:int}/delete", "Delete")
	b.Handle("PUT", "/{id:int}/rename", "Rename")
	b.Handle("POST", "/{id:int}/mkdir", "Mkdir")
	b.Handle("POST", "/{id:int}/compress", "Compress")
	b.Handle("POST", "/{id:int}/decompress", "Decompress")
	b.Handle("GET", "/{id:int}/usage", "Usage")
	// 下载与预览
	b.Handle("GET", "/{id:int}/download/{p:path}", "Download")
	b.Handle("POST", "/{id:int}/access", "CreateTempAccess")
	// 成员管理
	b.Handle("GET", "/users", "ListUsers")
	b.Handle("GET", "/{id:int}/members", "ListMembers")
	b.Handle("PUT", "/{id:int}/members", "UpdateMembers")
	b.Handle("PUT", "/{id:int}/access", "UpdateAccess")
}

// List 列出所有团队项目 GET /api/teamProject/list
func (controller *TeamProjectController) List() freedom.Result {
	rsp, err := controller.TPSev.List(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Get 查询单个团队项目详情 GET /api/teamProject/:id
func (controller *TeamProjectController) Get(id int) freedom.Result {
	rsp, err := controller.TPSev.Get(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Create 创建团队项目 POST /api/teamProject
func (controller *TeamProjectController) Create() freedom.Result {
	var req vo.TeamProjectCreateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.Create(
		controller.Request.GetUserId(),
		req.ProjectName,
		req.Description,
	)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// DeleteProject 删除团队项目 DELETE /api/teamProject/:id
func (controller *TeamProjectController) DeleteProject(id int) freedom.Result {
	if err := controller.TPSev.DeleteProject(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateDescription 更新简介 PUT /api/teamProject/:id
func (controller *TeamProjectController) UpdateDescription(id int) freedom.Result {
	var req vo.TeamProjectUpdateReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.UpdateDescription(controller.Request.GetUserId(), id, req.Description); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ListFiles 列出项目内文件 GET /api/teamProject/:id/list
func (controller *TeamProjectController) ListFiles(id int) freedom.Result {
	var req vo.FileManagerListReq
	if err := controller.Request.ReadQuery(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.ListFiles(id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Upload 上传文件到项目内 POST /api/teamProject/:id/upload
func (controller *TeamProjectController) Upload(id int) freedom.Result {
	var req vo.FileManagerUploadReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.Upload(id, controller.Request.GetUserId(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Delete 删除项目内文件 DELETE /api/teamProject/:id/delete
func (controller *TeamProjectController) Delete(id int) freedom.Result {
	var req vo.FileManagerDeleteReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.Delete(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Rename 重命名项目内文件 PUT /api/teamProject/:id/rename
func (controller *TeamProjectController) Rename(id int) freedom.Result {
	var req vo.FileManagerRenameReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.Rename(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Mkdir 新建目录 POST /api/teamProject/:id/mkdir
func (controller *TeamProjectController) Mkdir(id int) freedom.Result {
	var req vo.FileManagerMkdirReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.Mkdir(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Compress 压缩项目内文件 POST /api/teamProject/:id/compress
func (controller *TeamProjectController) Compress(id int) freedom.Result {
	var req vo.FileManagerCompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.Compress(id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Decompress 解压项目内 zip POST /api/teamProject/:id/decompress
func (controller *TeamProjectController) Decompress(id int) freedom.Result {
	var req vo.FileManagerDecompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.Decompress(id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Usage 项目磁盘使用统计 GET /api/teamProject/:id/usage
func (controller *TeamProjectController) Usage(id int) freedom.Result {
	rsp, err := controller.TPSev.Usage(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// CreateTempAccess 为团队项目内文件/目录创建临时访问凭证 POST /api/teamProject/:id/access
func (controller *TeamProjectController) CreateTempAccess(id int) freedom.Result {
	var req vo.TempAccessReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.CreateTempAccess(id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Download 下载团队项目内文件 GET /api/teamProject/:id/download/:p
func (controller *TeamProjectController) Download(id int, p string) {
	ctx := controller.Worker.IrisContext()
	reqPath := "/" + strings.TrimPrefix(p, "/")
	if reqPath == "/" {
		ctx.StatusCode(400)
		ctx.WriteString("path is required")
		return
	}
	absPath, fileName, err := controller.TPSev.Download(id, strings.TrimPrefix(reqPath, "/"))
	if err != nil {
		ctx.StatusCode(404)
		ctx.WriteString(err.Error())
		return
	}
	ctx.Header("Cache-Control", "max-age=86400")
	ctx.SendFile(absPath, fileName)
}

// ListMembers 查询项目成员列表 GET /api/teamProject/:id/members
func (controller *TeamProjectController) ListMembers(id int) freedom.Result {
	rsp, err := controller.TPSev.ListMembers(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// UpdateMembers 编辑项目成员 PUT /api/teamProject/:id/members
func (controller *TeamProjectController) UpdateMembers(id int) freedom.Result {
	var req vo.TeamProjectMemberUpdateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.UpdateMembers(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// UpdateAccess 设置访问权限 PUT /api/teamProject/:id/access
func (controller *TeamProjectController) UpdateAccess(id int) freedom.Result {
	var req vo.TeamProjectAccessUpdateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.UpdateAccess(controller.Request.GetUserId(), id, req.Access); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ListUsers 分页查询用户列表 GET /api/teamProject/users?page=1&pageSize=10
func (controller *TeamProjectController) ListUsers() freedom.Result {
	var req vo.TeamProjectUserListReq
	if err := controller.Request.ReadQuery(&req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.ListUsers(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

package controller

import (
	"strings"

	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/myProject", &MyProjectController{}, infra.NewAuth(true))
	})
}

// MyProjectController 个人项目控制器
type MyProjectController struct {
	MPSev   *service.MyProjectService
	Worker  freedom.Worker
	Request *infra.Request
}

// BeforeActivation 注册路由
func (controller *MyProjectController) BeforeActivation(b freedom.BeforeActivation) {
	// 项目管理
	b.Handle("GET", "/list", "List")
	b.Handle("GET", "/{id:int}", "Get")
	b.Handle("POST", "/create", "Create")
	b.Handle("PUT", "/{id:int}", "Update")
	b.Handle("DELETE", "/{id:int}", "DeleteProject")
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
}

// List 列出我的项目 GET /api/myProject/list
func (controller *MyProjectController) List() freedom.Result {
	rsp, err := controller.MPSev.List(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Get 项目详情 GET /api/myProject/:id
func (controller *MyProjectController) Get(id int) freedom.Result {
	rsp, err := controller.MPSev.Get(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Create 创建项目 POST /api/myProject/create
func (controller *MyProjectController) Create() freedom.Result {
	var req vo.MyProjectCreateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.MPSev.Create(controller.Request.GetUserId(), req.ProjectName, req.Description)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Update 更新项目（简介/改名） PUT /api/myProject/:id
func (controller *MyProjectController) Update(id int) freedom.Result {
	var req vo.MyProjectUpdateReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.MPSev.Update(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// DeleteProject 删除项目 DELETE /api/myProject/:id
func (controller *MyProjectController) DeleteProject(id int) freedom.Result {
	if err := controller.MPSev.DeleteProject(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ListFiles 项目内文件列表 GET /api/myProject/:id/list
func (controller *MyProjectController) ListFiles(id int) freedom.Result {
	var req vo.FileManagerListReq
	if err := controller.Request.ReadQuery(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.MPSev.ListFiles(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Upload 上传文件到项目内 POST /api/myProject/:id/upload
func (controller *MyProjectController) Upload(id int) freedom.Result {
	var req vo.FileManagerUploadReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.MPSev.Upload(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Delete 删除项目内文件 DELETE /api/myProject/:id/delete
func (controller *MyProjectController) Delete(id int) freedom.Result {
	var req vo.FileManagerDeleteReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.MPSev.Delete(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Rename 重命名项目内文件 PUT /api/myProject/:id/rename
func (controller *MyProjectController) Rename(id int) freedom.Result {
	var req vo.FileManagerRenameReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.MPSev.Rename(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Mkdir 项目内新建目录 POST /api/myProject/:id/mkdir
func (controller *MyProjectController) Mkdir(id int) freedom.Result {
	var req vo.FileManagerMkdirReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.MPSev.Mkdir(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Compress 压缩 POST /api/myProject/:id/compress
func (controller *MyProjectController) Compress(id int) freedom.Result {
	var req vo.FileManagerCompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.MPSev.Compress(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Decompress 解压 POST /api/myProject/:id/decompress
func (controller *MyProjectController) Decompress(id int) freedom.Result {
	var req vo.FileManagerDecompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.MPSev.Decompress(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// Usage 磁盘统计 GET /api/myProject/:id/usage
func (controller *MyProjectController) Usage(id int) freedom.Result {
	rsp, err := controller.MPSev.Usage(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// CreateTempAccess 临时下载凭证 POST /api/myProject/:id/access
func (controller *MyProjectController) CreateTempAccess(id int) freedom.Result {
	var req vo.TempAccessReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.MPSev.CreateTempAccess(
		controller.Request.GetUserId(),
		infra.GetUserName(controller.Worker),
		id, &req,
	)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// Download 下载项目内文件 GET /api/myProject/:id/download/:p
func (controller *MyProjectController) Download(id int, p string) {
	ctx := controller.Worker.IrisContext()
	reqPath := "/" + strings.TrimPrefix(p, "/")
	if reqPath == "/" {
		ctx.StatusCode(400)
		ctx.WriteString("path is required")
		return
	}
	absPath, fileName, err := controller.MPSev.Download(controller.Request.GetUserId(), id, strings.TrimPrefix(reqPath, "/"))
	if err != nil {
		ctx.StatusCode(404)
		ctx.WriteString(err.Error())
		return
	}
	ctx.Header("Cache-Control", "max-age=86400")
	ctx.SendFile(absPath, fileName)
}

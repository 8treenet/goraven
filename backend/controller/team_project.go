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

type TeamProjectController struct {
	TPSev   *service.TeamProjectService
	Worker  freedom.Worker
	Request *infra.Request
}

func (controller *TeamProjectController) BeforeActivation(b freedom.BeforeActivation) {

	b.Handle("GET", "/list", "List")
	b.Handle("GET", "/{id:int}", "Get")
	b.Handle("POST", "/share", "Share")
	b.Handle("DELETE", "/{id:int}", "Unshare")
	b.Handle("PUT", "/{id:int}", "UpdateDescription")

	b.Handle("GET", "/{id:int}/list", "ListFiles")
	b.Handle("POST", "/{id:int}/upload", "Upload")
	b.Handle("DELETE", "/{id:int}/delete", "Delete")
	b.Handle("PUT", "/{id:int}/rename", "Rename")
	b.Handle("POST", "/{id:int}/mkdir", "Mkdir")
	b.Handle("POST", "/{id:int}/compress", "Compress")
	b.Handle("POST", "/{id:int}/decompress", "Decompress")
	b.Handle("GET", "/{id:int}/usage", "Usage")

	b.Handle("GET", "/{id:int}/download/{p:path}", "Download")
	b.Handle("POST", "/{id:int}/access", "CreateTempAccess")
}

func (controller *TeamProjectController) List() freedom.Result {
	rsp, err := controller.TPSev.List(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *TeamProjectController) Get(id int) freedom.Result {
	rsp, err := controller.TPSev.Get(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *TeamProjectController) Share() freedom.Result {
	var req vo.TeamProjectShareReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.Share(
		controller.Request.GetUserId(),
		req.ProjectName,
		req.Description,
	)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *TeamProjectController) Unshare(id int) freedom.Result {
	if err := controller.TPSev.Unshare(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

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

func (controller *TeamProjectController) Usage(id int) freedom.Result {
	rsp, err := controller.TPSev.Usage(id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

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

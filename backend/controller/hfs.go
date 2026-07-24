package controller

import (
	"os"
	"path/filepath"
	"raven/backend/infra"
	"raven/backend/service"
	"raven/backend/vo"
	"raven/config"
	"raven/core/sandbox"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/hfs", &HFSController{}, infra.NewAuth(true, "/public"))
	})
}

type HFSController struct {
	HFSSev  *service.HFSService
	Worker  freedom.Worker
	Request *infra.Request
}

func (controller *HFSController) BeforeActivation(b freedom.BeforeActivation) {

	b.Handle("GET", "/public/{linkId:string}", "PublicDownload")
	b.Handle("GET", "/private", "PrivateDownload")
	b.Handle("GET", "/file/{p:path}", "AbsoluteDownload")

	b.Handle("POST", "/upload/create", "CreateUpload")
	b.Handle("PUT", "/upload/chunk", "UploadChunk")
	b.Handle("POST", "/upload/merge", "MergeUpload")

	b.Handle("POST", "/assets", "Assets")
}

func (controller *HFSController) PublicDownload(linkId string) {

	pureLinkId := strings.TrimSuffix(linkId, filepath.Ext(linkId))
	_, userName, filePath, fileName, err := controller.HFSSev.ResolveFile(pureLinkId)
	if err != nil {
		controller.Worker.IrisContext().StatusCode(404)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}

	sb, sberr := sandbox.NewSandbox(userName)
	if sberr != nil {
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(sberr.Error())
		return
	}
	exists, err := sb.Exists(filepath.Join(sb.GetWorkspace(), filePath))
	if err != nil {
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}
	if !exists {
		controller.Worker.IrisContext().StatusCode(404)
		controller.Worker.IrisContext().WriteString("File not found")
		return
	}
	absPath, err := sb.Download(filepath.Join(sb.GetWorkspace(), filePath))
	if err != nil {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}

	infra.ServeFile(controller.Worker.IrisContext(), absPath, fileName)
}

func (controller *HFSController) PrivateDownload() {

	var req struct {
		Path string `url:"path" validate:"required"`
	}
	if err := controller.Request.ReadQuery(&req, true); err != nil {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}
	sb, sberr := sandbox.NewSandbox(controller.Request.GetUserName())
	if sberr != nil {
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(sberr.Error())
		return
	}
	exists, err := sb.Exists(filepath.Join(sb.GetWorkspace(), req.Path))
	if err != nil {
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}
	if !exists {
		controller.Worker.IrisContext().StatusCode(404)
		controller.Worker.IrisContext().WriteString("File not found")
		return
	}
	absPath, err := sb.Download(filepath.Join(sb.GetWorkspace(), req.Path))
	if err != nil {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}

	infra.ServeFile(controller.Worker.IrisContext(), absPath, filepath.Base(req.Path))
}

func (controller *HFSController) AbsoluteDownload(p string) {
	absPath := "/" + strings.TrimPrefix(p, "/")
	if absPath == "/" {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString("path is required")
		return
	}
	cleanAbs := filepath.Clean(absPath)
	userSpaceRoot := filepath.Clean(config.Get().Paths.UserSpace)
	if !strings.HasPrefix(cleanAbs, userSpaceRoot+string(filepath.Separator)) {
		controller.Worker.IrisContext().StatusCode(403)
		controller.Worker.IrisContext().WriteString("path is outside user space")
		return
	}
	info, err := os.Stat(cleanAbs)
	if err != nil {
		if os.IsNotExist(err) {
			controller.Worker.IrisContext().StatusCode(404)
			controller.Worker.IrisContext().WriteString("File not found")
			return
		}
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}
	if info.IsDir() {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString("path is a directory")
		return
	}
	infra.ServeFile(controller.Worker.IrisContext(), cleanAbs, filepath.Base(cleanAbs))
}

func (controller *HFSController) CreateUpload() freedom.Result {
	var req vo.ChunkUploadCreateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	resp, err := controller.HFSSev.CreateUpload(controller.Request.GetUserId(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: resp}
}

func (controller *HFSController) UploadChunk() freedom.Result {
	var req vo.ChunkUploadReq
	if err := controller.Request.ReadQuery(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	file, _, err := controller.Worker.IrisContext().FormFile("file")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	defer file.Close()

	if err := controller.HFSSev.UploadChunk(controller.Request.GetUserId(), &req, file); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *HFSController) MergeUpload() freedom.Result {
	var req vo.ChunkMergeReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	resp, err := controller.HFSSev.MergeUpload(controller.Request.GetUserId(), req.UploadId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: resp}
}

func (controller *HFSController) Assets() freedom.Result {
	var req vo.AssetsReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	resp, err := controller.HFSSev.CommitAssets(controller.Request.GetUserId(), req.UploadId)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: resp}
}

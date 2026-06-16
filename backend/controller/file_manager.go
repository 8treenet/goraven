package controller

import (
	"fmt"
	"path/filepath"
	"raven/backend/infra"
	"raven/backend/service"
	"raven/backend/vo"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/fileManager", &FileManagerController{}, infra.NewAuth(true))
	})
}

type FileManagerController struct {
	FMSev	*service.FileManagerService
	Worker	freedom.Worker
	Request	*infra.Request
}

func (controller *FileManagerController) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/list", "List")
	b.Handle("POST", "/upload", "Upload")
	b.Handle("POST", "/mkdir", "Mkdir")
	b.Handle("PUT", "/rename", "Rename")
	b.Handle("DELETE", "/delete", "Delete")
	b.Handle("POST", "/compress", "Compress")
	b.Handle("POST", "/decompress", "Decompress")
	b.Handle("GET", "/usage", "Usage")
	b.Handle("GET", "/profile", "ProfileList")
	b.Handle("POST", "/profile", "ProfileCreate")
	b.Handle("PUT", "/profile", "ProfileUpdate")
	b.Handle("DELETE", "/profile", "ProfileDelete")
}

func (controller *FileManagerController) List() freedom.Result {
	var req vo.FileManagerListReq
	if err := controller.Request.ReadQuery(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.FMSev.List(controller.Request.GetUserId(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *FileManagerController) Upload() freedom.Result {
	var req vo.FileManagerUploadReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPath(req.Dir); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.FMSev.Upload(controller.Request.GetUserId(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *FileManagerController) Mkdir() freedom.Result {
	var req vo.FileManagerMkdirReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPath(req.Path); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.Mkdir(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) Rename() freedom.Result {
	var req vo.FileManagerRenameReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPaths(req.OldPath, req.NewPath); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.Rename(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) Delete() freedom.Result {
	var req vo.FileManagerDeleteReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPaths(req.Paths...); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.Delete(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) Compress() freedom.Result {
	var req vo.FileManagerCompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPaths(req.Paths...); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.FMSev.Compress(controller.Request.GetUserId(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *FileManagerController) Decompress() freedom.Result {
	var req vo.FileManagerDecompressReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := checkSkillsPath(req.Path); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.Decompress(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) Usage() freedom.Result {
	rsp, err := controller.FMSev.Usage(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *FileManagerController) ProfileList() freedom.Result {
	rsp, err := controller.FMSev.ProfileList(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

func (controller *FileManagerController) ProfileCreate() freedom.Result {
	var req vo.FileManagerProfileCreateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.ProfileCreate(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) ProfileUpdate() freedom.Result {
	var req vo.FileManagerProfileUpdateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.ProfileUpdate(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func (controller *FileManagerController) ProfileDelete() freedom.Result {
	var req vo.FileManagerProfileDeleteReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.FMSev.ProfileDelete(controller.Request.GetUserId(), &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

func checkSkillsPath(path string) error {
	cleaned := filepath.Clean(path)
	if cleaned == "skills" || strings.HasPrefix(cleaned, "skills/") {
		return fmt.Errorf("skills directory is managed by system, file operations not allowed")
	}
	return nil
}

func checkSkillsPaths(paths ...string) error {
	for _, p := range paths {
		if err := checkSkillsPath(p); err != nil {
			return err
		}
	}
	return nil
}

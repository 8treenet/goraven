package controller

import (
	"fmt"
	"path/filepath"
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/fileManager", &FileManagerController{}, infra.NewAuth(true))
	})
}

// FileManagerController 文件管理控制器
type FileManagerController struct {
	FMSev    *service.FileManagerService
	Worker   freedom.Worker
	Request  *infra.Request
}

// BeforeActivation 注册路由
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

// List 列出指定目录的文件和子目录 GET /api/fileManager/list
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

// Upload 提交文件到用户空间 POST /api/fileManager/upload
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

// Mkdir 创建目录 POST /api/fileManager/mkdir
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

// Rename 重命名文件或目录 PUT /api/fileManager/rename
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

// Delete 删除文件或目录 DELETE /api/fileManager/delete
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

// Compress 压缩为 zip POST /api/fileManager/compress
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

// Decompress 解压 zip 文件 POST /api/fileManager/decompress
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

// Usage 返回磁盘使用统计 GET /api/fileManager/usage
func (controller *FileManagerController) Usage() freedom.Result {
	rsp, err := controller.FMSev.Usage(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ProfileList 列出 .profile 中的环境变量 GET /api/fileManager/profile
func (controller *FileManagerController) ProfileList() freedom.Result {
	rsp, err := controller.FMSev.ProfileList(controller.Request.GetUserId())
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// ProfileCreate 新增环境变量 POST /api/fileManager/profile
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

// ProfileUpdate 更新环境变量 PUT /api/fileManager/profile
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

// ProfileDelete 删除环境变量 DELETE /api/fileManager/profile
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

// checkSkillsPath 检查单个路径是否为 skills 目录
// skills 是用户技能目录，通过数据库管理，禁止文件操作
// filepath.Clean 将 skills/、./skills 归一化为 skills，不会误伤 documents/skills 等子目录
func checkSkillsPath(path string) error {
	cleaned := filepath.Clean(path)
	if cleaned == "skills" || strings.HasPrefix(cleaned, "skills/") {
		return fmt.Errorf("skills directory is managed by system, file operations not allowed")
	}
	return nil
}

// checkSkillsPaths 检查多个路径是否包含 skills 目录
func checkSkillsPaths(paths ...string) error {
	for _, p := range paths {
		if err := checkSkillsPath(p); err != nil {
			return err
		}
	}
	return nil
}

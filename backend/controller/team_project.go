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
	// 项目 Git
	b.Handle("POST", "/git/test", "TestGitRemoteDirect")
	b.Handle("GET", "/{id:int}/git", "GetGitStatus")
	b.Handle("POST", "/{id:int}/git/clone-retry", "RetryGitClone")
	b.Handle("POST", "/{id:int}/git/resolve-unrelated", "ResolveGitUnrelated")
	b.Handle("POST", "/{id:int}/git/commit", "GitCommit")
	b.Handle("POST", "/{id:int}/git/push", "GitPush")
	b.Handle("POST", "/{id:int}/git/pull", "GitPull")
	b.Handle("GET", "/{id:int}/git/log", "GitLog")
	b.Handle("GET", "/{id:int}/git/diff", "GitDiff")
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
// Git 来源校验、项目创建与 Git 初始化/克隆均由服务层编排，失败自动回滚。
func (controller *TeamProjectController) Create() freedom.Result {
	var req vo.TeamProjectCreateReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.Create(controller.Request.GetUserId(), req.ProjectName, req.Description, req.Git)
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

// --- 项目 Git ---

// GetGitStatus 状态聚合 GET /api/teamProject/:id/git
func (controller *TeamProjectController) GetGitStatus(id int) freedom.Result {
	rsp, err := controller.TPSev.GitStatus(controller.Request.GetUserId(), id)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// RetryGitClone 克隆失败后重试（仅创建者） POST /api/teamProject/:id/git/clone-retry
func (controller *TeamProjectController) RetryGitClone(id int) freedom.Result {
	if err := controller.TPSev.RetryGitClone(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// ResolveGitUnrelated 历史无关处理 POST /api/teamProject/:id/git/resolve-unrelated
func (controller *TeamProjectController) ResolveGitUnrelated(id int) freedom.Result {
	var req vo.GitUnrelatedReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	if err := controller.TPSev.ResolveGitUnrelated(controller.Request.GetUserId(), id, &req); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GitCommit 手动提交 POST /api/teamProject/:id/git/commit
func (controller *TeamProjectController) GitCommit(id int) freedom.Result {
	var req vo.GitCommitReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.GitCommit(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GitPush 推送 POST /api/teamProject/:id/git/push
func (controller *TeamProjectController) GitPush(id int) freedom.Result {
	if err := controller.TPSev.GitPush(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GitPull 拉取 POST /api/teamProject/:id/git/pull
func (controller *TeamProjectController) GitPull(id int) freedom.Result {
	if err := controller.TPSev.GitPull(controller.Request.GetUserId(), id); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: map[string]string{"status": "ok"}}
}

// GitLog 提交历史 GET /api/teamProject/:id/git/log
func (controller *TeamProjectController) GitLog(id int) freedom.Result {
	var req vo.GitLogReq
	if err := controller.Request.ReadQuery(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.GitLog(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// GitDiff diff GET /api/teamProject/:id/git/diff
func (controller *TeamProjectController) GitDiff(id int) freedom.Result {
	var req vo.GitDiffReq
	if err := controller.Request.ReadQuery(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.GitDiff(controller.Request.GetUserId(), id, &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

// TestGitRemoteDirect 新建项目前的测试连接 POST /api/teamProject/git/test
func (controller *TeamProjectController) TestGitRemoteDirect() freedom.Result {
	var req vo.GitTestReq
	if err := controller.Request.ReadJSON(&req, false); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	rsp, err := controller.TPSev.TestGitRemoteDirect(&req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	return &infra.JSONResponse{Object: rsp}
}

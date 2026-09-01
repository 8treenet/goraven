package controller

import (
	"errors"
	"os"
	"path/filepath"
	"goraven/backend/infra"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/sandbox"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/hfs", &HFSController{}, infra.NewAuth(true, "/public", "/ak"))
	})
}

// HFSController .
type HFSController struct {
	HFSSev  *service.HFSService
	Worker  freedom.Worker
	Request *infra.Request
}

// BeforeActivation .
func (controller *HFSController) BeforeActivation(b freedom.BeforeActivation) {
	// 下载接口
	b.Handle("GET", "/public/{linkId:string}", "PublicDownload")
	b.Handle("GET", "/private/{p:path}", "PrivateDownload")
	// 绝对路径下载接口（团队共享：任意已登录用户可读任意用户空间内文件）
	b.Handle("GET", "/file/{p:path}", "AbsoluteDownload")
	// 临时访问凭证下载接口（infra.NewAuth 跳过 /ak 前缀鉴权，由 ak + p:path 校验）
	// p:path 为绑定的用户空间相对路径，用于校验与 URL 可读性，不支持子路径浏览
	b.Handle("GET", "/ak/{ak:string}/{p:path}", "AkDownload")
	// 申请临时访问凭证（需要鉴权，避开 /ak 前缀以免被跳过）
	b.Handle("POST", "/access", "CreateTempAccess")
	// 断点续传上传接口
	b.Handle("POST", "/upload/create", "CreateUpload")
	b.Handle("PUT", "/upload/chunk", "UploadChunk")
	b.Handle("POST", "/upload/merge", "MergeUpload")
	// 静态资源
	b.Handle("POST", "/assets", "Assets")
}

// PublicDownload 通过外链ID下载文件
func (controller *HFSController) PublicDownload(linkId string) {
	// URL可能带扩展名（如 abc.pdf），剥离后查询
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

// PrivateDownload 私有下载
// p 为用户空间的相对路径（如 documents/foo/index.html），由 REST 路径参数捕获，统一补回前导 /
func (controller *HFSController) PrivateDownload(p string) {
	//鉴权拦截器里拿到用户id，直接从工作空间取，后续多沙盒需要改取文件
	reqPath := "/" + strings.TrimPrefix(p, "/")
	if reqPath == "/" {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString("path is required")
		return
	}
	sb, sberr := sandbox.NewSandbox(controller.Request.GetUserName())
	if sberr != nil {
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(sberr.Error())
		return
	}
	exists, err := sb.Exists(filepath.Join(sb.GetWorkspace(), reqPath))
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
	absPath, err := sb.Download(filepath.Join(sb.GetWorkspace(), reqPath))
	if err != nil {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString(err.Error())
		return
	}

	infra.ServeFile(controller.Worker.IrisContext(), absPath, filepath.Base(reqPath))
}

// AbsoluteDownload 按文件系统绝对路径下载文件
// 团队共享场景：任意已登录用户可读取任意用户空间内的文件。
// p:path 为完整 FS 绝对路径（形如 /goraven/data/users/<user>/documents/foo.pdf），
// 校验路径必须落在 config.paths.user_space 之下，否则 403。
func (controller *HFSController) AbsoluteDownload(p string) {
	absPath := "/" + strings.TrimPrefix(p, "/")
	if absPath == "/" {
		controller.Worker.IrisContext().StatusCode(400)
		controller.Worker.IrisContext().WriteString("path is required")
		return
	}
	cleanAbs := filepath.Clean(absPath)
	// 兼容相对路径配置：统一转为绝对路径后再比较
	userSpaceRoot := filepath.Clean(config.Get().Paths.UserSpace)
	if absRoot, absErr := filepath.Abs(userSpaceRoot); absErr == nil {
		userSpaceRoot = absRoot
	}
	if !strings.HasPrefix(cleanAbs, userSpaceRoot+string(filepath.Separator)) {
		controller.Worker.IrisContext().StatusCode(403)
		controller.Worker.IrisContext().WriteString("path is outside user space")
		return
	}
	// 符号链接逃逸防护：解析后的真实路径必须仍位于用户空间内
	resolvedAbs, resolveErr := filepath.EvalSymlinks(cleanAbs)
	if resolveErr != nil {
		if os.IsNotExist(resolveErr) {
			controller.Worker.IrisContext().StatusCode(404)
			controller.Worker.IrisContext().WriteString("File not found")
			return
		}
		controller.Worker.IrisContext().StatusCode(500)
		controller.Worker.IrisContext().WriteString(resolveErr.Error())
		return
	}
	resolvedRoot, rootErr := filepath.EvalSymlinks(userSpaceRoot)
	if rootErr != nil {
		resolvedRoot = userSpaceRoot
	}
	if !strings.HasPrefix(resolvedAbs, resolvedRoot+string(filepath.Separator)) && resolvedAbs != resolvedRoot {
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

// CreateTempAccess 申请临时访问凭证
// 用户为自身空间内的文件或目录申请一个 15 分钟有效的临时 ak，供外链访问
func (controller *HFSController) CreateTempAccess() freedom.Result {
	var req vo.TempAccessReq
	if err := controller.Request.ReadJSON(&req, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	resp, err := controller.HFSSev.CreateTempAccess(controller.Request.GetUserId(), controller.Request.GetUserName(), &req)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: resp}
}

// AkDownload 通过临时凭证下载文件
// p:path 为待下载文件在用户空间的相对路径；
// file 类型凭证要求 p:path 与绑定路径完全一致，
// dir 类型凭证要求 p:path 位于绑定目录内
func (controller *HFSController) AkDownload(ak string, p string) {
	ctx := controller.Worker.IrisContext()
	reqPath := "/" + strings.TrimPrefix(p, "/")

	absPath, err := controller.HFSSev.ResolveAkDownload(ak, reqPath)
	if err != nil {
		status := 403
		if errors.Is(err, errs.ErrTempAccessFileNotFound) {
			status = 404
		} else if errors.Is(err, errs.ErrTempAccessNotFile) {
			status = 400
		}
		ctx.StatusCode(status)
		ctx.WriteString(err.Error())
		return
	}

	// ak 路由用于 iframe 预览，sandbox 不带 allow-same-origin 时 iframe origin 为 null，
	// 浏览器对跨源资源要求 CORS 头。ak 本身就是公开凭证（在 URL 路径里），放开 CORS 不降低安全性。
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Cache-Control", "no-cache")
	// 用 ServeFile 而非 SendFile：避免 Content-Disposition: attachment 导致
	// iframe 预览 HTML 时浏览器当下载处理而显示白板。ServeFile 不设 disposition，
	// 浏览器按 Content-Type 内联渲染（HTML/CSS/JS/图片均可，二进制则下载）。
	if err := ctx.ServeFile(absPath, false); err != nil {
		ctx.StatusCode(500)
		ctx.WriteString(err.Error())
	}
}

// CreateUpload 创建分片上传任务
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

// UploadChunk 上传分片
//
// ⚠️ 重要: chunkIndex 必须从 0 开始！
// - chunkIndex 有效范围: [0, totalChunks-1]
// - chunkIndex=0 代表第一个分片，chunkIndex=totalChunks-1 代表最后一个分片
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

// MergeUpload 合并分片
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

// PostAssets 提交上传文件为静态资源
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

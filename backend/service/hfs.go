package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/iface"
	"goraven/core/sandbox"
	"goraven/util"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *HFSService {
			return &HFSService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *HFSService) {
			initiator.FetchService(ctx, &service)
			return
		})
		initiator.BindBooting(func(bootManager freedom.BootManager) {
			freedom.ServiceLocator().Call(func(service *HFSService) error {
				service.Worker.DeferRecycle()
				service.VerifierTimer()
				return nil
			})
		})
	})
}

// HFSService HTTP File Server 业务服务
type HFSService struct {
	Worker         freedom.Worker
	HFSRepo        *repository.HFSRepository
	SysCfgRepo     *repository.SystemSettingRepository
	UserRepository *repository.UserRepository
}

var _ iface.FileURLGenerator = (*HFSService)(nil)

// GenerateURL 实现 iface.FileURLGenerator 接口
// 将用户目录下的文件生成为可外部访问的URL
func (service *HFSService) GenerateURL(userID string, filePath string) (string, error) {
	//暂时不用
	sysconf, err := service.SysCfgRepo.LoadConfig()
	if err != nil {
		return "", err
	}
	user, err := service.UserRepository.FindByUserId(userID)
	if err != nil {
		return "", err
	}
	sb, sberr := sandbox.NewSandbox(user.Username)
	if sberr != nil {
		return "", sberr
	}
	exists, err := sb.Exists(filepath.Join(sb.GetWorkspace(), filePath))
	if err != nil {
		return "", fmt.Errorf("验证文件失败: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	existing, err := service.HFSRepo.GetFileLinkByPath(userID, filePath)
	if err == nil && existing != nil && existing.Deleted == 0 && !existing.IsExpired() {
		return service.buildURL(existing.LinkId, filePath, sysconf.GeneralDomain), nil
	}

	linkId := util.UUID()
	fileName := filepath.Base(filePath)
	expiresHours := sysconf.FileLinkExpiresHours

	link := &po.FileLink{
		LinkId:    linkId,
		UserId:    userID,
		FilePath:  filePath,
		FileName:  fileName,
		ExpiresAt: time.Now().Add(time.Duration(expiresHours) * time.Hour),
	}
	if err := service.HFSRepo.CreateFileLink(link); err != nil {
		return "", fmt.Errorf("创建外链记录失败: %w", err)
	}

	return service.buildURL(linkId, filePath, sysconf.GeneralDomain), nil
}

// ResolveFile 解析外链ID，返回用户ID和文件路径
func (service *HFSService) ResolveFile(linkId string) (userID, userName string, filePath string, fileName string, err error) {
	link, err := service.HFSRepo.GetFileLinkByLinkId(linkId)
	if err != nil {
		return "", "", "", "", fmt.Errorf("外链不存在: %s", linkId)
	}
	if link.Deleted == 1 {
		return "", "", "", "", fmt.Errorf("外链已被删除: %s", linkId)
	}
	if link.IsExpired() {
		return "", "", "", "", fmt.Errorf("外链已过期: %s", linkId)
	}

	user, err := service.UserRepository.FindByUserId(link.UserId)
	if err != nil || user == nil {
		return "", "", "", "", fmt.Errorf("外链已过期: %s", linkId)
	}
	return link.UserId, user.Username, link.FilePath, link.FileName, nil
}

// buildURL 构建文件外链URL
// ext 来自 filepath.Ext，自带前导点（如 .pdf），与 linkId 拼接为 abc123.pdf 作为单一路径段，
// controller 侧解析时通过 TrimSuffix 剥离扩展名再查库
func (service *HFSService) buildURL(linkId string, filePath, domain string) string {
	ext := filepath.Ext(filePath)
	return fmt.Sprintf("%s/api/hfs/public/%s%s", strings.TrimSuffix(domain, "/"), linkId, ext)
}

const tempAkPrefix = "rvnt_"

// CreateTempAccess 创建临时访问凭证
// type 为 "file"（单文件）或 "dir"（目录），凭证 15 分钟内有效
// file 类型绑定到具体文件；dir 类型绑定到目录，授权访问该目录下的任意文件
func (service *HFSService) CreateTempAccess(userID, userName string, req *vo.TempAccessReq) (*vo.TempAccessRsp, error) {
	if req.Type != "file" && req.Type != "dir" {
		return nil, errs.ErrTempAccessTypeInvalid
	}

	sb, sberr := sandbox.NewSandbox(userName)
	if sberr != nil {
		return nil, sberr
	}

	absPath := filepath.Join(sb.GetWorkspace(), req.Path)
	cleanPath := filepath.Clean(absPath)
	workspace := filepath.Clean(sb.GetWorkspace())
	if !strings.HasPrefix(cleanPath, workspace+string(filepath.Separator)) && cleanPath != workspace {
		return nil, errs.ErrTempAccessPathInvalid
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errs.NewFormatError("path not found: %s", "路径不存在: %s", req.Path)
		}
		return nil, fmt.Errorf("stat path: %w", err)
	}
	if req.Type == "file" && info.IsDir() {
		return nil, errs.ErrTempAccessNotFile
	}
	if req.Type == "dir" && !info.IsDir() {
		return nil, errs.ErrTempAccessNotDir
	}

	ak := tempAkPrefix + util.UUID()
	cache := &repository.TempAccessCache{
		UserName: userName,
		Space:    repository.TempSpaceUser,
		Path:     req.Path,
		Type:     req.Type,
	}
	if err := service.HFSRepo.SetTempAccess(ak, cache); err != nil {
		return nil, fmt.Errorf("create temp access: %w", err)
	}

	return &vo.TempAccessRsp{
		Ak:        ak,
		ExpiresAt: time.Now().Add(repository.TempAkTTL).Unix(),
	}, nil
}

// ResolveAkDownload 校验临时凭证并解析出待下载文件的绝对路径
// reqPath 为 URL 中携带的 p:path，与凭证 Path 同属一个路径空间：
// 用户空间凭证为用户工作区相对路径；共享空间凭证形如 /projects/<项目名>/<项目内相对路径>
// file 类型：reqPath 必须与绑定路径完全一致；
// dir 类型：reqPath 必须是绑定目录内的某个文件
func (service *HFSService) ResolveAkDownload(ak, reqPath string) (string, error) {
	cache, err := service.HFSRepo.GetTempAccess(ak)
	if err != nil || cache == nil {
		return "", errs.ErrTempAccessInvalid
	}

	space := cache.Space
	if space == "" {
		space = repository.TempSpaceUser
	}

	if space == repository.TempSpaceShared {
		// 共享空间凭证：文件位于团队项目目录，不在用户沙盒内
		absPath, terr := resolveSharedAkPath(config.Get().GetTeamProjectDir(), cache.Path, cache.Type, reqPath)
		if terr != nil {
			return "", terr
		}
		info, serr := os.Stat(absPath)
		if serr != nil {
			if os.IsNotExist(serr) {
				return "", errs.ErrTempAccessFileNotFound
			}
			return "", serr
		}
		if info.IsDir() {
			return "", errs.ErrTempAccessNotFile
		}
		return absPath, nil
	}

	sb, sberr := sandbox.NewSandbox(cache.UserName)
	if sberr != nil {
		return "", sberr
	}

	cleanReq := filepath.Clean(reqPath)
	cleanBound := filepath.Clean(cache.Path)

	if cache.Type == "file" {
		if cleanReq != cleanBound {
			return "", errs.ErrTempAccessPathNotAllowed
		}
	} else {
		// dir 类型：reqPath 必须严格位于绑定目录内，不能是目录本身
		if cleanReq == cleanBound || !strings.HasPrefix(cleanReq, cleanBound+string(filepath.Separator)) {
			return "", errs.ErrTempAccessPathNotAllowed
		}
	}

	absPath := filepath.Join(sb.GetWorkspace(), cleanReq)
	info, serr := os.Stat(absPath)
	if serr != nil {
		if os.IsNotExist(serr) {
			return "", errs.ErrTempAccessFileNotFound
		}
		return "", serr
	}
	if !info.IsDir() {
		// 通过沙盒校验路径未越出工作空间
		validated, verr := sb.Download(absPath)
		if verr != nil {
			return "", errs.ErrTempAccessPathNotAllowed
		}
		return validated, nil
	}
	return "", errs.ErrTempAccessNotFile
}

// CreateUpload 创建分片上传任务
func (service *HFSService) CreateUpload(userID string, req *vo.ChunkUploadCreateReq) (*vo.ChunkUploadCreateRsp, error) {
	uploadId := util.UUID()
	tempBase := config.Get().GetUploadTempDir()
	tempDir := filepath.Join(tempBase, uploadId)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	totalChunks := req.TotalChunks
	if totalChunks <= 0 {
		totalChunks = int(req.FileSize / int64(req.ChunkSize))
		if req.FileSize%int64(req.ChunkSize) != 0 {
			totalChunks++
		}
	}

	upload := &po.ChunkUpload{
		UploadId:    uploadId,
		UserId:      userID,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		ChunkSize:   req.ChunkSize,
		TotalChunks: totalChunks,
		TempDir:     tempDir,
		Status:      po.UploadStatusPending,
	}

	if err := service.HFSRepo.CreateUpload(upload); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("create upload record: %w", err)
	}

	return &vo.ChunkUploadCreateRsp{
		UploadId: uploadId,
		TempDir:  tempDir,
	}, nil
}

// UploadChunk 上传分片
func (service *HFSService) UploadChunk(userID string, req *vo.ChunkUploadReq, chunkData io.Reader) error {
	upload, err := service.HFSRepo.GetUploadByUploadId(req.UploadId)
	if err != nil {
		return fmt.Errorf("upload not found: %s", req.UploadId)
	}

	if upload.UserId != userID {
		return fmt.Errorf("permission denied")
	}

	if upload.Status != po.UploadStatusPending {
		return fmt.Errorf("upload already completed or cancelled")
	}

	if req.ChunkIndex < 0 || req.ChunkIndex >= upload.TotalChunks {
		return fmt.Errorf("invalid chunk index: %d", req.ChunkIndex)
	}

	chunkPath := filepath.Join(upload.TempDir, fmt.Sprintf("chunk_%d", req.ChunkIndex))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("create chunk file: %w", err)
	}
	defer chunkFile.Close()

	hash := md5.New()
	tee := io.TeeReader(chunkData, hash)

	if _, err := io.Copy(chunkFile, tee); err != nil {
		return fmt.Errorf("write chunk: %w", err)
	}

	actualMd5 := hex.EncodeToString(hash.Sum(nil))
	if req.ChunkMd5 != "" && actualMd5 != req.ChunkMd5 {
		os.Remove(chunkPath)
		return fmt.Errorf("chunk md5 mismatch: expected %s, got %s", req.ChunkMd5, actualMd5)
	}

	return nil
}

// MergeUpload 合并分片
func (service *HFSService) MergeUpload(userID string, uploadId string) (*vo.ChunkMergeRsp, error) {
	upload, err := service.HFSRepo.GetUploadByUploadId(uploadId)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %s", uploadId)
	}

	if upload.UserId != userID {
		return nil, fmt.Errorf("permission denied")
	}

	if upload.Status != po.UploadStatusPending {
		return nil, fmt.Errorf("upload already completed or cancelled")
	}

	for i := 0; i < upload.TotalChunks; i++ {
		chunkPath := filepath.Join(upload.TempDir, fmt.Sprintf("chunk_%d", i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("incomplete upload: missing chunk %d", i)
		}
	}

	mergedPath := filepath.Join(upload.TempDir, upload.FileName)
	mergedFile, err := os.Create(mergedPath)
	if err != nil {
		return nil, fmt.Errorf("create merged file: %w", err)
	}
	defer mergedFile.Close()

	for i := 0; i < upload.TotalChunks; i++ {
		chunkPath := filepath.Join(upload.TempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("open chunk %d: %w", i, err)
		}
		if _, err := io.Copy(mergedFile, chunkFile); err != nil {
			chunkFile.Close()
			return nil, fmt.Errorf("merge chunk %d: %w", i, err)
		}
		chunkFile.Close()
		os.Remove(chunkPath)
	}

	if err := service.HFSRepo.UpdateUploadStatus(uploadId, po.UploadStatusCompleted); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	return &vo.ChunkMergeRsp{
		UploadId: uploadId,
		FilePath: mergedPath,
		FileName: upload.FileName,
		FileSize: upload.FileSize,
	}, nil
}

// CommitAssets 提交上传文件为静态资源
func (service *HFSService) CommitAssets(userID string, uploadId string) (*vo.AssetsRsp, error) {
	upload, err := service.HFSRepo.GetUploadByUploadId(uploadId)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %s", uploadId)
	}

	if upload.UserId != userID {
		return nil, fmt.Errorf("permission denied")
	}

	if upload.Status != po.UploadStatusCompleted {
		return nil, fmt.Errorf("upload not completed")
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found in temp dir")
	}

	dateDir := time.Now().Format("20060102")
	uploadDir := filepath.Join(config.Get().GetUploadDir(), dateDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	ext := filepath.Ext(upload.FileName)
	dstFileName := uploadId + ext
	dstPath := filepath.Join(uploadDir, dstFileName)

	if err := os.Rename(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("move file: %w", err)
	}

	if err := service.HFSRepo.MarkUploadUsed(uploadId); err != nil {
		os.Rename(dstPath, srcPath)
		return nil, fmt.Errorf("mark upload used: %w", err)
	}

	os.RemoveAll(upload.TempDir)

	return &vo.AssetsRsp{
		Path: fmt.Sprintf("/upload/%s/%s", dateDir, dstFileName),
	}, nil
}

// VerifierTimer
func (service *HFSService) VerifierTimer() {
	if !config.Get().System.Initialized {
		return
	}
	go func() {
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Second)
		for {
			service.cleanupExpiredUploads()
			time.Sleep(100 * time.Minute)
		}
	}()
}

func (service *HFSService) cleanupExpiredUploads() {
	if err := service.HFSRepo.SoftDeleteExpiredUploads(2); err != nil {
		service.Worker.Logger().Errorf("cleanup expired uploads db: %v", err)
	}

	removeCall := func(removedir string, cutoff time.Time) {
		entries, err := os.ReadDir(removedir)
		if err != nil {
			service.Worker.Logger().Errorf("cleanup read temp dir: %v", err)
			return
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if !info.ModTime().Before(cutoff) {
				continue
			}
			path := filepath.Join(removedir, entry.Name())
			if entry.IsDir() {
				os.RemoveAll(path)
			} else {
				os.Remove(path)
			}
		}
	}

	removeCall(config.Get().GetUploadTempDir(), time.Now().Add(-48*time.Hour))
	removeCall(config.Get().GetDownloadTempDir(), time.Now().Add(-48*time.Hour))
}

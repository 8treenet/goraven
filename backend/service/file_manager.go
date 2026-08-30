package service

import (
	"fmt"
	"os"
	"path/filepath"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/core/sandbox"
	"goraven/util/envfile"
	"strings"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *FileManagerService {
			return &FileManagerService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *FileManagerService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// FileManagerService 文件管理业务服务
// 通过 sandbox.NewSandbox(userID).NewFileManager() 操作用户空间文件
type FileManagerService struct {
	Worker  freedom.Worker
	HFSRepo *repository.HFSRepository
}

// List 列出指定目录的文件和子目录
func (service *FileManagerService) List(userID string, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	fm, workspace, err := service.newFileManager(userID)
	if err != nil {
		return nil, err
	}

	sortBy := req.Sort
	if sortBy == "" {
		sortBy = "name"
	}
	order := req.Order
	if order == "" {
		order = "asc"
	}

	items, err := fm.List(req.Dir, sortBy, order)
	if err != nil {
		return nil, err
	}

	listItems := make([]vo.FileManagerListItem, 0, len(items))
	isRoot := req.Dir == "" || req.Dir == "/"

	for _, fi := range items {
		// 根目录下隐藏 skills 目录（技能通过数据库管理，禁止文件操作）
		if isRoot && fi.IsDir && fi.Name == "skills" {
			continue
		}
		// 隐藏 . 开头的文件和目录（如 .DS_Store）
		if strings.HasPrefix(fi.Name, ".") {
			continue
		}
		// 根目录下系统初始化创建的条目标记为 isDefault，前端据此显示锁图标并禁用删除
		isDefault := false
		if isRoot {
			if _, ok := protectedPaths[fi.Name]; ok {
				isDefault = true
			}
		}
		listItems = append(listItems, vo.FileManagerListItem{
			Name:      fi.Name,
			Path:      filepath.Join(workspace, req.Dir, fi.Name),
			IsDir:     fi.IsDir,
			Size:      fi.Size,
			ModTime:   fi.ModTime,
			IsDefault: isDefault,
		})
	}

	return &vo.FileManagerListRsp{Items: listItems}, nil
}

// Upload 将 HFS 分片上传合并后的文件移入用户空间指定目录
// 流程：前端先完成 hfs 分片上传（create → chunk → merge），拿到 uploadId，
// 再调用本接口传入 uploadId 和目标目录，后端将临时文件移入用户空间
func (service *FileManagerService) Upload(userID string, req *vo.FileManagerUploadReq) (*vo.FileManagerUploadRsp, error) {
	upload, err := service.HFSRepo.GetUploadByUploadId(req.UploadId)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %s", req.UploadId)
	}
	if upload.UserId != userID {
		return nil, fmt.Errorf("permission denied")
	}
	if upload.Status != po.UploadStatusCompleted {
		return nil, fmt.Errorf("upload not completed")
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("merged file not found in temp dir")
	}

	sb, sberr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sberr != nil {
		return nil, sberr
	}

	// 检查目标路径是否落入 skills 目录
	relPath := filepath.Join(req.Dir, upload.FileName)
	cleaned := filepath.Clean(relPath)
	if cleaned == "skills" || strings.HasPrefix(cleaned, "skills/") {
		return nil, fmt.Errorf("skills directory is managed by system, file operations not allowed")
	}

	dstPath := filepath.Join(sb.GetWorkspace(), relPath)
	if err := sb.Upload(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("failed to move file to user space: %w", err)
	}

	// 移动成功后标记上传任务已使用并清理临时目录
	service.HFSRepo.MarkUploadUsed(req.UploadId)
	os.RemoveAll(upload.TempDir)

	relPath = filepath.Join(req.Dir, upload.FileName)
	return &vo.FileManagerUploadRsp{Path: relPath}, nil
}

// Mkdir 创建目录
func (service *FileManagerService) Mkdir(userID string, req *vo.FileManagerMkdirReq) error {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Mkdir(req.Path)
}

// Rename 重命名文件或目录
// 系统初始化创建的目录和文件（见 protectedPaths）禁止重命名。
func (service *FileManagerService) Rename(userID string, req *vo.FileManagerRenameReq) error {
	cleaned := filepath.Clean(strings.TrimSpace(req.OldPath))
	if _, ok := protectedPaths[cleaned]; ok {
		return fmt.Errorf("%s is a system directory and cannot be renamed", cleaned)
	}
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Rename(req.OldPath, req.NewPath)
}

// protectedPaths 是用户空间内由系统初始化的目录和文件，禁止通过文件管理器删除。
// 与 config.GetUserSpace 中 MkdirAll/WriteFile 的清单保持一致。
var protectedPaths = map[string]struct{}{
	"documents": {},
	"temp":      {},
	"downloads": {},
	"images":    {},
	"videos":    {},
	"projects":  {},
	"skills":    {},
	".profile":  {},
}

// Delete 删除文件或目录
// 系统初始化创建的目录和文件（documents/temp/downloads/images/videos/projects/skills/.profile）禁止删除。
func (service *FileManagerService) Delete(userID string, req *vo.FileManagerDeleteReq) error {
	for _, p := range req.Paths {
		cleaned := filepath.Clean(strings.TrimSpace(p))
		if _, ok := protectedPaths[cleaned]; ok {
			return fmt.Errorf("%s is a system directory and cannot be deleted", cleaned)
		}
	}
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Delete(req.Paths)
}

// Compress 压缩文件或目录为 zip
func (service *FileManagerService) Compress(userID string, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return nil, err
	}
	zipPath, err := fm.Compress(req.Paths, req.OutputName)
	if err != nil {
		return nil, err
	}
	return &vo.FileManagerCompressRsp{ZipPath: zipPath}, nil
}

// Decompress 解压 zip 文件
func (service *FileManagerService) Decompress(userID string, req *vo.FileManagerDecompressReq) error {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Decompress(req.Path, req.ToSubDir)
}

// Usage 获取磁盘使用统计
func (service *FileManagerService) Usage(userID string) (*vo.FileManagerUsageRsp, error) {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return nil, err
	}
	usedSize, fileCount, err := fm.GetUsage()
	if err != nil {
		return nil, err
	}

	sb, sberr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sberr != nil {
		return nil, sberr
	}
	capacity, capErr := sb.GetStorageCapacity()
	if capErr != nil {
		return nil, capErr
	}

	totalSize := capacity.TotalBytes
	if usedSize > 0 && capacity.FreeBytes > usedSize*100 {
		totalSize = 0
	}

	return &vo.FileManagerUsageRsp{
		TotalSize: totalSize,
		UsedSize:  usedSize,
		FileCount: fileCount,
	}, nil
}

// newFileManager 创建用户文件管理器实例，返回文件管理器及用户工作空间的系统绝对路径
func (service *FileManagerService) newFileManager(userID string) (sandbox.FileManager, string, error) {
	sb, err := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if err != nil {
		return nil, "", err
	}
	return sb.NewFileManager(), sb.GetWorkspace(), nil
}

// profileFileName 是用户空间根目录下保存环境变量的 dotenv 文件名。
const profileFileName = ".profile"

// ProfileList 读取 .profile 并返回全部环境变量。
// 文件不存在视为空列表，便于初始化空间直接调用。
func (service *FileManagerService) ProfileList(userID string) (*vo.FileManagerProfileListRsp, error) {
	entries, err := service.readProfile(userID)
	if err != nil {
		return nil, err
	}
	items := make([]vo.FileManagerProfileEntry, 0, len(entries))
	for _, e := range entries {
		k, v := splitEntry(e)
		items = append(items, vo.FileManagerProfileEntry{Key: k, Value: v})
	}
	return &vo.FileManagerProfileListRsp{Items: items}, nil
}

// ProfileCreate 新增一个环境变量；若 key 已存在则返回错误。
// 读取-修改-整体覆盖写入。
func (service *FileManagerService) ProfileCreate(userID string, req *vo.FileManagerProfileCreateReq) error {
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return fmt.Errorf("key is required")
	}
	entries, err := service.readProfile(userID)
	if err != nil {
		return err
	}
	if _, idx := findKey(entries, key); idx >= 0 {
		return fmt.Errorf("env %s already exists", key)
	}
	entries = append(entries, key+"="+req.Value)
	return service.writeProfile(userID, entries)
}

// ProfileUpdate 更新指定 key 的值；不存在则返回错误。
// 顺序保持不变。
func (service *FileManagerService) ProfileUpdate(userID string, req *vo.FileManagerProfileUpdateReq) error {
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return fmt.Errorf("key is required")
	}
	entries, err := service.readProfile(userID)
	if err != nil {
		return err
	}
	_, idx := findKey(entries, key)
	if idx < 0 {
		return fmt.Errorf("env %s not found", key)
	}
	entries[idx] = key + "=" + req.Value
	return service.writeProfile(userID, entries)
}

// ProfileDelete 删除指定 key；不存在则返回错误，便于前端感知操作是否生效。
func (service *FileManagerService) ProfileDelete(userID string, req *vo.FileManagerProfileDeleteReq) error {
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return fmt.Errorf("key is required")
	}
	entries, err := service.readProfile(userID)
	if err != nil {
		return err
	}
	_, idx := findKey(entries, key)
	if idx < 0 {
		return fmt.Errorf("env %s not found", key)
	}
	entries = append(entries[:idx], entries[idx+1:]...)
	return service.writeProfile(userID, entries)
}

// readProfile 通过 sandbox 读取 .profile 并解析为 KEY=VALUE 列表。
// 文件不存在返回空列表（不报错）。
func (service *FileManagerService) readProfile(userID string) ([]string, error) {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return nil, err
	}
	data, err := fm.ReadFile(profileFileName)
	if err != nil {
		// 文件不存在视为空配置
		if strings.Contains(err.Error(), "not found") {
			return []string{}, nil
		}
		return nil, err
	}
	entries, perr := envfile.Parse(data)
	if perr != nil {
		return nil, perr
	}
	return entries, nil
}

// writeProfile 序列化整份 .profile 并通过 sandbox 覆盖写入。
// 任何修改都是全量重写——dotenv 格式不支持原地追加/更新。
func (service *FileManagerService) writeProfile(userID string, entries []string) error {
	data, err := envfile.Serialize(entries)
	if err != nil {
		return err
	}
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.WriteFile(profileFileName, data)
}

// findKey 在 KEY=VALUE 列表里查找首次出现的 key，返回值和索引；不存在时索引为 -1。
func findKey(entries []string, key string) (string, int) {
	prefix := key + "="
	for i, e := range entries {
		if strings.HasPrefix(e, prefix) {
			return e[len(prefix):], i
		}
	}
	return "", -1
}

// splitEntry 拆分 KEY=VALUE；envfile.Parse 已保证 '=' 存在，此处仅做安全兜底。
func splitEntry(e string) (string, string) {
	if i := strings.IndexByte(e, '='); i >= 0 {
		return e[:i], e[i+1:]
	}
	return e, ""
}

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/config"
	"goraven/core/fs"
	"goraven/util/disk"
	"goraven/util/envfile"

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

// FileManagerService 文件管理业务服务。
// 所有文件操作均直接读写文件系统，不再经由 sandbox（sandbox 仅用于 agent shell）。
// 个人项目与团队项目通过嵌入本服务，复用 root 作用域的 root 方法（ListRoot/UploadRoot 等），
// 把项目物理目录作为独立的根目录进行受限操作。
// 文件系统原语（os.Root 句柄、路径安全校验、zip 等）封装在 core/fs 包中。
type FileManagerService struct {
	Worker  freedom.Worker
	HFSRepo *repository.HFSRepository
}

// userWorkspace 返回当前登录用户的用户空间根目录。
// 注意：user_space 目录以 username 命名（config.GetUserSpace(userName)），
// 而 session/表中的主语是 userId，必须经登录会话中的 username 映射，禁止直接用 userId 拼路径。
func (service *FileManagerService) userWorkspace() string {
	return config.Get().GetUserSpace(infra.GetUserName(service.Worker))
}

// ============================ 用户空间（Files 页面）操作 ============================
// 这些接口面向当前登录用户的用户空间根目录（userWorkspace），
// 复用 core/fs.Root 的根目录操作，仅补充用户空间根目录特有的保护规则。

// listRootDir 列出根目录内指定目录的条目并组装为响应（目录恒在文件前，隐藏点号条目）。
// userSpace 表示根目录为用户空间根目录：额外应用两条规则——
// 隐藏系统管理的 skills 目录，并把系统初始化条目标记为 IsDefault（前端据此显示锁图标并禁用删除）。
func listRootDir(rt *fs.Root, dir, sortBy, order string, userSpace bool) (*vo.FileManagerListRsp, error) {
	if dir == "/" {
		dir = ""
	}
	if _, err := rt.Resolve(dir); err != nil {
		return nil, err
	}
	if sortBy == "" {
		sortBy = "name"
	}
	if order == "" {
		order = "asc"
	}

	items, err := rt.List(dir, sortBy, order)
	if err != nil {
		return nil, err
	}

	isRoot := dir == ""
	listItems := make([]vo.FileManagerListItem, 0, len(items))
	for _, fi := range items {
		if strings.HasPrefix(fi.Name, ".") {
			continue
		}
		// 根目录下隐藏 skills 目录（技能通过数据库管理，禁止文件操作）
		if userSpace && isRoot && fi.IsDir && fi.Name == "skills" {
			continue
		}
		isDefault := userSpace && isRoot && protectedRootEntry(fi.Name) != ""
		listItems = append(listItems, vo.FileManagerListItem{
			Name:      fi.Name,
			Path:      filepath.Join(rt.AbsDir(), dir, fi.Name),
			IsDir:     fi.IsDir,
			Size:      fi.Size,
			ModTime:   fi.ModTime,
			IsDefault: isDefault,
		})
	}
	return &vo.FileManagerListRsp{Items: listItems}, nil
}

// List 列出指定目录的文件和子目录
func (service *FileManagerService) List(userID string, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return nil, err
	}
	return listRootDir(rt, req.Dir, req.Sort, req.Order, true)
}

// Upload 将 HFS 分片上传合并后的文件移入用户空间指定目录。
// 流程：前端先完成 hfs 分片上传（create → chunk → merge），拿到 uploadId，
// 再调用本接口传入 uploadId 和目标目录，后端将临时文件移入用户空间。
func (service *FileManagerService) Upload(userID string, req *vo.FileManagerUploadReq) (*vo.FileManagerUploadRsp, error) {
	// 检查目标目录是否落入 skills 目录（controller 已校验，此处再兜底）
	if isSkillsPath(req.Dir) {
		return nil, fmt.Errorf("skills directory is managed by system, file operations not allowed")
	}
	return service.UploadRoot(service.userWorkspace(), userID, req)
}

// Mkdir 创建目录
func (service *FileManagerService) Mkdir(userID string, req *vo.FileManagerMkdirReq) error {
	return service.MkdirRoot(service.userWorkspace(), req)
}

// Rename 重命名文件或目录。
// 系统初始化创建的目录和文件（见 protectedPaths）禁止重命名。
func (service *FileManagerService) Rename(userID string, req *vo.FileManagerRenameReq) error {
	if entry := protectedRootEntry(req.OldPath); entry != "" {
		return fmt.Errorf("%s is a system directory and cannot be renamed", entry)
	}
	return service.RenameRoot(service.userWorkspace(), req)
}

// Delete 删除文件或目录。
// 系统初始化创建的目录和文件（见 protectedPaths）禁止删除。
func (service *FileManagerService) Delete(userID string, req *vo.FileManagerDeleteReq) error {
	for _, p := range req.Paths {
		if entry := protectedRootEntry(p); entry != "" {
			return fmt.Errorf("%s is a system directory and cannot be deleted", entry)
		}
	}
	return service.DeleteRoot(service.userWorkspace(), req)
}

// Compress 压缩文件或目录为 zip，zip 生成在第一个源路径所在目录。
func (service *FileManagerService) Compress(userID string, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return nil, err
	}
	zipPath, err := rt.Zip(req.Paths, req.OutputName)
	if err != nil {
		return nil, err
	}
	return &vo.FileManagerCompressRsp{ZipPath: zipPath}, nil
}

// Decompress 解压 zip 文件
func (service *FileManagerService) Decompress(userID string, req *vo.FileManagerDecompressReq) error {
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return err
	}
	return rt.Unzip(req.Path, req.ToSubDir)
}

// Usage 获取磁盘使用统计
func (service *FileManagerService) Usage(userID string) (*vo.FileManagerUsageRsp, error) {
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return nil, err
	}
	usedSize, fileCount, err := rt.Usage()
	if err != nil {
		return nil, err
	}

	totalSize, freeBytes, err := workspaceCapacity(rt.AbsDir())
	if err != nil {
		return nil, err
	}
	if usedSize > 0 && freeBytes > usedSize*100 {
		totalSize = 0
	}

	return &vo.FileManagerUsageRsp{
		TotalSize: totalSize,
		UsedSize:  usedSize,
		FileCount: fileCount,
	}, nil
}

// workspaceCapacity 查找用户空间所在磁盘分区的总容量与剩余空间。
// workspace 必须已归一化为绝对路径（来自 fs.Root.AbsDir）。
func workspaceCapacity(workspace string) (totalBytes, freeBytes int64, err error) {
	var best disk.Info
	for _, m := range disk.GetMountPoints() {
		if strings.HasPrefix(workspace, m.MountPoint) {
			if best.MountPoint == "" || len(m.MountPoint) > len(best.MountPoint) {
				best = m
			}
		}
	}
	if best.MountPoint == "" {
		return 0, 0, fmt.Errorf("no mount point found for workspace: %s", workspace)
	}
	return best.TotalBytes, best.FreeBytes, nil
}

// ============================ 项目根目录（个人/团队项目）操作 ============================
// 以下 root 方法以传入的 root 目录为独立根进行操作，由 MyProjectService/TeamProjectService
// 通过嵌入 FileManagerService 复用。个人项目与团队项目的文件操作均走这里。

// ListRoot 列出指定根目录下的文件和子目录。
func (service *FileManagerService) ListRoot(root string, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	rt, err := fs.Open(root)
	if err != nil {
		return nil, err
	}
	return listRootDir(rt, req.Dir, req.Sort, req.Order, false)
}

// UploadRoot 将 HFS 分片上传合并后的文件移入指定根目录。
func (service *FileManagerService) UploadRoot(root, userID string, req *vo.FileManagerUploadReq) (*vo.FileManagerUploadRsp, error) {
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
	if err := validateUploadFileName(upload.FileName); err != nil {
		return nil, err
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("merged file not found in temp dir")
	}

	rt, err := fs.Open(root)
	if err != nil {
		return nil, err
	}
	if _, err := rt.Resolve(req.Dir); err != nil {
		return nil, err
	}
	relPath := filepath.Join(req.Dir, upload.FileName)
	if _, err := rt.Resolve(relPath); err != nil {
		return nil, err
	}

	// 通过 os.Root 稳定句柄创建目标目录并写入目标文件，
	// 避免先校验路径后 os.Rename 的竞态。
	parentRel, err := rt.Clean(filepath.Dir(relPath))
	if err != nil {
		return nil, err
	}
	if err := rt.MkdirAll(parentRel); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := rt.CopyFileInto(srcPath, relPath); err != nil {
		return nil, fmt.Errorf("failed to move file to root: %w", err)
	}

	// 移动成功后标记上传任务已使用并清理临时目录；
	// 任一失败都返回错误，避免在落地状态不完整时报告上传成功。
	if err := service.HFSRepo.MarkUploadUsed(req.UploadId); err != nil {
		return nil, fmt.Errorf("failed to mark upload as used: %w", err)
	}
	if err := os.RemoveAll(upload.TempDir); err != nil {
		return nil, fmt.Errorf("failed to clean up temp dir: %w", err)
	}

	return &vo.FileManagerUploadRsp{Path: relPath}, nil
}

// MkdirRoot 在指定根目录下创建目录。
func (service *FileManagerService) MkdirRoot(root string, req *vo.FileManagerMkdirReq) error {
	rt, err := fs.Open(root)
	if err != nil {
		return err
	}
	if _, err := rt.Resolve(req.Path); err != nil {
		return err
	}
	return rt.MkdirAll(req.Path)
}

// RenameRoot 在指定根目录下重命名文件或目录。
// 根目录本身不可作为重命名的源或目标。
func (service *FileManagerService) RenameRoot(root string, req *vo.FileManagerRenameReq) error {
	if fs.IsRootEquivalentPath(req.OldPath) {
		return fmt.Errorf("refusing to rename the root directory: %s", req.OldPath)
	}
	if fs.IsRootEquivalentPath(req.NewPath) {
		return fmt.Errorf("refusing to rename to the root directory: %s", req.NewPath)
	}
	rt, err := fs.Open(root)
	if err != nil {
		return err
	}
	if _, err := rt.Resolve(req.OldPath); err != nil {
		return err
	}
	if _, err := rt.Resolve(req.NewPath); err != nil {
		return err
	}
	return rt.Rename(req.OldPath, req.NewPath)
}

// DeleteRoot 删除指定根目录下的文件或目录。
// 根目录本身（空路径、"foo/.." 等等价形式）禁止删除，避免整个根被清空。
func (service *FileManagerService) DeleteRoot(root string, req *vo.FileManagerDeleteReq) error {
	for _, path := range req.Paths {
		if fs.IsRootEquivalentPath(path) {
			return fmt.Errorf("refusing to delete the root directory: %s", path)
		}
	}
	rt, err := fs.Open(root)
	if err != nil {
		return err
	}
	for _, path := range req.Paths {
		if _, err := rt.Resolve(path); err != nil {
			return err
		}
	}
	return rt.RemoveAll(req.Paths)
}

// CompressRoot 将指定根目录下的文件或目录压缩到根目录下。
func (service *FileManagerService) CompressRoot(root string, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	rt, err := fs.Open(root)
	if err != nil {
		return nil, err
	}
	for _, path := range req.Paths {
		if _, err := rt.Resolve(path); err != nil {
			return nil, err
		}
	}

	outputName := req.OutputName
	if outputName == "" {
		outputName = "archive.zip"
	}
	if !strings.HasSuffix(outputName, ".zip") {
		outputName += ".zip"
	}
	if _, err := rt.Resolve(outputName); err != nil {
		return nil, err
	}

	managerOutputName := outputName
	if len(req.Paths) > 0 {
		// 前端路径为根相对写法（如 "/foo"），filepath.Dir 会返回绝对路径 "/"，
		// 与相对的 outputName 混用会导致 filepath.Rel 报错；
		// 两侧先归一化为根相对路径，再计算 zip 输出相对源目录的位置。
		srcDir, err := rt.Clean(filepath.Dir(req.Paths[0]))
		if err != nil {
			return nil, err
		}
		outRel, err := rt.Clean(outputName)
		if err != nil {
			return nil, err
		}
		managerOutputName, err = filepath.Rel(srcDir, outRel)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve archive output: %w", err)
		}
	}
	if _, err := rt.Zip(req.Paths, managerOutputName); err != nil {
		return nil, err
	}
	return &vo.FileManagerCompressRsp{ZipPath: outputName}, nil
}

// DecompressRoot 解压指定根目录下的 zip 文件。
func (service *FileManagerService) DecompressRoot(root string, req *vo.FileManagerDecompressReq) error {
	rt, err := fs.Open(root)
	if err != nil {
		return err
	}
	if _, err := rt.Resolve(req.Path); err != nil {
		return err
	}
	return rt.Unzip(req.Path, req.ToSubDir)
}

// UsageRoot 获取指定根目录的磁盘使用统计。
func (service *FileManagerService) UsageRoot(root string) (*vo.FileManagerUsageRsp, error) {
	rt, err := fs.Open(root)
	if err != nil {
		return nil, err
	}
	usedSize, fileCount, err := rt.Usage()
	if err != nil {
		return nil, err
	}
	return &vo.FileManagerUsageRsp{UsedSize: usedSize, FileCount: fileCount}, nil
}

// DownloadRoot 校验并返回指定根目录下的普通文件绝对路径。
func (service *FileManagerService) DownloadRoot(root, subPath string) (string, string, error) {
	rt, err := fs.Open(root)
	if err != nil {
		return "", "", err
	}
	absPath, err := rt.Resolve(subPath)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("file not found")
		}
		return "", "", err
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("path is not a regular file")
	}
	return absPath, filepath.Base(subPath), nil
}

// ============================ 用户空间保护规则 ============================

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

// protectedRootEntry 若 path 指向系统初始化创建的用户空间根目录条目（顶层精确匹配），
// 返回清理后的条目名；否则返回空串。
func protectedRootEntry(path string) string {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if _, ok := protectedPaths[cleaned]; ok {
		return cleaned
	}
	return ""
}

// isSkillsPath 判断相对路径是否为根目录下 skills 目录自身或其子路径。
// skills 目录由系统管理（技能经数据库安装），用户空间文件操作禁止触碰。
func isSkillsPath(path string) bool {
	cleaned := filepath.Clean(path)
	return cleaned == "skills" || strings.HasPrefix(cleaned, "skills"+string(filepath.Separator))
}

// validateUploadFileName 校验上传文件名是安全的单级文件名（不含路径分隔符、绝对路径等）。
func validateUploadFileName(name string) error {
	if name == "" || name == "." || name == ".." || filepath.IsAbs(name) || filepath.VolumeName(name) != "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) || strings.ContainsRune(name, 0) {
		return fmt.Errorf("unsafe upload filename: %s", name)
	}
	return nil
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

// readProfile 直接读取用户空间根目录下的 .profile 并解析为 KEY=VALUE 列表。
// 文件不存在返回空列表（不报错）。
func (service *FileManagerService) readProfile(userID string) ([]string, error) {
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return nil, err
	}
	data, err := rt.ReadFile(profileFileName)
	if err != nil {
		// 文件不存在视为空配置
		if os.IsNotExist(err) {
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

// writeProfile 序列化整份 .profile 并覆盖写入。
// 任何修改都是全量重写——dotenv 格式不支持原地追加/更新。
func (service *FileManagerService) writeProfile(userID string, entries []string) error {
	data, err := envfile.Serialize(entries)
	if err != nil {
		return err
	}
	rt, err := fs.Open(service.userWorkspace())
	if err != nil {
		return err
	}
	return rt.WriteFile(profileFileName, data)
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

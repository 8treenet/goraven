package service

import (
	"fmt"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/core/sandbox"
	"goraven/util/envfile"
	"os"
	"path/filepath"
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

type FileManagerService struct {
	Worker    freedom.Worker
	HFSRepo   *repository.HFSRepository
	ShareRepo *repository.SharedProjectRepository
}

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

	var sharedSet map[string]int
	if strings.TrimPrefix(strings.TrimSpace(req.Dir), "/") == "projects" {
		sharedProjects, err := service.ShareRepo.ListByOwner(userID)
		if err == nil && len(sharedProjects) > 0 {
			sharedSet = make(map[string]int, len(sharedProjects))
			for _, sp := range sharedProjects {
				sharedSet[sp.ProjectName] = sp.Id
			}
		}
	}

	for _, fi := range items {

		if isRoot && fi.IsDir && fi.Name == "skills" {
			continue
		}

		if strings.HasPrefix(fi.Name, ".") {
			continue
		}

		isDefault := false
		if isRoot {
			if _, ok := protectedPaths[fi.Name]; ok {
				isDefault = true
			}
		}

		isShared := false
		sharedId := 0
		if sharedSet != nil && fi.IsDir {
			if id, ok := sharedSet[fi.Name]; ok {
				isShared = true
				sharedId = id
			}
		}
		listItems = append(listItems, vo.FileManagerListItem{
			Name:      fi.Name,
			Path:      filepath.Join(workspace, req.Dir, fi.Name),
			IsDir:     fi.IsDir,
			Size:      fi.Size,
			ModTime:   fi.ModTime,
			IsDefault: isDefault,
			IsShared:  isShared,
			SharedId:  sharedId,
		})
	}

	return &vo.FileManagerListRsp{Items: listItems}, nil
}

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

	relPath := filepath.Join(req.Dir, upload.FileName)
	cleaned := filepath.Clean(relPath)
	if cleaned == "skills" || strings.HasPrefix(cleaned, "skills/") {
		return nil, fmt.Errorf("skills directory is managed by system, file operations not allowed")
	}

	dstPath := filepath.Join(sb.GetWorkspace(), relPath)
	if err := sb.Upload(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("failed to move file to user space: %w", err)
	}

	service.HFSRepo.MarkUploadUsed(req.UploadId)
	os.RemoveAll(upload.TempDir)

	relPath = filepath.Join(req.Dir, upload.FileName)
	return &vo.FileManagerUploadRsp{Path: relPath}, nil
}

func (service *FileManagerService) Mkdir(userID string, req *vo.FileManagerMkdirReq) error {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Mkdir(req.Path)
}

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

func (service *FileManagerService) Decompress(userID string, req *vo.FileManagerDecompressReq) error {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return err
	}
	return fm.Decompress(req.Path, req.ToSubDir)
}

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

func (service *FileManagerService) newFileManager(userID string) (sandbox.FileManager, string, error) {
	sb, err := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if err != nil {
		return nil, "", err
	}
	return sb.NewFileManager(), sb.GetWorkspace(), nil
}

const profileFileName = ".profile"

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

func (service *FileManagerService) readProfile(userID string) ([]string, error) {
	fm, _, err := service.newFileManager(userID)
	if err != nil {
		return nil, err
	}
	data, err := fm.ReadFile(profileFileName)
	if err != nil {

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

func findKey(entries []string, key string) (string, int) {
	prefix := key + "="
	for i, e := range entries {
		if strings.HasPrefix(e, prefix) {
			return e[len(prefix):], i
		}
	}
	return "", -1
}

func splitEntry(e string) (string, string) {
	if i := strings.IndexByte(e, '='); i >= 0 {
		return e[:i], e[i+1:]
	}
	return e, ""
}

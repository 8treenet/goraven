package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/core/sandbox"
	"goraven/util"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *TeamProjectService {
			return &TeamProjectService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *TeamProjectService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type TeamProjectService struct {
	Worker    freedom.Worker
	ShareRepo *repository.SharedProjectRepository
	HFSRepo   *repository.HFSRepository
}

func (service *TeamProjectService) List(userId string) (*vo.TeamProjectListRsp, error) {
	projects, err := service.ShareRepo.ListAll()
	if err != nil {
		return nil, err
	}

	ownerIds := make([]string, 0, len(projects))
	seen := map[string]bool{}
	for _, p := range projects {
		if !seen[p.OwnerId] {
			seen[p.OwnerId] = true
			ownerIds = append(ownerIds, p.OwnerId)
		}
	}
	userMap, err := service.ShareRepo.GetUsersByIDs(ownerIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.TeamProjectItem, 0, len(projects))
	for _, p := range projects {
		item := vo.TeamProjectItem{
			Id:          p.Id,
			OwnerId:     p.OwnerId,
			ProjectName: p.ProjectName,
			Description: p.Description,
			UpdatedAt:   p.Updated,
			IsOwner:     p.OwnerId == userId,
		}
		if u, ok := userMap[p.OwnerId]; ok {
			item.OwnerName = u.Nickname
			if item.OwnerName == "" {
				item.OwnerName = u.Username
			}
			item.OwnerAvatar = u.Avatar
		}
		items = append(items, item)
	}
	return &vo.TeamProjectListRsp{Items: items}, nil
}

func (service *TeamProjectService) Get(userId string, id int) (*vo.TeamProjectItem, error) {
	project, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return nil, errs.ErrSharedProjectNotFound
	}
	item := vo.TeamProjectItem{
		Id:          project.Id,
		OwnerId:     project.OwnerId,
		ProjectName: project.ProjectName,
		Description: project.Description,
		UpdatedAt:   project.Updated,
		IsOwner:     project.OwnerId == userId,
	}
	if u, err := service.ShareRepo.GetUserByID(project.OwnerId); err == nil {
		item.OwnerName = u.Nickname
		if item.OwnerName == "" {
			item.OwnerName = u.Username
		}
		item.OwnerAvatar = u.Avatar
	}
	return &item, nil
}

func (service *TeamProjectService) Share(userId, projectName, description string) (*vo.TeamProjectShareRsp, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" || strings.Contains(projectName, "/") || strings.Contains(projectName, "\\") {
		return nil, errs.ErrSharedProjectInvalidName
	}

	user, err := service.ShareRepo.GetUserByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	sb, err := sandbox.NewSandbox(user.Username)
	if err != nil {
		return nil, err
	}
	if !dirExistsInSandbox(sb, filepath.Join("projects", projectName)) {
		return nil, errs.ErrSharedProjectDirNotFound
	}

	existing, err := service.ShareRepo.GetByOwnerAndProject(userId, projectName)
	if err == nil && existing != nil {
		return nil, errs.ErrSharedProjectAlreadyShared
	}

	record := &po.SharedProject{
		OwnerId:     userId,
		ProjectName: projectName,
		Description: description,
	}
	if err := service.ShareRepo.Create(record); err != nil {
		return nil, err
	}
	return &vo.TeamProjectShareRsp{SharedId: record.Id}, nil
}

func (service *TeamProjectService) Unshare(userId string, id int) error {
	project, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return errs.ErrSharedProjectNotFound
	}
	if project.OwnerId != userId {
		return errs.ErrSharedProjectPermission
	}
	return service.ShareRepo.Delete(id)
}

func (service *TeamProjectService) UpdateDescription(userId string, id int, description string) error {
	project, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return errs.ErrSharedProjectNotFound
	}
	if project.OwnerId != userId {
		return errs.ErrSharedProjectPermission
	}
	return service.ShareRepo.UpdateDescription(id, description)
}

func (service *TeamProjectService) AdminListSharedProjects(req *vo.AdminSharedProjectListReq) (*infra.PageResponse, error) {
	projects, pr, err := service.ShareRepo.PaginateSharedProjects(req)
	if err != nil {
		return nil, err
	}

	ownerIds := make([]string, 0, len(projects))
	seen := map[string]bool{}
	for _, p := range projects {
		if !seen[p.OwnerId] {
			seen[p.OwnerId] = true
			ownerIds = append(ownerIds, p.OwnerId)
		}
	}
	userMap, err := service.ShareRepo.GetUsersByIDs(ownerIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminSharedProjectItem, 0, len(projects))
	for _, p := range projects {
		item := vo.AdminSharedProjectItem{
			Id:          p.Id,
			OwnerId:     p.OwnerId,
			ProjectName: p.ProjectName,
			Description: p.Description,
			VisitCount:  p.VisitCount,
			Created:     p.Created,
			Updated:     p.Updated,
		}
		if !p.LastActiveAt.IsZero() {
			item.LastActiveAt = &p.LastActiveAt
		}
		if u, ok := userMap[p.OwnerId]; ok {
			item.OwnerName = u.Nickname
			if item.OwnerName == "" {
				item.OwnerName = u.Username
			}
			item.OwnerAvatar = u.Avatar
		}
		locked, sessionId, _ := service.ShareRepo.IsSharedProjectLocked(p.Id)
		item.Locked = locked
		item.LockedBy = sessionId
		items = append(items, item)
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *TeamProjectService) AdminUnshare(id int) error {
	_, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return errs.ErrSharedProjectNotFound
	}
	return service.ShareRepo.Delete(id)
}

func (service *TeamProjectService) newSandboxForProject(id int) (sandbox.Sandbox, sandbox.FileManager, string, *po.SharedProject, error) {
	project, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return nil, nil, "", nil, errs.ErrSharedProjectNotFound
	}

	user, err := service.ShareRepo.GetUserByID(project.OwnerId)
	if err != nil {
		return nil, nil, "", nil, fmt.Errorf("failed to get owner info: %w", err)
	}

	sb, err := sandbox.NewSandbox(user.Username)
	if err != nil {
		return nil, nil, "", nil, err
	}

	projectRoot := filepath.Join("projects", project.ProjectName)
	if !dirExistsInSandbox(sb, projectRoot) {
		return nil, nil, "", nil, errs.ErrSharedProjectDirNotFound
	}

	fm := sb.NewFileManager()
	return sb, fm, projectRoot, project, nil
}

func joinProjectPath(projectRoot, subPath string) string {
	subPath = strings.TrimSpace(subPath)
	if subPath == "" || subPath == "/" {
		return projectRoot
	}
	return filepath.Join(projectRoot, subPath)
}

func dirExistsInSandbox(sb sandbox.Sandbox, relPath string) bool {
	absPath := filepath.Join(sb.GetWorkspace(), relPath)
	info, err := os.Stat(absPath)
	return err == nil && info.IsDir()
}

func (service *TeamProjectService) ListFiles(id int, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	sb, fm, projectRoot, _, err := service.newSandboxForProject(id)
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

	dir := joinProjectPath(projectRoot, req.Dir)
	items, err := fm.List(dir, sortBy, order)
	if err != nil {
		return nil, err
	}

	listItems := make([]vo.FileManagerListItem, 0, len(items))
	for _, fi := range items {
		if strings.HasPrefix(fi.Name, ".") {
			continue
		}
		listItems = append(listItems, vo.FileManagerListItem{
			Name:    fi.Name,
			Path:    filepath.Join(sb.GetWorkspace(), projectRoot, req.Dir, fi.Name),
			IsDir:   fi.IsDir,
			Size:    fi.Size,
			ModTime: fi.ModTime,
		})
	}
	return &vo.FileManagerListRsp{Items: listItems}, nil
}

func (service *TeamProjectService) Upload(id int, userId string, req *vo.FileManagerUploadReq) (*vo.FileManagerUploadRsp, error) {
	upload, err := service.HFSRepo.GetUploadByUploadId(req.UploadId)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %s", req.UploadId)
	}
	if upload.UserId != userId {
		return nil, fmt.Errorf("permission denied")
	}
	if upload.Status != po.UploadStatusCompleted {
		return nil, fmt.Errorf("upload not completed")
	}

	sb, _, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return nil, err
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("merged file not found in temp dir")
	}

	relPath := filepath.Join(projectRoot, req.Dir, upload.FileName)
	dstPath := filepath.Join(sb.GetWorkspace(), relPath)
	if err := sb.Upload(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("failed to move file to project: %w", err)
	}

	service.HFSRepo.MarkUploadUsed(req.UploadId)
	os.RemoveAll(upload.TempDir)

	returnPath := filepath.Join(req.Dir, upload.FileName)
	return &vo.FileManagerUploadRsp{Path: returnPath}, nil
}

func (service *TeamProjectService) Mkdir(id int, req *vo.FileManagerMkdirReq) error {
	_, fm, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return err
	}
	return fm.Mkdir(joinProjectPath(projectRoot, req.Path))
}

func (service *TeamProjectService) Rename(id int, req *vo.FileManagerRenameReq) error {
	_, fm, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return err
	}
	oldPath := joinProjectPath(projectRoot, req.OldPath)
	newPath := joinProjectPath(projectRoot, req.NewPath)
	return fm.Rename(oldPath, newPath)
}

func (service *TeamProjectService) Delete(id int, req *vo.FileManagerDeleteReq) error {
	_, fm, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return err
	}
	paths := make([]string, len(req.Paths))
	for i, p := range req.Paths {
		paths[i] = joinProjectPath(projectRoot, p)
	}
	return fm.Delete(paths)
}

func (service *TeamProjectService) Compress(id int, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	_, fm, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(req.Paths))
	for i, p := range req.Paths {
		paths[i] = joinProjectPath(projectRoot, p)
	}
	zipPath, err := fm.Compress(paths, req.OutputName)
	if err != nil {
		return nil, err
	}
	relPath := strings.TrimPrefix(zipPath, projectRoot+"/")
	return &vo.FileManagerCompressRsp{ZipPath: relPath}, nil
}

func (service *TeamProjectService) Decompress(id int, req *vo.FileManagerDecompressReq) error {
	_, fm, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return err
	}
	return fm.Decompress(joinProjectPath(projectRoot, req.Path), req.ToSubDir)
}

func (service *TeamProjectService) Usage(id int) (*vo.FileManagerUsageRsp, error) {
	sb, _, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return nil, err
	}
	absProjectDir := filepath.Join(sb.GetWorkspace(), projectRoot)
	var usedSize int64
	var fileCount int
	err = filepath.Walk(absProjectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			usedSize += info.Size()
			fileCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &vo.FileManagerUsageRsp{
		UsedSize:  usedSize,
		FileCount: fileCount,
	}, nil
}

func (service *TeamProjectService) CreateTempAccess(id int, req *vo.TempAccessReq) (*vo.TempAccessRsp, error) {
	if req.Type != "file" && req.Type != "dir" {
		return nil, errs.ErrTempAccessTypeInvalid
	}

	project, err := service.ShareRepo.GetByID(id)
	if err != nil {
		return nil, errs.ErrSharedProjectNotFound
	}
	user, err := service.ShareRepo.GetUserByID(project.OwnerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner info: %w", err)
	}

	projectRoot := filepath.Join("projects", project.ProjectName)
	fullPath := "/" + joinProjectPath(projectRoot, req.Path)

	sb, sberr := sandbox.NewSandbox(user.Username)
	if sberr != nil {
		return nil, sberr
	}

	absPath := filepath.Join(sb.GetWorkspace(), fullPath)
	cleanPath := filepath.Clean(absPath)
	workspace := filepath.Clean(sb.GetWorkspace())
	if !strings.HasPrefix(cleanPath, workspace+string(filepath.Separator)) && cleanPath != workspace {
		return nil, errs.ErrTempAccessPathInvalid
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errs.NewFormatError("path not found: %s", "路径不存在: %s", fullPath)
		}
		return nil, fmt.Errorf("stat path: %w", err)
	}
	if req.Type == "file" && info.IsDir() {
		return nil, errs.ErrTempAccessNotFile
	}
	if req.Type == "dir" && !info.IsDir() {
		return nil, errs.ErrTempAccessNotDir
	}

	ak := "rvnt_" + util.UUID()
	cache := &repository.TempAccessCache{
		UserName: user.Username,
		Path:     fullPath,
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

func (service *TeamProjectService) Download(id int, subPath string) (string, string, error) {
	sb, _, projectRoot, _, err := service.newSandboxForProject(id)
	if err != nil {
		return "", "", err
	}
	fullRelPath := joinProjectPath(projectRoot, subPath)
	absPath := filepath.Join(sb.GetWorkspace(), fullRelPath)
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("file not found")
		}
		return "", "", err
	}
	if info.IsDir() {
		return "", "", fmt.Errorf("path is a directory, not a file")
	}
	validated, err := sb.Download(absPath)
	if err != nil {
		return "", "", err
	}
	return validated, filepath.Base(fullRelPath), nil
}

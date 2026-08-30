package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
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

// TeamProjectService 团队项目服务
type TeamProjectService struct {
	Worker  freedom.Worker
	TPRepo  *repository.TeamProjectRepository
	HFSRepo *repository.HFSRepository
}

var teamProjectNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// --- 项目管理 ---

// List 列出当前用户可见的团队项目（创建者 OR 成员）
func (service *TeamProjectService) List(userId string) (*vo.TeamProjectListRsp, error) {
	projects, err := service.TPRepo.ListByUser(userId)
	if err != nil {
		return nil, err
	}

	creatorIds := make([]string, 0, len(projects))
	seen := map[string]bool{}
	for _, p := range projects {
		if !seen[p.CreatorId] {
			seen[p.CreatorId] = true
			creatorIds = append(creatorIds, p.CreatorId)
		}
	}
	userMap, err := service.TPRepo.GetUsersByIDs(creatorIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.TeamProjectItem, 0, len(projects))
	for _, p := range projects {
		item := vo.TeamProjectItem{
			Id:          p.Id,
			CreatorId:   p.CreatorId,
			ProjectName: p.ProjectName,
			Description: p.Description,
			Access:      p.Access,
			UpdatedAt:   p.Updated,
			IsCreator:   p.CreatorId == userId,
		}
		if u, ok := userMap[p.CreatorId]; ok {
			item.CreatorName = u.Nickname
			if item.CreatorName == "" {
				item.CreatorName = u.Username
			}
			item.CreatorAvatar = u.Avatar
		}
		items = append(items, item)
	}
	return &vo.TeamProjectListRsp{Items: items}, nil
}

// Get 查询单个团队项目详情
func (service *TeamProjectService) Get(userId string, id int) (*vo.TeamProjectItem, error) {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return nil, errs.ErrTeamProjectNotFound
	}
	item := vo.TeamProjectItem{
		Id:          project.Id,
		CreatorId:   project.CreatorId,
		ProjectName: project.ProjectName,
		Description: project.Description,
		Access:      project.Access,
		UpdatedAt:   project.Updated,
		IsCreator:   project.CreatorId == userId,
	}
	if u, err := service.TPRepo.GetUserByID(project.CreatorId); err == nil {
		item.CreatorName = u.Nickname
		if item.CreatorName == "" {
			item.CreatorName = u.Username
		}
		item.CreatorAvatar = u.Avatar
	}
	return &item, nil
}

// Create 创建团队项目（建目录 + 写DB）
func (service *TeamProjectService) Create(userId, projectName, description string) (*vo.TeamProjectCreateRsp, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" || !teamProjectNameRegexp.MatchString(projectName) {
		return nil, errs.ErrTeamProjectInvalidName
	}

	existing, err := service.TPRepo.GetByName(projectName)
	if err == nil && existing != nil {
		return nil, errs.ErrTeamProjectAlreadyExists
	}

	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	record := &po.TeamProject{
		CreatorId:   userId,
		ProjectName: projectName,
		Description: description,
	}
	if err := service.TPRepo.Create(record); err != nil {
		os.RemoveAll(projectDir)
		return nil, err
	}
	return &vo.TeamProjectCreateRsp{Id: record.Id}, nil
}

// DeleteProject 删除团队项目（仅创建者，删目录 + 删DB + 删成员）
func (service *TeamProjectService) DeleteProject(userId string, id int) error {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	if project.CreatorId != userId {
		return errs.ErrTeamProjectPermission
	}
	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	os.RemoveAll(projectDir)
	service.TPRepo.RemoveMembersByProjectId(id)
	return service.TPRepo.Delete(id)
}

// UpdateDescription 更新简介（仅创建者）
func (service *TeamProjectService) UpdateDescription(userId string, id int, description string) error {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	if project.CreatorId != userId {
		return errs.ErrTeamProjectPermission
	}
	return service.TPRepo.UpdateDescription(id, description)
}

// UpdateAccess 设置访问权限（仅创建者）
func (service *TeamProjectService) UpdateAccess(userId string, id int, access uint8) error {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	if project.CreatorId != userId {
		return errs.ErrTeamProjectPermission
	}
	return service.TPRepo.UpdateAccess(id, access)
}

// AdminListTeamProjects 管理端列出所有团队项目（分页+搜索，含访问统计、锁状态）
func (service *TeamProjectService) AdminListTeamProjects(req *vo.AdminTeamProjectListReq) (*infra.PageResponse, error) {
	projects, pr, err := service.TPRepo.Paginate(req)
	if err != nil {
		return nil, err
	}

	creatorIds := make([]string, 0, len(projects))
	seen := map[string]bool{}
	for _, p := range projects {
		if !seen[p.CreatorId] {
			seen[p.CreatorId] = true
			creatorIds = append(creatorIds, p.CreatorId)
		}
	}
	userMap, err := service.TPRepo.GetUsersByIDs(creatorIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminTeamProjectItem, 0, len(projects))
	for _, p := range projects {
		item := vo.AdminTeamProjectItem{
			Id:          p.Id,
			CreatorId:   p.CreatorId,
			ProjectName: p.ProjectName,
			Description: p.Description,
			Access:      p.Access,
			VisitCount:  p.VisitCount,
			Created:     p.Created,
			Updated:     p.Updated,
		}
		if p.LastActiveTime != nil {
			item.LastActiveAt = p.LastActiveTime
		}
		if u, ok := userMap[p.CreatorId]; ok {
			item.CreatorName = u.Nickname
			if item.CreatorName == "" {
				item.CreatorName = u.Username
			}
			item.CreatorAvatar = u.Avatar
		}
		locked, sessionId, _ := service.TPRepo.IsTeamProjectLocked(p.Id)
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

// AdminDeleteProject 管理端删除团队项目（无需创建者校验）
func (service *TeamProjectService) AdminDeleteProject(id int) error {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	projectDir := filepath.Join(config.Get().GetTeamProjectDir(), project.ProjectName)
	os.RemoveAll(projectDir)
	service.TPRepo.RemoveMembersByProjectId(id)
	return service.TPRepo.Delete(id)
}

// AdminUpdateDescription 管理端更新团队项目简介（无需创建者校验）
func (service *TeamProjectService) AdminUpdateDescription(id int, description string) error {
	_, err := service.TPRepo.GetByID(id)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	return service.TPRepo.UpdateDescription(id, description)
}

// --- 文件操作 ---

// getProjectDir 获取团队项目的物理目录绝对路径
func getProjectDir(projectName string) string {
	return filepath.Join(config.Get().GetTeamProjectDir(), projectName)
}

// resolveSharedAkPath 校验共享空间（团队项目）临时凭证的请求路径，解析出待下载文件的绝对路径
// reqPath 为 URL 中携带的路径，boundPath 为签发凭证时绑定的路径，二者均为 ak 空间路径，
// 形如 /projects/<项目名>/<项目内相对路径>。
// file 类型：reqPath 必须与绑定路径完全一致；
// dir 类型：reqPath 必须严格位于绑定目录内，不能是目录本身
func resolveSharedAkPath(teamRoot, boundPath, typ, reqPath string) (string, error) {
	cleanReq := filepath.Clean(reqPath)
	cleanBound := filepath.Clean(boundPath)
	if typ == "file" {
		if cleanReq != cleanBound {
			return "", errs.ErrTempAccessPathNotAllowed
		}
	} else if cleanReq == cleanBound || !strings.HasPrefix(cleanReq, cleanBound+string(filepath.Separator)) {
		return "", errs.ErrTempAccessPathNotAllowed
	}

	// 从 ak 空间路径解析项目名与项目内相对路径：/projects/<项目名>/<相对路径>
	// 路径比较已保证 cleanReq 必然位于签发时绑定的 /projects/<项目名>/ 之下
	name, sub, ok := strings.Cut(strings.TrimPrefix(cleanReq, "/projects/"), "/")
	if !ok || name == "" || name == "." || name == ".." {
		return "", errs.ErrTempAccessPathNotAllowed
	}

	projectDir := filepath.Join(teamRoot, name)
	absPath := filepath.Join(projectDir, sub)
	if !strings.HasPrefix(filepath.Clean(absPath), filepath.Clean(projectDir)+string(filepath.Separator)) {
		return "", errs.ErrTempAccessPathNotAllowed
	}
	return absPath, nil
}

// validateProject 校验团队项目存在且目录有效，返回项目记录和物理目录
func (service *TeamProjectService) validateProject(id int) (*po.TeamProject, string, error) {
	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return nil, "", errs.ErrTeamProjectNotFound
	}
	projectDir := getProjectDir(project.ProjectName)
	info, statErr := os.Stat(projectDir)
	if statErr != nil || !info.IsDir() {
		return nil, "", errs.ErrTeamProjectDirNotFound
	}
	return project, projectDir, nil
}

// ListFiles 列出项目内指定目录的文件
func (service *TeamProjectService) ListFiles(id int, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return nil, err
	}

	dir := req.Dir
	if dir == "" || dir == "/" {
		dir = "."
	}
	absDir := filepath.Join(projectDir, dir)

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	listItems := make([]vo.FileManagerListItem, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, _ := entry.Info()
		var size int64
		var modTime time.Time
		if info != nil {
			size = info.Size()
			modTime = info.ModTime()
		}
		listItems = append(listItems, vo.FileManagerListItem{
			Name:    entry.Name(),
			Path:    filepath.Join(absDir, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    size,
			ModTime: modTime,
		})
	}
	return &vo.FileManagerListRsp{Items: listItems}, nil
}

// Upload 将 HFS 分片上传合并后的文件移入团队项目目录
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

	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return nil, err
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("merged file not found in temp dir")
	}

	dstDir := filepath.Join(projectDir, req.Dir)
	os.MkdirAll(dstDir, 0755)
	dstPath := filepath.Join(dstDir, upload.FileName)
	if err := moveFileCrossDevice(srcPath, dstPath, os.Rename); err != nil {
		return nil, fmt.Errorf("failed to move file to project: %w", err)
	}

	service.HFSRepo.MarkUploadUsed(req.UploadId)
	os.RemoveAll(upload.TempDir)

	returnPath := filepath.Join(req.Dir, upload.FileName)
	return &vo.FileManagerUploadRsp{Path: returnPath}, nil
}

// moveFileCrossDevice moves a file with a copy fallback for paths on different mounts.
func moveFileCrossDevice(srcPath, dstPath string, rename func(string, string) error) error {
	if err := rename(srcPath, dstPath); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Remove(srcPath)
}

// Mkdir 在项目内新建目录
func (service *TeamProjectService) Mkdir(id int, req *vo.FileManagerMkdirReq) error {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(projectDir, req.Path), 0755)
}

// Rename 重命名项目内文件
func (service *TeamProjectService) Rename(id int, req *vo.FileManagerRenameReq) error {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return err
	}
	return os.Rename(filepath.Join(projectDir, req.OldPath), filepath.Join(projectDir, req.NewPath))
}

// Delete 删除项目内文件
func (service *TeamProjectService) Delete(id int, req *vo.FileManagerDeleteReq) error {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return err
	}
	cleanBase := filepath.Clean(projectDir)
	for _, p := range req.Paths {
		absPath := filepath.Join(projectDir, p)
		if !strings.HasPrefix(filepath.Clean(absPath), cleanBase+string(filepath.Separator)) {
			continue
		}
		os.RemoveAll(absPath)
	}
	return nil
}

// Compress 压缩项目内文件
func (service *TeamProjectService) Compress(id int, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return nil, err
	}
	absPaths := make([]string, len(req.Paths))
	for i, p := range req.Paths {
		absPaths[i] = filepath.Join(projectDir, p)
	}
	outputName := req.OutputName
	if outputName == "" {
		outputName = "archive.zip"
	}
	if !strings.HasSuffix(outputName, ".zip") {
		outputName += ".zip"
	}
	zipPath := filepath.Join(projectDir, outputName)
	if err := util.CreateZip(absPaths, zipPath, projectDir); err != nil {
		return nil, err
	}
	return &vo.FileManagerCompressRsp{ZipPath: outputName}, nil
}

// Decompress 解压项目内 zip 文件
func (service *TeamProjectService) Decompress(id int, req *vo.FileManagerDecompressReq) error {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return err
	}
	zipPath := filepath.Join(projectDir, req.Path)
	destDir := projectDir
	if req.ToSubDir {
		base := strings.TrimSuffix(filepath.Base(req.Path), filepath.Ext(req.Path))
		destDir = filepath.Join(projectDir, base)
	}
	return util.ExtractZip(zipPath, destDir)
}

// Usage 项目磁盘使用统计（仅统计项目目录）
func (service *TeamProjectService) Usage(id int) (*vo.FileManagerUsageRsp, error) {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return nil, err
	}
	var usedSize int64
	var fileCount int
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
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

// CreateTempAccess 为团队项目内文件/目录创建临时访问凭证
func (service *TeamProjectService) CreateTempAccess(id int, req *vo.TempAccessReq) (*vo.TempAccessRsp, error) {
	if req.Type != "file" && req.Type != "dir" {
		return nil, errs.ErrTempAccessTypeInvalid
	}

	project, err := service.TPRepo.GetByID(id)
	if err != nil {
		return nil, errs.ErrTeamProjectNotFound
	}

	projectDir := getProjectDir(project.ProjectName)
	absPath := filepath.Join(projectDir, req.Path)
	cleanPath := filepath.Clean(absPath)
	cleanProjectDir := filepath.Clean(projectDir)
	if !strings.HasPrefix(cleanPath, cleanProjectDir+string(filepath.Separator)) && cleanPath != cleanProjectDir {
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

	user, err := service.TPRepo.GetUserByID(project.CreatorId)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator info: %w", err)
	}

	ak := "rvnt_" + util.UUID()
	// Path 统一存 ak 空间路径 /projects/<项目名>/<项目内相对路径>，与前端 buildAkPath 及 ak 下载 URL 一致
	cache := &repository.TempAccessCache{
		UserName: user.Username,
		Space:    repository.TempSpaceShared,
		Path:     filepath.Join("/projects", project.ProjectName, filepath.Clean("/"+strings.TrimPrefix(req.Path, "/"))),
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

// Download 解析团队项目内文件的绝对路径，供 controller SendFile
func (service *TeamProjectService) Download(id int, subPath string) (string, string, error) {
	_, projectDir, err := service.validateProject(id)
	if err != nil {
		return "", "", err
	}
	absPath := filepath.Join(projectDir, subPath)
	if !strings.HasPrefix(filepath.Clean(absPath), filepath.Clean(projectDir)+string(filepath.Separator)) {
		return "", "", fmt.Errorf("invalid path")
	}
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
	return absPath, filepath.Base(subPath), nil
}

// --- 成员管理 ---

// ListUsers 分页查询所有可用用户（成员选择器用）
func (service *TeamProjectService) ListUsers(req *vo.TeamProjectUserListReq) (*infra.PageResponse, error) {
	items, pr, err := service.TPRepo.PaginateActiveUsers(req)
	if err != nil {
		return nil, err
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// ListMembers 查询项目成员列表
func (service *TeamProjectService) ListMembers(userId string, projectId int) (*vo.TeamProjectMembersRsp, error) {
	project, err := service.TPRepo.GetByID(projectId)
	if err != nil {
		return nil, errs.ErrTeamProjectNotFound
	}

	members, err := service.TPRepo.ListMembers(projectId)
	if err != nil {
		return nil, err
	}

	memberIds := make([]string, 0, len(members))
	for _, m := range members {
		memberIds = append(memberIds, m.UserId)
	}
	return &vo.TeamProjectMembersRsp{
		CreatorId: project.CreatorId,
		MemberIds: memberIds,
	}, nil
}

// UpdateMembers 编辑项目成员（仅创建者操作）
func (service *TeamProjectService) UpdateMembers(userId string, projectId int, req *vo.TeamProjectMemberUpdateReq) error {
	project, err := service.TPRepo.GetByID(projectId)
	if err != nil {
		return errs.ErrTeamProjectNotFound
	}
	if project.CreatorId != userId {
		return errs.ErrTeamProjectPermission
	}

	// 添加成员
	for _, uid := range req.AddUserIds {
		if uid == project.CreatorId {
			continue // 创建者不进成员表
		}
		isMember, _ := service.TPRepo.IsMember(projectId, uid)
		if isMember {
			continue
		}
		if err := service.TPRepo.AddMember(projectId, uid); err != nil {
			return err
		}
	}

	// 移除成员
	for _, uid := range req.RemoveUserIds {
		if uid == project.CreatorId {
			return errs.ErrTeamProjectCannotRemoveCreator
		}
		if err := service.TPRepo.RemoveMember(projectId, uid); err != nil {
			return err
		}
	}
	return nil
}

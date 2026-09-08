package service

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/fs"
	"goraven/util"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *MyProjectService {
			return &MyProjectService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *MyProjectService) {
			initiator.FetchService(ctx, &service)
			return
		})
		initiator.BindBooting(func(bootManager freedom.BootManager) {
			freedom.ServiceLocator().Call(func(service *MyProjectService) error {
				service.Worker.DeferRecycle()
				service.Rebuild()
				return nil
			})
		})
	})
}

// MyProjectService 个人项目服务
// 物理目录位于用户空间 projects/{ProjectName}，元数据由 user_project 表管理。
// 文件操作复用 FileManagerService 的 root 作用域方法（嵌入继承）。
type MyProjectService struct {
	FileManagerService
	Worker   freedom.Worker
	Repo     *repository.UserProjectRepository
	UserRepo *repository.UserRepository
}

// validateProjectName 校验项目名是否符合 Linux 目录命名规则（项目名即物理目录名）。
func validateProjectName(projectName string) error {
	return fs.ValidateDirName(projectName)
}

// userSpaceDir 解析用户的物理空间根目录。
// 注意：user_space 目录以 username 命名（config.GetUserSpace(userName)），
// 而 session/表中的主语是 userId，必须经 username 映射，禁止直接把 userId 拼进路径。
func (service *MyProjectService) userSpaceDir(userId string) (string, error) {
	username, err := service.Repo.GetUsernameByUserID(userId)
	if err != nil {
		return "", err
	}
	return config.Get().GetUserSpace(username), nil
}

// projectDir 计算个人项目物理目录
func (service *MyProjectService) projectDir(userSpace, projectName string) string {
	return filepath.Join(userSpace, "projects", projectName)
}

// projectDirUpdatedAt 返回项目目录的最后修改时间，stat 失败时回退表更新时间。
func projectDirUpdatedAt(dir string, fallback time.Time) time.Time {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fallback
	}
	return info.ModTime()
}

// validateProject 校验项目存在、属于当前用户且目录有效，返回记录与物理目录
func (service *MyProjectService) validateProject(userId string, id int) (*po.UserProject, string, error) {
	project, err := service.Repo.GetByID(id)
	if err != nil {
		return nil, "", errs.ErrUserProjectNotFound
	}
	if project.UserId != userId {
		return nil, "", errs.ErrUserProjectNotFound // 不暴露他人项目存在性
	}
	userSpace, err := service.userSpaceDir(userId)
	if err != nil {
		return nil, "", err
	}
	projectDir := service.projectDir(userSpace, project.ProjectName)
	if info, statErr := os.Stat(projectDir); statErr != nil || !info.IsDir() {
		return nil, "", errs.ErrUserProjectDirNotFound
	}
	return project, projectDir, nil
}

// Rebuild 启动引导钩子：延迟 10 秒后把磁盘上的 projects 目录补录进 user_project 表。
func (service *MyProjectService) Rebuild() {
	go func() {
		time.Sleep(10 * time.Second)
		service.syncUserProjectsToDB()
	}()
}

// syncUserProjectsToDB 遍历 user_space 下每个用户目录的 projects/，
// 为磁盘上存在但表中没有记录的目录补建元数据（只补录，不做反向清理）。
func (service *MyProjectService) syncUserProjectsToDB() {
	defer func() {
		if r := recover(); r != nil {
			freedom.Logger().Errorf("syncUserProjectsToDB panic: %v", r)
		}
	}()

	userSpaceRoot := config.Get().Paths.UserSpace
	userDirs, err := os.ReadDir(userSpaceRoot)
	if err != nil {
		freedom.Logger().Errorf("syncUserProjectsToDB read user_space %s err: %v", userSpaceRoot, err)
		return
	}
	for _, userDir := range userDirs {
		if !userDir.IsDir() || strings.HasPrefix(userDir.Name(), ".") {
			continue
		}
		service.syncUserProjects(filepath.Join(userSpaceRoot, userDir.Name()), userDir.Name())
	}
}

// syncUserProjects 补录单个用户 projects/ 目录下缺失的表记录。
func (service *MyProjectService) syncUserProjects(userSpace, username string) {
	user, err := service.UserRepo.FindByUsername(username)
	if err != nil {
		// 目录存在但无有效用户（软删除/残留目录）时不建记录
		freedom.Logger().Infof("syncUserProjects skip user dir %s: %v", username, err)
		return
	}

	entries, err := os.ReadDir(filepath.Join(userSpace, "projects"))
	if err != nil {
		return // 该用户尚未创建 projects 目录
	}

	existing, err := service.Repo.ListByUser(user.UserId)
	if err != nil {
		freedom.Logger().Errorf("syncUserProjects list user %s projects err: %v", username, err)
		return
	}
	projectNames := make(map[string]struct{}, len(existing))
	for _, project := range existing {
		projectNames[project.ProjectName] = struct{}{}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectName := entry.Name()
		if err := validateProjectName(projectName); err != nil {
			freedom.Logger().Warnf("syncUserProjects skip invalid project dir %s/%s: %v", username, projectName, err)
			continue
		}
		if _, ok := projectNames[projectName]; ok {
			continue
		}
		if err := service.Repo.Create(&po.UserProject{UserId: user.UserId, ProjectName: projectName}); err != nil {
			freedom.Logger().Errorf("syncUserProjects create %s/%s err: %v", username, projectName, err)
			continue
		}
		freedom.Logger().Infof("syncUserProjects created record for %s/%s", username, projectName)
	}
}

// --- 项目管理 ---

// List 列出当前用户的个人项目
func (service *MyProjectService) List(userId string) (*vo.MyProjectListRsp, error) {
	projects, err := service.Repo.ListByUser(userId)
	if err != nil {
		return nil, err
	}
	userSpace, err := service.userSpaceDir(userId)
	if err != nil {
		return nil, err
	}
	items := make([]vo.MyProjectItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, vo.MyProjectItem{
			Id:          p.Id,
			ProjectName: p.ProjectName,
			Description: p.Description,
			GitUrl:      p.GitUrl,
			UpdatedAt:   projectDirUpdatedAt(service.projectDir(userSpace, p.ProjectName), p.Updated),
			Created:     p.Created,
		})
	}
	return &vo.MyProjectListRsp{Items: items}, nil
}

// Get 查询单个个人项目
func (service *MyProjectService) Get(userId string, id int) (*vo.MyProjectItem, error) {
	project, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}
	return &vo.MyProjectItem{
		Id:          project.Id,
		ProjectName: project.ProjectName,
		Description: project.Description,
		GitUrl:      project.GitUrl,
		UpdatedAt:   projectDirUpdatedAt(projectDir, project.Updated),
		Created:     project.Created,
	}, nil
}

// Create 创建个人项目（建目录 + 写表）
func (service *MyProjectService) Create(userId, projectName, description string) (*vo.MyProjectCreateRsp, error) {
	projectName = strings.TrimSpace(projectName)
	if err := validateProjectName(projectName); err != nil {
		return nil, errs.ErrUserProjectInvalidName
	}
	if _, err := service.Repo.GetByName(userId, projectName); err == nil {
		return nil, errs.ErrUserProjectAlreadyExists
	}

	userSpace, err := service.userSpaceDir(userId)
	if err != nil {
		return nil, err
	}
	projectDir := service.projectDir(userSpace, projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, err
	}

	record := &po.UserProject{
		UserId:      userId,
		ProjectName: projectName,
		Description: description,
	}
	if err := service.Repo.Create(record); err != nil {
		os.RemoveAll(projectDir)
		return nil, err
	}
	return &vo.MyProjectCreateRsp{Id: record.Id}, nil
}

// Update 更新个人项目：改简介和/或改名（改名同步目录与 session/automation 引用）
func (service *MyProjectService) Update(userId string, id int, req *vo.MyProjectUpdateReq) error {
	project, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}

	newName := strings.TrimSpace(req.ProjectName)
	rename := newName != "" && newName != project.ProjectName
	if rename {
		if err := validateProjectName(newName); err != nil {
			return errs.ErrUserProjectInvalidName
		}
		if _, err := service.Repo.GetByName(userId, newName); err == nil {
			return errs.ErrUserProjectAlreadyExists
		}
	}

	// 简介仅在未要求改名或明确传入非空简介时更新
	if req.Description != "" || !rename {
		if err := service.Repo.UpdateDescription(id, req.Description); err != nil {
			return err
		}
	}

	if rename {
		userSpace, err := service.userSpaceDir(userId)
		if err != nil {
			return err
		}
		newDir := service.projectDir(userSpace, newName)
		if err := os.Rename(projectDir, newDir); err != nil {
			return err
		}
		if err := service.Repo.RenameProject(project, newName); err != nil {
			// 回滚目录名
			os.Rename(newDir, projectDir)
			return err
		}
	}
	return nil
}

// DeleteProject 删除个人项目（删目录 + 清引用 + 删表）
func (service *MyProjectService) DeleteProject(userId string, id int) error {
	project, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}
	os.RemoveAll(projectDir)
	if err := service.Repo.ClearProjectReferences(project.UserId, project.ProjectName); err != nil {
		return err
	}
	return service.Repo.Delete(id)
}

// --- 文件操作（转发共享层）---

// ListFiles 列出项目内文件
func (service *MyProjectService) ListFiles(userId string, id int, req *vo.FileManagerListReq) (*vo.FileManagerListRsp, error) {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}
	return service.ListRoot(projectDir, req)
}

// Upload 上传文件到项目内
func (service *MyProjectService) Upload(userId string, id int, req *vo.FileManagerUploadReq) (*vo.FileManagerUploadRsp, error) {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}
	return service.UploadRoot(projectDir, userId, req)
}

// Mkdir 项目内新建目录
func (service *MyProjectService) Mkdir(userId string, id int, req *vo.FileManagerMkdirReq) error {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}
	return service.MkdirRoot(projectDir, req)
}

// Rename 重命名项目内文件
func (service *MyProjectService) Rename(userId string, id int, req *vo.FileManagerRenameReq) error {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}
	return service.RenameRoot(projectDir, req)
}

// Delete 删除项目内文件
func (service *MyProjectService) Delete(userId string, id int, req *vo.FileManagerDeleteReq) error {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}
	return service.DeleteRoot(projectDir, req)
}

// Compress 压缩项目内文件
func (service *MyProjectService) Compress(userId string, id int, req *vo.FileManagerCompressReq) (*vo.FileManagerCompressRsp, error) {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}
	return service.CompressRoot(projectDir, req)
}

// Decompress 解压项目内 zip
func (service *MyProjectService) Decompress(userId string, id int, req *vo.FileManagerDecompressReq) error {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return err
	}
	return service.DecompressRoot(projectDir, req)
}

// Usage 项目磁盘使用统计
func (service *MyProjectService) Usage(userId string, id int) (*vo.FileManagerUsageRsp, error) {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}
	return service.UsageRoot(projectDir)
}

// Download 解析项目内文件绝对路径，供 controller SendFile
func (service *MyProjectService) Download(userId string, id int, subPath string) (string, string, error) {
	_, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return "", "", err
	}
	return service.DownloadRoot(projectDir, subPath)
}

// CreateTempAccess 为项目内文件/目录创建临时下载凭证。
// 个人项目位于用户空间内，直接复用 TempSpaceUser 凭证，
// Path 为 ak 空间路径 /projects/{项目名}/{相对路径}，与前端 buildAkPath 及 /api/hfs/ak 解析逻辑一致。
func (service *MyProjectService) CreateTempAccess(userId, userName string, id int, req *vo.TempAccessReq) (*vo.TempAccessRsp, error) {
	if req.Type != "file" && req.Type != "dir" {
		return nil, errs.ErrTempAccessTypeInvalid
	}
	project, projectDir, err := service.validateProject(userId, id)
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}
	if req.Type == "file" && info.IsDir() {
		return nil, errs.ErrTempAccessNotFile
	}
	if req.Type == "dir" && !info.IsDir() {
		return nil, errs.ErrTempAccessNotDir
	}

	ak := tempAkPrefix + util.UUID()
	// Path 与团队项目凭证一致存 ak 空间路径 /projects/<项目名>/<相对路径>，
	// 与前端 buildAkPath + controller 归一化后的请求路径（带前导 /）匹配
	cache := &repository.TempAccessCache{
		UserName: userName,
		Space:    repository.TempSpaceUser,
		Path:     filepath.Join("/projects", project.ProjectName, filepath.Clean("/"+strings.TrimPrefix(req.Path, "/"))),
		Type:     req.Type,
	}
	if err := service.HFSRepo.SetTempAccess(ak, cache); err != nil {
		return nil, err
	}
	return &vo.TempAccessRsp{Ak: ak, ExpiresAt: time.Now().Add(repository.TempAkTTL).Unix()}, nil
}

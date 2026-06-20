package service

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/backend/vo/errs"
	"raven/config"
	"raven/core/iface"
	"raven/core/sandbox"
	"raven/util"
	"strings"

	"github.com/8treenet/freedom"
	"github.com/cloudwego/eino/adk/filesystem"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *SkillService {
			return &SkillService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *SkillService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type SkillService struct {
	Worker            freedom.Worker
	SkillRepo         *repository.SkillRepository
	HFSRepo           *repository.HFSRepository
	ClawHubRepo       *repository.ClawHubRepository
	SystemSettingRepo *repository.SystemSettingRepository
	AIModelRepo       *repository.ProviderRepository
	DailyStatsRepo    *repository.DailyStatsRepository
	PersonaRepo       *repository.PersonaRepository
	UserRepo          *repository.UserRepository
}

func (service *SkillService) BeginRequest(worker freedom.Worker) {
	settingConf, err := service.SystemSettingRepo.LoadConfig()
	if err != nil {
		return
	}
	service.ClawHubRepo.ClawHubAPIURL = settingConf.ClawHubAPIURL
	service.ClawHubRepo.ClawHubToken = settingConf.ClawHubToken
}

func (service *SkillService) ListSystemSkills(req *vo.AdminSystemSkillListReq) (*infra.PageResponse, error) {
	skills, pr, err := service.SkillRepo.PaginateSystemSkills(req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminSystemSkillItem, 0, len(skills))
	for _, s := range skills {
		items = append(items, vo.AdminSystemSkillItem{
			SkillId:     s.SkillId,
			Name:        s.Name,
			Description: s.Description,
			Status:      s.Status,
			Updated:     s.Updated,
		})
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *SkillService) GetSystemSkillDetail(skillId int) (*vo.AdminSystemSkillDetailRsp, error) {
	skill, err := service.SkillRepo.GetSystemSkillByID(skillId)
	if err != nil {
		return nil, errs.ErrSystemSkillNotFound
	}
	return &vo.AdminSystemSkillDetailRsp{
		SkillId:     skill.SkillId,
		Name:        skill.Name,
		Description: skill.Description,
		Content:     skill.Content,
		Status:      skill.Status,
		Created:     skill.Created,
		Updated:     skill.Updated,
	}, nil
}

func (service *SkillService) CreateSystemSkill(req *vo.AdminCreateSystemSkillReq) error {
	if req.Content == "" {
		return errs.ErrSystemSkillContentRequired
	}

	name, description, err := util.ParseSkillFrontmatter(req.Content)
	if err != nil {
		return service.translateParseError(err)
	}
	if !strings.HasPrefix(name, "raven-") {
		return errors.New("skill name must start with 'raven-'")
	}

	if _, err := service.SkillRepo.FindSystemSkillByName(name); err == nil {
		return errs.ErrSystemSkillNameAlreadyExists
	}

	skill := &po.SystemSkill{
		Name:        name,
		Description: description,
		Content:     req.Content,
		Status:      po.SystemSkillStatusEnabled,
	}
	return service.SkillRepo.CreateSystemSkill(skill)
}

func (service *SkillService) UpdateSystemSkill(skillId int, req *vo.AdminUpdateSystemSkillReq) error {
	skill, err := service.SkillRepo.GetSystemSkillByID(skillId)
	if err != nil {
		return errs.ErrSystemSkillNotFound
	}
	fmt.Println(skill)

	if req.Content == "" {
		return nil
	}

	name, description, err := util.ParseSkillFrontmatter(req.Content)
	if err != nil {
		return service.translateParseError(err)
	}
	if !strings.HasPrefix(name, "raven-") {
		return errors.New("skill name must start with 'raven-'")
	}

	existing, err := service.SkillRepo.FindSystemSkillByName(name)
	if err == nil && existing.SkillId != skillId {
		return errs.ErrSystemSkillNameAlreadyExists
	}

	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"content":     req.Content,
	}
	return service.SkillRepo.UpdateSystemSkill(skillId, updates)
}

func (service *SkillService) UpdateSystemSkillStatus(skillId int, status uint8) error {
	_, err := service.SkillRepo.GetSystemSkillByID(skillId)
	if err != nil {
		return errs.ErrSystemSkillNotFound
	}
	return service.SkillRepo.UpdateSystemSkill(skillId, map[string]interface{}{"status": int(status)})
}

func (service *SkillService) DeleteSystemSkill(skillId int) error {
	_, err := service.SkillRepo.GetSystemSkillByID(skillId)
	if err != nil {
		return errs.ErrSystemSkillNotFound
	}
	return service.SkillRepo.SoftDeleteSystemSkill(skillId)
}

func (service *SkillService) ListMarketSkills(req *vo.AdminMarketSkillListReq) (*infra.PageResponse, error) {
	skills, pr, err := service.SkillRepo.PaginateMarketSkills(req)
	if err != nil {
		return nil, err
	}

	categoryMap := service.batchGetCategoryMap(skills)

	items := make([]vo.AdminMarketSkillItem, 0, len(skills))
	for _, s := range skills {
		item := vo.AdminMarketSkillItem{
			SkillId:        s.SkillId,
			Name:           s.Name,
			Description:    s.Description,
			Icon:           s.Icon,
			Source:         s.Source,
			CategoryId:     s.CategoryId,
			InstalledCount: s.InstalledCount,
			Status:         s.Status,
			SortOrder:      s.SortOrder,
			Updated:        s.Updated,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
			item.CategoryIcon = cat.Icon
		}
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

func (service *SkillService) GetMarketSkillDetail(skillId int) (*vo.AdminMarketSkillDetailRsp, error) {
	skill, err := service.SkillRepo.GetMarketSkillByID(skillId)
	if err != nil {
		return nil, errs.ErrMarketSkillNotFound
	}

	rsp := &vo.AdminMarketSkillDetailRsp{
		SkillId:        skill.SkillId,
		Name:           skill.Name,
		Description:    skill.Description,
		Icon:           skill.Icon,
		Source:         skill.Source,
		SourceUrl:      skill.SourceUrl,
		CategoryId:     skill.CategoryId,
		Status:         skill.Status,
		SortOrder:      skill.SortOrder,
		InstalledCount: skill.InstalledCount,
		Remark:         skill.Remark,
		Created:        skill.Created,
		Updated:        skill.Updated,
	}

	if skill.CategoryId > 0 {
		if cat, err := service.SkillRepo.GetSkillCategoryByID(skill.CategoryId); err == nil {
			rsp.CategoryName = cat.Name
			rsp.CategoryIcon = cat.Icon
		}
	}

	skillMdPath := filepath.Join(config.Get().Paths.SkillsHub, skill.Name, "SKILL.md")
	if data, err := os.ReadFile(skillMdPath); err == nil {
		rsp.Content = string(data)
	}

	return rsp, nil
}

func (service *SkillService) UpdateMarketSkill(skillId int, req *vo.AdminUpdateMarketSkillReq) error {
	if _, err := service.SkillRepo.GetMarketSkillByID(skillId); err != nil {
		return errs.ErrMarketSkillNotFound
	}

	updates := map[string]interface{}{}

	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.CategoryId != nil {
		if _, err := service.SkillRepo.GetSkillCategoryByID(*req.CategoryId); err != nil {
			return errs.ErrSkillCategoryNotFound
		}
		updates["category_id"] = *req.CategoryId
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}

	if len(updates) == 0 {
		return nil
	}

	return service.SkillRepo.UpdateMarketSkill(skillId, updates)
}

func (service *SkillService) UpdateMarketSkillStatus(skillId int, status uint8) error {
	if _, err := service.SkillRepo.GetMarketSkillByID(skillId); err != nil {
		return errs.ErrMarketSkillNotFound
	}
	return service.SkillRepo.UpdateMarketSkill(skillId, map[string]interface{}{"status": int(status)})
}

func (service *SkillService) DeleteMarketSkill(skillId int, cascade bool) error {
	skill, err := service.SkillRepo.GetMarketSkillByID(skillId)
	if err != nil {
		return errs.ErrMarketSkillNotFound
	}

	if cascade {

		userSkillIds, _ := service.SkillRepo.ListUserSkillIdsByMarketSkillId(skillId)

		if err := service.SkillRepo.DeleteUserSkillsBySkillId(skillId); err != nil {
			return errs.NewFormatError("delete user skills failed: %v", "删除用户技能记录失败: %v", err)
		}

		_ = service.PersonaRepo.DeletePersonaToolsBySkillIds(userSkillIds)
	}

	skillDir := filepath.Join(config.Get().Paths.SkillsHub, skill.Name)
	os.RemoveAll(skillDir)

	return service.SkillRepo.SoftDeleteMarketSkill(skillId)
}

func (service *SkillService) GetMarketSkillUsers(skillId int, req *vo.AdminMarketSkillUserListReq) (*infra.PageResponse, error) {
	if _, err := service.SkillRepo.GetMarketSkillByID(skillId); err != nil {
		return nil, errs.ErrMarketSkillNotFound
	}

	userSkills, pr, err := service.SkillRepo.PaginateMarketSkillUsers(skillId, &repository.PageQuery{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminMarketSkillUserItem, 0, len(userSkills))
	for _, us := range userSkills {
		items = append(items, vo.AdminMarketSkillUserItem{
			UserId:        us.UserId,
			InstallStatus: us.InstallStatus,
			Created:       us.Created,
		})
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *SkillService) PublishMarketSkill(req *vo.AdminPublishMarketSkillReq) error {
	upload, err := service.HFSRepo.GetUploadByUploadId(req.UploadId)
	if err != nil {
		return errs.ErrSkillUploadNotFound
	}

	if upload.Status != po.UploadStatusCompleted {
		return errs.ErrSkillUploadNotCompleted
	}

	zipPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return errs.ErrSkillUploadNotFound
	}

	if _, err := service.SkillRepo.GetSkillCategoryByID(req.CategoryId); err != nil {
		return errs.ErrSkillCategoryNotFound
	}

	if err := service.processSkillZip(zipPath, po.SkillSourceCustomUpload, upload.FileName, req.Icon, req.CategoryId); err != nil {
		return err
	}

	if err := service.HFSRepo.MarkUploadUsed(req.UploadId); err != nil {
		service.Worker.Logger().Errorf("PublishMarketSkill mark upload used failed: uploadId=%s, err=%v", req.UploadId, err)
	}
	os.RemoveAll(upload.TempDir)

	return nil
}

func (service *SkillService) ImportClawHubSkill(req *vo.AdminImportClawHubSkillReq) error {
	if _, err := service.SkillRepo.GetSkillCategoryByID(req.CategoryId); err != nil {
		return errs.ErrSkillCategoryNotFound
	}

	zipPath, err := service.ClawHubRepo.Download(req.Slug)
	if err != nil {
		return errs.ErrClawHubImportFailed
	}

	if err := service.processSkillZip(zipPath, po.SkillSourceClawHub, req.Slug, req.Icon, req.CategoryId); err != nil {
		return err
	}

	return nil
}

func (service *SkillService) SearchClawHub(req *vo.ClawHubSearchReq) (*repository.ClawHubSearchResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	return service.ClawHubRepo.Search(req.Query, limit)
}

func (service *SkillService) ExploreClawHub(req *vo.ClawHubExploreReq) (*repository.ClawHubExploreResponse, error) {
	return service.ClawHubRepo.Explore(req.Sort)
}

func (service *SkillService) GetClawHubSkillDetail(slug string) (*vo.AdminClawHubSkillDetailRsp, error) {
	content, err := service.ClawHubRepo.FetchFile(slug, "SKILL.md")
	if err != nil {
		return nil, errs.NewFormatError("clawhub fetch SKILL.md failed: %v", "ClawHub 获取 SKILL.md 失败: %v", err)
	}

	name, description, err := util.ParseSkillFrontmatter(content)
	if err != nil {
		return nil, errs.NewFormatError("parse SKILL.md failed: %v", "解析 SKILL.md 失败: %v", err)
	}

	return &vo.AdminClawHubSkillDetailRsp{
		Slug:        slug,
		Name:        name,
		Description: description,
		Content:     content,
		Version:     "",
	}, nil
}

func (service *SkillService) processSkillZip(zipPath string, source string, sourceUrl string, icon string, categoryId int) error {
	skillName, description, err := service.parseSkillZip(zipPath)
	if err != nil {
		return err
	}

	if _, err := service.SkillRepo.FindMarketSkillByName(skillName); err == nil {
		return errs.ErrMarketSkillNameAlreadyExists
	}

	hubSkillsDir := config.Get().Paths.SkillsHub
	destDir := filepath.Join(hubSkillsDir, skillName)

	if err := service.extractAndVerifySkillZip(zipPath, destDir); err != nil {
		return err
	}

	skill := &po.SkillMarket{
		Name:        skillName,
		Description: description,
		Icon:        icon,
		Source:      source,
		SourceUrl:   sourceUrl,
		CategoryId:  categoryId,
		Status:      po.SkillStatusEnabled,
	}

	if err := service.SkillRepo.CreateMarketSkill(skill); err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("create market skill: %w", err)
	}

	return nil
}

func (service *SkillService) parseSkillZip(zipPath string) (name, description string, err error) {
	const maxExtractSize = 50 * 1024 * 1024

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", "", errs.ErrSkillPublishInvalidZip
	}
	defer reader.Close()

	var hasSkillMd bool
	var totalSize int64
	var skillMdContent string

	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		entryName := filepath.ToSlash(f.Name)
		if util.IsSystemZipEntry(entryName) {
			continue
		}

		totalSize += int64(f.UncompressedSize64)

		if totalSize > maxExtractSize {
			return "", "", errs.ErrSkillPublishTooLarge
		}

		if strings.EqualFold(filepath.Base(entryName), "SKILL.md") {
			if !hasSkillMd || entryName == "SKILL.md" {
				hasSkillMd = true
				rc, e := f.Open()
				if e != nil {
					return "", "", errs.ErrSkillPublishInvalidZip
				}
				data, e := io.ReadAll(rc)
				rc.Close()
				if e != nil {
					return "", "", errs.ErrSkillPublishInvalidZip
				}
				skillMdContent = string(data)
			}
		}
	}

	if !hasSkillMd {
		return "", "", errs.ErrSkillPublishNoSkillMd
	}

	skillName, description, parseErr := util.ParseSkillFrontmatter(skillMdContent)
	if parseErr != nil {
		return "", "", errs.NewFormatError("parse SKILL.md failed: %v", "解析 SKILL.md 失败: %v", parseErr)
	}

	return skillName, description, nil
}

func (service *SkillService) extractAndVerifySkillZip(zipPath, destDir string) error {
	if err := util.ExtractZip(zipPath, destDir); err != nil {
		return errs.ErrSkillPublishInvalidZip
	}

	if _, err := os.Stat(filepath.Join(destDir, "SKILL.md")); err == nil {
		return nil
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		os.RemoveAll(destDir)
		return errs.ErrSkillPublishInvalidZip
	}

	var subDirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() && !isSystemDirName(e.Name()) {
			subDirs = append(subDirs, e)
		}
	}

	if len(subDirs) != 1 {
		os.RemoveAll(destDir)
		return errs.ErrSkillPublishNoSkillMd
	}

	subDir := filepath.Join(destDir, subDirs[0].Name())
	if _, err := os.Stat(filepath.Join(subDir, "SKILL.md")); os.IsNotExist(err) {
		os.RemoveAll(destDir)
		return errs.ErrSkillPublishNoSkillMd
	}

	subEntries, err := os.ReadDir(subDir)
	if err != nil {
		os.RemoveAll(destDir)
		return errs.ErrSkillPublishInvalidZip
	}
	for _, e := range subEntries {
		if isSystemFileName(e.Name()) {
			continue
		}
		if err := os.Rename(filepath.Join(subDir, e.Name()), filepath.Join(destDir, e.Name())); err != nil {
			os.RemoveAll(destDir)
			return errs.ErrSkillPublishInvalidZip
		}
	}

	os.RemoveAll(subDir)
	return nil
}

func isSystemDirName(name string) bool {
	return strings.HasPrefix(name, ".")
}

func isSystemFileName(name string) bool {
	return strings.HasPrefix(name, "._") || name == ".DS_Store" || strings.HasPrefix(name, ".")
}

func (service *SkillService) ListSkillCategories(req *vo.AdminSkillCategoryListReq) (*infra.PageResponse, error) {
	categories, pr, err := service.SkillRepo.PaginateSkillCategories(req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminSkillCategoryItem, 0, len(categories))
	for _, c := range categories {
		skillCount, _ := service.SkillRepo.CountSkillsByCategoryId(c.CategoryId)
		items = append(items, vo.AdminSkillCategoryItem{
			CategoryId: c.CategoryId,
			Name:       c.Name,
			Icon:       c.Icon,
			IsDefault:  c.IsDefault,
			SkillCount: skillCount,
			Updated:    c.Updated,
		})
	}

	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *SkillService) GetAllSkillCategories() ([]vo.AdminSkillCategoryItem, error) {
	return service.SkillRepo.GetAllSkillCategories()
}

func (service *SkillService) GetSkillCategoryDetail(categoryId int) (*vo.AdminSkillCategoryDetailRsp, error) {
	cat, err := service.SkillRepo.GetSkillCategoryByID(categoryId)
	if err != nil {
		return nil, errs.ErrSkillCategoryNotFound
	}
	return &vo.AdminSkillCategoryDetailRsp{
		CategoryId: cat.CategoryId,
		Name:       cat.Name,
		Icon:       cat.Icon,
		IsDefault:  cat.IsDefault,
		Created:    cat.Created,
		Updated:    cat.Updated,
	}, nil
}

func (service *SkillService) CreateSkillCategory(req *vo.AdminCreateSkillCategoryReq) error {
	cat := &po.SkillCategory{
		Name: req.Name,
		Icon: req.Icon,
	}
	return service.SkillRepo.CreateSkillCategory(cat)
}

func (service *SkillService) UpdateSkillCategory(categoryId int, req *vo.AdminUpdateSkillCategoryReq) error {
	cat, err := service.SkillRepo.GetSkillCategoryByID(categoryId)
	if err != nil {
		return errs.ErrSkillCategoryNotFound
	}
	if cat.IsDefault == 1 {
		return errs.ErrSkillCategoryIsDefault
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}

	if len(updates) == 0 {
		return nil
	}

	return service.SkillRepo.UpdateSkillCategory(categoryId, updates)
}

func (service *SkillService) DeleteSkillCategory(categoryId int) error {
	cat, err := service.SkillRepo.GetSkillCategoryByID(categoryId)
	if err != nil {
		return errs.ErrSkillCategoryNotFound
	}
	if cat.IsDefault == 1 {
		return errs.ErrSkillCategoryIsDefault
	}

	defaultCat, err := service.SkillRepo.GetDefaultSkillCategory()
	if err != nil {
		return errs.ErrSkillCategoryNotFound
	}

	if err := service.SkillRepo.ReassignSkillsToCategory(categoryId, defaultCat.CategoryId); err != nil {
		return err
	}
	if err := service.SkillRepo.ReassignUserSkillsToCategory(categoryId, defaultCat.CategoryId); err != nil {
		return err
	}

	return service.SkillRepo.SoftDeleteSkillCategory(categoryId)
}

func (service *SkillService) batchGetCategoryMap(skills []po.SkillMarket) map[int]po.SkillCategory {
	ids := make([]int, 0)
	for _, s := range skills {
		if s.CategoryId > 0 {
			ids = append(ids, s.CategoryId)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	cats, err := service.SkillRepo.BatchGetSkillCategories(ids)
	if err != nil {
		return nil
	}

	m := make(map[int]po.SkillCategory, len(cats))
	for _, c := range cats {
		m[c.CategoryId] = c
	}
	return m
}

func (service *SkillService) translateParseError(err error) error {
	if err == util.ErrInvalidSkillFormat {
		return errs.ErrSystemSkillInvalidFormat
	}
	if err == util.ErrInvalidSkillName {
		return errs.ErrSystemSkillInvalidName
	}
	if err == util.ErrMissingSkillName {
		return errs.ErrSystemSkillNameRequired
	}
	return errs.NewFormatError("invalid skill content: %v", "技能内容无效: %v", err)
}

func (service *SkillService) ListAvailableSkills(userId string) ([]vo.UserAvailableSkillItem, error) {
	userSkills, err := service.SkillRepo.FindInstalledUserSkills(userId)
	if err != nil {
		return nil, err
	}

	categoryMap := service.batchGetSkillCategoryMap(userSkills)

	items := make([]vo.UserAvailableSkillItem, 0, len(userSkills))
	for _, s := range userSkills {
		if s.AlwaysOn == 1 {
			continue
		}

		item := vo.UserAvailableSkillItem{
			UserSkillId: s.UserSkillId,
			SkillName:   s.SkillName,
			Description: s.Description,
			Icon:        s.Icon,
			Source:      s.Source,
			CategoryId:  s.CategoryId,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
		}
		items = append(items, item)
	}

	return items, nil
}

func (service *SkillService) ListAvailableSkillsByIDs(userId string, userSkillIds []int) ([]vo.UserAvailableSkillItem, error) {
	userSkills, err := service.SkillRepo.FindInstalledUserSkillsByIDs(userId, userSkillIds)
	if err != nil {
		return nil, err
	}

	categoryMap := service.batchGetSkillCategoryMap(userSkills)

	items := make([]vo.UserAvailableSkillItem, 0, len(userSkills))
	for _, s := range userSkills {
		item := vo.UserAvailableSkillItem{
			UserSkillId: s.UserSkillId,
			SkillName:   s.SkillName,
			Description: s.Description,
			Icon:        s.Icon,
			Source:      s.Source,
			CategoryId:  s.CategoryId,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
		}
		items = append(items, item)
	}

	return items, nil
}

func (service *SkillService) batchGetSkillCategoryMap(skills []po.UserSkill) map[int]po.SkillCategory {
	ids := make([]int, 0)
	for _, s := range skills {
		if s.CategoryId > 0 {
			ids = append(ids, s.CategoryId)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	cats, err := service.SkillRepo.BatchGetSkillCategories(ids)
	if err != nil {
		return nil
	}

	m := make(map[int]po.SkillCategory, len(cats))
	for _, c := range cats {
		m[c.CategoryId] = c
	}
	return m
}

func (service *SkillService) ListUserSkills(userId string, req *vo.UserSkillListReq) (*infra.PageResponse, error) {
	skills, pr, err := service.SkillRepo.PaginateUserSkills(userId, req)
	if err != nil {
		return nil, err
	}

	categoryMap := service.batchGetSkillCategoryMap(skills)
	items := make([]vo.UserSkillItem, 0, len(skills))
	for _, s := range skills {
		item := vo.UserSkillItem{
			UserSkillId:   s.UserSkillId,
			SkillName:     s.SkillName,
			Description:   s.Description,
			Icon:          s.Icon,
			MarketSkillId: s.MarketSkillId,
			CategoryId:    s.CategoryId,
			Source:        s.Source,
			InstallStatus: s.InstallStatus,
			InstallError:  s.InstallError,
			AlwaysOn:      s.AlwaysOn,
			Created:       s.Created,
			Updated:       s.Updated,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
		}
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

func (service *SkillService) GetUserSkillDetail(userSkillId int, userId string) (*vo.UserSkillDetailRsp, error) {
	skill, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return nil, errs.ErrUserSkillNotFound
	}

	var categoryName string
	if skill.CategoryId > 0 {
		if cat, err := service.SkillRepo.GetSkillCategoryByID(skill.CategoryId); err == nil {
			categoryName = cat.Name
		}
	}

	var content string
	sb, sberr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sberr != nil {
		return nil, sberr
	}
	backend, err := sb.NewBackend()
	if err == nil {
		skillMdPath := filepath.Join(sb.GetWorkspace(), "skills", skill.SkillName, "SKILL.md")
		if result, readErr := backend.Read(context.Background(), &filesystem.ReadRequest{FilePath: skillMdPath}); readErr == nil {
			content = result.Content
		}
	}

	return &vo.UserSkillDetailRsp{
		UserSkillId:   skill.UserSkillId,
		SkillName:     skill.SkillName,
		Description:   skill.Description,
		Icon:          skill.Icon,
		MarketSkillId: skill.MarketSkillId,
		CategoryId:    skill.CategoryId,
		CategoryName:  categoryName,
		Source:        skill.Source,
		InstallStatus: skill.InstallStatus,
		InstallError:  skill.InstallError,
		Content:       content,
		Created:       skill.Created,
		Updated:       skill.Updated,
	}, nil
}

func (service *SkillService) InstallSkill(userId string, req *vo.UserSkillInstallReq) (*vo.UserSkillInstallRsp, error) {

	marketSkill, err := service.SkillRepo.GetMarketSkillByID(req.SkillId)
	if err != nil {
		return nil, errs.ErrMarketSkillNotFound
	}
	if !marketSkill.IsEnabled() {
		return nil, errs.ErrMarketSkillNotAvailable
	}

	existing, err := service.SkillRepo.FindUserSkillByUserIdAndName(userId, marketSkill.Name)
	if err == nil && existing != nil {
		return nil, errs.ErrUserSkillAlreadyInstalled
	}

	userSkill := &po.UserSkill{
		UserId:        userId,
		SkillName:     marketSkill.Name,
		Description:   marketSkill.Description,
		Icon:          marketSkill.Icon,
		MarketSkillId: marketSkill.SkillId,
		CategoryId:    marketSkill.CategoryId,
		Source:        "market",
		InstallStatus: po.UserSkillInstalled,
	}
	if err := service.SkillRepo.CreateUserSkill(userSkill); err != nil {
		return nil, err
	}

	if err := service.SkillRepo.IncrMarketSkillInstalledCount(marketSkill.SkillId); err != nil {
		service.Worker.Logger().Errorf("IncrMarketSkillInstalledCount failed: skillId=%d, err=%v", marketSkill.SkillId, err)
	}

	service.Worker.DeferRecycle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("SkillService InstallSkill panic: %v", r)
			}
		}()

		if err := service.copySkillToUserSpace(userId, marketSkill.Name, userSkill.UserSkillId); err != nil {
			service.Worker.Logger().Errorf("copySkillToUserSpace %v", err)
			return
		}

	}()

	return &vo.UserSkillInstallRsp{
		UserSkillId: userSkill.UserSkillId,
	}, nil
}

func (service *SkillService) RetryInstallSkill(userSkillId int, userId string) error {
	skill, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return errs.ErrUserSkillNotFound
	}
	if skill.InstallStatus != po.UserSkillInstallFailed {
		return errs.ErrUserSkillNotFailed
	}

	chatModel, err := service.AIModelRepo.GetDefaultChatModel()
	if err != nil {
		return errs.ErrDefaultModelNotSet
	}

	if err := service.SkillRepo.UpdateUserSkill(userSkillId, map[string]interface{}{
		"install_status": po.UserSkillInstallProgress,
		"install_error":  "",
	}); err != nil {
		return err
	}

	service.Worker.DeferRecycle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("SkillService RetryInstallSkill panic: %v", r)
			}
		}()

		if err := service.copySkillToUserSpace(userId, skill.SkillName, userSkillId); err != nil {
			service.Worker.Logger().Errorf("copySkillToUserSpace %v", err)
			return
		}
		service.installSkillAsync(userSkillId, userId, skill.SkillName, chatModel)
	}()

	return nil
}

func (service *SkillService) copySkillToUserSpace(userId string, skillName string, userSkillId int) error {
	box, boxerr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if boxerr != nil {
		return boxerr
	}
	srcPath := filepath.Join(config.Get().Paths.SkillsHub, skillName)
	dstPath := filepath.Join(box.GetWorkspace(), "skills", skillName)

	if err := box.Upload(srcPath, dstPath); err != nil {
		service.SkillRepo.UpdateUserSkill(userSkillId, map[string]interface{}{
			"install_status": po.UserSkillInstallFailed,
			"install_error":  "复制技能文件失败: " + err.Error(),
		})
		return err
	}
	return nil
}

func (service *SkillService) installSkillAsync(userSkillId int, userId string, skillName string, chatModel iface.BaseChatModel) {

	sb, sbErr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sbErr == nil {
		if installed, err := sb.IsSkillInstalled(skillName); err == nil && installed {
			service.Worker.Logger().Infof("skill %s already installed, skip SystemAgent", skillName)
			return
		}
	}

	sysCfg, err := service.SystemSettingRepo.LoadConfig()
	if err != nil {
		sysCfg = repository.NewDefaultSystemConfig()
	}

	content := sysCfg.ClawHubToken
	if err := sb.MarkSkillInstalled(skillName, content); err != nil {
		service.Worker.Logger().Errorf("MarkSkillInstalled failed: skillName=%s, err=%v", skillName, err)
	}
}

func (service *SkillService) GetUserSkillStatus(userSkillId int, userId string) (*vo.UserSkillStatusRsp, error) {
	skill, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return nil, errs.ErrUserSkillNotFound
	}
	return &vo.UserSkillStatusRsp{
		UserSkillId:   skill.UserSkillId,
		InstallStatus: skill.InstallStatus,
		InstallError:  skill.InstallError,
	}, nil
}

func (service *SkillService) DeleteUserSkill(userSkillId int, userId string) error {
	skill, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return errs.ErrUserSkillNotFound
	}

	if err := service.SkillRepo.DeleteUserSkill(userSkillId, userId); err != nil {
		return err
	}

	_ = service.PersonaRepo.DeletePersonaToolBySkillId(userSkillId)

	sb, sberr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sberr != nil {
		return sberr
	}
	skillDir := filepath.Join(sb.GetWorkspace(), "skills", skill.SkillName)
	sb.DeleteFile(skillDir)

	return nil
}

func (service *SkillService) UpdateUserSkill(userSkillId int, userId string, req *vo.UserSkillUpdateReq) error {
	_, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return errs.ErrUserSkillNotFound
	}

	updates := make(map[string]interface{})
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.CategoryId != nil {
		updates["category_id"] = *req.CategoryId
	}
	if len(updates) == 0 {
		return nil
	}

	return service.SkillRepo.UpdateUserSkill(userSkillId, updates)
}

func (service *SkillService) ToggleAlwaysOn(userSkillId int, userId string, req *vo.UserSkillToggleAlwaysOnReq) error {
	_, err := service.SkillRepo.GetUserSkillByID(userSkillId, userId)
	if err != nil {
		return errs.ErrUserSkillNotFound
	}

	return service.SkillRepo.UpdateUserSkill(userSkillId, map[string]interface{}{
		"always_on": req.AlwaysOn,
	})
}

func (service *SkillService) RefreshUserSkills(userId string) (*vo.UserSkillRefreshRsp, error) {
	rsp := &vo.UserSkillRefreshRsp{}

	sb, sberr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if sberr != nil {
		return nil, sberr
	}
	backend, err := sb.NewBackend()
	if err != nil {
		return nil, err
	}
	skillsDir := filepath.Join(sb.GetWorkspace(), "skills")

	entries, err := backend.LsInfo(context.Background(), &filesystem.LsInfoRequest{Path: skillsDir})
	if err != nil {
		return nil, err
	}
	if entries == nil {
		return rsp, nil
	}

	existingSkills, _, err := service.SkillRepo.PaginateUserSkills(userId, &vo.UserSkillListReq{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, err
	}
	existingMap := make(map[string]*po.UserSkill, len(existingSkills))
	for i := range existingSkills {
		existingMap[existingSkills[i].SkillName] = &existingSkills[i]
	}

	var defaultCategoryId int
	if defaultCat, err := service.SkillRepo.GetDefaultSkillCategory(); err == nil {
		defaultCategoryId = defaultCat.CategoryId
	}

	dirSet := make(map[string]bool, len(entries))
	for _, entry := range entries {
		skillName := entry.Path
		dirSet[skillName] = true

		skillMdPath := filepath.Join(skillsDir, skillName, "SKILL.md")
		result, err := backend.Read(context.Background(), &filesystem.ReadRequest{FilePath: skillMdPath})
		if err != nil {
			os.RemoveAll(filepath.Join(skillsDir, skillName))
			continue
		}

		name, description, err := util.ParseSkillFrontmatter(result.Content)
		if err != nil {
			os.RemoveAll(filepath.Join(skillsDir, skillName))
			continue
		}
		if name != skillName {
			os.RemoveAll(filepath.Join(skillsDir, skillName))
			continue
		}

		if _, ok := existingMap[skillName]; !ok {

			userSkill := &po.UserSkill{
				UserId:        userId,
				SkillName:     name,
				Description:   description,
				MarketSkillId: 0,
				CategoryId:    defaultCategoryId,
				Source:        "custom",
				InstallStatus: po.UserSkillInstalled,
			}
			if err := service.SkillRepo.CreateUserSkill(userSkill); err != nil {
				freedom.Logger().Errorf("RefreshUserSkills create user_skill error: %v", err)
				continue
			}
			rsp.Added++

		}
	}

	for _, existing := range existingSkills {
		if !dirSet[existing.SkillName] {
			if err := service.SkillRepo.DeleteUserSkill(existing.UserSkillId, userId); err != nil {
				freedom.Logger().Errorf("RefreshUserSkills delete orphan user_skill error: %v", err)
				continue
			}
			rsp.Removed++
		}
	}

	return rsp, nil
}

func (service *SkillService) ListMarketSkillsForUser(userId string, req *vo.UserMarketSkillListReq) (*infra.PageResponse, error) {
	skills, pr, err := service.SkillRepo.PaginateUserMarketSkills(req)
	if err != nil {
		return nil, err
	}

	installedNames, err := service.SkillRepo.FindAllUserSkillNames(userId)
	if err != nil {
		return nil, err
	}
	installedSet := make(map[string]bool, len(installedNames))
	for _, n := range installedNames {
		installedSet[n] = true
	}

	categoryMap := service.batchGetCategoryMap(skills)
	items := make([]vo.UserMarketSkillItem, 0, len(skills))
	for _, s := range skills {
		item := vo.UserMarketSkillItem{
			SkillId:        s.SkillId,
			Name:           s.Name,
			Description:    s.Description,
			Icon:           s.Icon,
			Source:         s.Source,
			CategoryId:     s.CategoryId,
			InstalledCount: s.InstalledCount,
			UserInstalled:  installedSet[s.Name],
			Updated:        s.Updated,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
		}
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

func (service *SkillService) GetMarketSkillDetailForUser(skillId int) (*vo.AdminMarketSkillDetailRsp, error) {
	return service.GetMarketSkillDetail(skillId)
}

func (service *SkillService) ListSkillCategoriesForUser() ([]vo.SkillCategoryItem, error) {
	categories, err := service.SkillRepo.GetAllCategories()
	if err != nil {
		return nil, err
	}

	items := make([]vo.SkillCategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, vo.SkillCategoryItem{
			CategoryId: c.CategoryId,
			Name:       c.Name,
			Icon:       c.Icon,
		})
	}
	return items, nil
}

func (service *SkillService) CheckSkillNameConflicts(userId string, skillIds []int) error {
	if len(skillIds) <= 1 {
		return nil
	}

	names, err := service.SkillRepo.GetUserSkillNamesByIDs(userId, skillIds)
	if err != nil {
		return err
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			return errs.NewFormatError(
				"skill name conflict: '%s' is duplicated",
				"技能名称冲突：'%s' 重复",
				name,
			)
		}
		seen[name] = true
	}
	return nil
}

func (service *SkillService) ShareSkill(userId string, req *vo.ShareSkillReq) (*vo.ShareSkillRsp, error) {
	userSkill, err := service.SkillRepo.GetUserSkillByID(req.UserSkillId, userId)
	if err != nil {
		return nil, errs.ErrUserSkillNotFound
	}
	if userSkill.InstallStatus != po.UserSkillInstalled {
		return nil, errs.NewFormatError(
			"skill is not installed, cannot share",
			"技能未安装，无法共享",
		)
	}

	if userSkill.Source != "custom" {
		return nil, errs.NewFormatError(
			"only custom skills can be shared",
			"仅自定义技能可共享",
		)
	}

	existing, err := service.SkillRepo.GetSkillShareBySkillName(userSkill.SkillName)
	if err == nil && existing != nil {
		return nil, errs.NewFormatError(
			"skill '%s' has already been shared",
			"技能 '%s' 已被共享",
			userSkill.SkillName,
		)
	}

	box, boxErr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if boxErr != nil {
		return nil, boxErr
	}

	srcPath, err := box.Download(filepath.Join(box.GetWorkspace(), "skills", userSkill.SkillName))
	fmt.Println(infra.GetUserName(service.Worker), filepath.Join("skills", userSkill.SkillName), "sp", srcPath)
	if err != nil {
		return nil, errs.NewFormatError(
			"source skill files not found, cannot share",
			"源技能文件不存在，无法共享",
		)
	}
	dstPath := filepath.Join(config.Get().GetSkillShareDir(), userSkill.SkillName)

	os.RemoveAll(dstPath)
	if err := util.CopyDir(srcPath, dstPath); err != nil {
		return nil, errs.NewFormatError(
			"failed to copy skill files: %v",
			"复制技能文件失败: %v",
			err,
		)
	}

	share := &po.SkillShare{
		OwnerId:     userId,
		SkillName:   userSkill.SkillName,
		Description: userSkill.Description,
		Icon:        userSkill.Icon,
		CategoryId:  userSkill.CategoryId,
		Note:        req.Note,
	}
	if err := service.SkillRepo.CreateSkillShare(share); err != nil {
		os.RemoveAll(dstPath)
		return nil, err
	}

	return &vo.ShareSkillRsp{ShareId: share.ShareId}, nil
}

func (service *SkillService) UpdateSharedSkill(shareId int, userId string) error {
	share, err := service.SkillRepo.GetSkillShareByID(shareId)
	if err != nil {
		return errs.NewFormatError(
			"shared skill not found",
			"共享技能不存在",
		)
	}
	if share.OwnerId != userId {
		return errs.NewFormatError(
			"only the owner can update shared skill",
			"仅共享者可更新共享技能",
		)
	}

	box, boxErr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if boxErr != nil {
		return boxErr
	}
	srcPath, err := box.Download(filepath.Join(box.GetWorkspace(), "skills", share.SkillName))
	if err != nil {
		return errs.NewFormatError(
			"source skill files not found, cannot update",
			"源技能文件不存在，无法更新",
		)
	}
	dstPath := filepath.Join(config.Get().GetSkillShareDir(), share.SkillName)

	os.RemoveAll(dstPath)
	if err := util.CopyDir(srcPath, dstPath); err != nil {
		service.SkillRepo.DeleteSkillShare(shareId)
		return errs.NewFormatError(
			"failed to copy skill files: %v",
			"复制技能文件失败: %v",
			err,
		)
	}

	return service.SkillRepo.UpdateSkillShare(shareId, map[string]interface{}{
		"description": share.Description,
	})
}

func (service *SkillService) batchGetCategoryMapByIds(ids []int) map[int]po.SkillCategory {
	if len(ids) == 0 {
		return nil
	}
	cats, err := service.SkillRepo.BatchGetSkillCategories(ids)
	if err != nil {
		return nil
	}
	m := make(map[int]po.SkillCategory, len(cats))
	for _, c := range cats {
		m[c.CategoryId] = c
	}
	return m
}

func (service *SkillService) ListSkillShares(userId string, req *vo.SkillShareListReq) (*infra.PageResponse, error) {
	shares, pr, err := service.SkillRepo.PaginateSkillShares(req)
	if err != nil {
		return nil, err
	}

	// 批量查询分类名称
	categoryIds := make([]int, 0)
	for _, s := range shares {
		if s.CategoryId > 0 {
			categoryIds = append(categoryIds, s.CategoryId)
		}
	}
	categoryMap := service.batchGetCategoryMapByIds(categoryIds)

	isAdmin := infra.IsAdmin(service.Worker)
	items := make([]vo.SkillShareItem, 0, len(shares))
	for _, s := range shares {
		ownerName := ""
		if u, e := service.UserRepo.FindByUserId(s.OwnerId); e == nil {
			ownerName = u.Username
		}
		item := vo.SkillShareItem{
			ShareId:      s.ShareId,
			OwnerId:      s.OwnerId,
			OwnerName:    ownerName,
			SkillName:    s.SkillName,
			Description:  s.Description,
			Icon:         s.Icon,
			CategoryId:   s.CategoryId,
			Note:         s.Note,
			InstallCount: s.InstallCount,
			Created:      s.Created.Format("2006-01-02 15:04:05"),
			Updated:      s.Updated.Format("2006-01-02 15:04:05"),
			CanDelete:    isAdmin || s.OwnerId == userId,
		}
		if cat, ok := categoryMap[s.CategoryId]; ok {
			item.CategoryName = cat.Name
		}
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

func (service *SkillService) DeleteSkillShare(shareId int, userId string) error {
	share, err := service.SkillRepo.GetSkillShareByID(shareId)
	if err != nil {
		return errs.NewFormatError(
			"shared skill not found",
			"共享技能不存在",
		)
	}

	isAdmin := infra.IsAdmin(service.Worker)
	if share.OwnerId != userId && !isAdmin {
		return errs.NewFormatError(
			"only the owner or admin can delete shared skill",
			"仅共享者或管理员可删除共享技能",
		)
	}

	dstPath := filepath.Join(config.Get().GetSkillShareDir(), share.SkillName)
	os.RemoveAll(dstPath)

	return service.SkillRepo.DeleteSkillShare(shareId)
}

func (service *SkillService) GetSkillShareDetail(shareId int) (*vo.SkillShareDetailRsp, error) {
	share, err := service.SkillRepo.GetSkillShareByID(shareId)
	if err != nil {
		return nil, errs.NewFormatError(
			"shared skill not found",
			"共享技能不存在",
		)
	}

	ownerName := ""
	if u, e := service.UserRepo.FindByUserId(share.OwnerId); e == nil {
		ownerName = u.Username
	}

	var categoryName string
	if share.CategoryId > 0 {
		if cat, err := service.SkillRepo.GetSkillCategoryByID(share.CategoryId); err == nil {
			categoryName = cat.Name
		}
	}

	// 读取共享目录下的 SKILL.md 内容
	var content string
	skillMdPath := filepath.Join(config.Get().GetSkillShareDir(), share.SkillName, "SKILL.md")
	if data, err := os.ReadFile(skillMdPath); err == nil {
		content = string(data)
	}

	return &vo.SkillShareDetailRsp{
		ShareId:      share.ShareId,
		OwnerId:      share.OwnerId,
		OwnerName:    ownerName,
		SkillName:    share.SkillName,
		Description:  share.Description,
		Icon:         share.Icon,
		CategoryId:   share.CategoryId,
		CategoryName: categoryName,
		Note:         share.Note,
		Content:      content,
		Created:      share.Created.Format("2006-01-02 15:04:05"),
		Updated:      share.Updated.Format("2006-01-02 15:04:05"),
	}, nil
}

func (service *SkillService) InstallSharedSkill(userId string, shareId int) (*vo.UserSkillInstallRsp, error) {
	share, err := service.SkillRepo.GetSkillShareByID(shareId)
	if err != nil {
		return nil, errs.NewFormatError(
			"shared skill not found",
			"共享技能不存在",
		)
	}

	existing, err := service.SkillRepo.FindUserSkillByUserIdAndName(userId, share.SkillName)
	if err == nil && existing != nil {
		return nil, errs.ErrUserSkillAlreadyInstalled
	}

	userSkill := &po.UserSkill{
		UserId:        userId,
		SkillName:     share.SkillName,
		Description:   share.Description,
		CategoryId:    0,
		Source:        "share",
		InstallStatus: po.UserSkillInstalled,
	}
	if err := service.SkillRepo.CreateUserSkill(userSkill); err != nil {
		return nil, err
	}

	service.SkillRepo.IncrSkillShareInstallCount(share.ShareId)

	service.Worker.DeferRecycle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("SkillService InstallSharedSkill panic: %v", r)
			}
		}()
		if err := service.copySharedSkillToUserSpace(userId, share.SkillName, userSkill.UserSkillId); err != nil {
			service.Worker.Logger().Errorf("copySharedSkillToUserSpace %v", err)
			return
		}
	}()

	return &vo.UserSkillInstallRsp{
		UserSkillId: userSkill.UserSkillId,
	}, nil
}

func (service *SkillService) copySharedSkillToUserSpace(userId string, skillName string, userSkillId int) error {
	box, boxerr := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if boxerr != nil {
		return boxerr
	}
	srcPath := filepath.Join(config.Get().GetSkillShareDir(), skillName)
	dstPath := filepath.Join(box.GetWorkspace(), "skills", skillName)

	if err := box.Upload(srcPath, dstPath); err != nil {
		service.SkillRepo.UpdateUserSkill(userSkillId, map[string]interface{}{
			"install_status": po.UserSkillInstallFailed,
			"install_error":  "复制技能文件失败: " + err.Error(),
		})
		return err
	}
	return nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/backend/vo/errs"
	"raven/core/agent"
	"raven/core/knowledge"
	"raven/core/sandbox"
	"raven/core/tools"
	"raven/util"
	"strings"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *ChatService {
			return &ChatService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *ChatService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type ChatService struct {
	Worker		freedom.Worker
	MsgSessionRepo	*repository.MsgSessionRepository
	ModelRepo	*repository.ProviderRepository
	McpRepo		*repository.MCPRepository
	SkillRepo	*repository.SkillRepository
	PersonaRepo	*repository.PersonaRepository
	SysSettingRepo	*repository.SystemSettingRepository
	DailyStatsRepo	*repository.DailyStatsRepository
	HFSRepo		*repository.HFSRepository
}

func (service *ChatService) StartChat(
	ctx context.Context,
	userId string,
	req *vo.ChatReq,
	mcpService *McpService,
	skillService *SkillService,
) (*vo.ChatRsp, error) {
	if req.Content == "" {
		return nil, errs.ErrChatContentRequired
	}

	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession(userId, req)
	if err != nil {
		return nil, err
	}

	if session.Status == 1 {
		return nil, errs.ErrChatSessionActive
	}
	if _, ok := agent.GetRunner(session.SessionId); ok {
		return nil, errs.ErrChatSessionActive
	}

	reasoning := req.Reasoning == 1
	chatModel, aiModel, aiModelId, err := service.ModelRepo.CreateChatModelFromID(session.AIModelId, reasoning)
	if err != nil {
		return nil, err
	}
	session.AIModelId = aiModelId

	sysCfg, err := service.SysSettingRepo.LoadConfig()
	if err != nil {
		return nil, err
	}

	mcpObjects, mcpNames, err := service.buildMCPTools(userId, mcpIds, mcpService)
	if err != nil {
		return nil, err
	}
	fmt.Println(mcpNames)

	skillNames, err := service.SkillRepo.GetUserSkillNamesByIDs(userId, skillIds)
	if err != nil {
		return nil, err
	}

	alwaysOnNames, err := service.SkillRepo.FindAlwaysOnSkillNames(userId)
	if err != nil {
		return nil, err
	}
	skillNames = util.DedupStrings(append(skillNames, alwaysOnNames...))

	param := agent.AgentParam{
		Session:		session,
		MsgRepo:		service.MsgSessionRepo,
		ChatModel:		chatModel,
		UserRole:		userRole,
		SystemSkillProvider:	service.SkillRepo,
		SysCfg:			sysCfg,
		DailyStatsRepo:		service.DailyStatsRepo,
		Project:		session.Project,
		UserName:		infra.GetUserName(service.Worker),
	}

	attachmentTags, err := service.processAttachments(userId, req.Attachments)
	if err != nil {
		return nil, err
	}

	if err = service.MsgSessionRepo.SaveSession(session); err != nil {
		return nil, err
	}

	mainAgent, err := agent.NewMainAgent(param)
	if err != nil {
		return nil, err
	}

	for _, mcpObj := range mcpObjects {
		fmt.Println(mcpObj)
	}

	agent.RegisterRunnerHold(session.SessionId)
	service.Worker.DeferRecycle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("StartChat panic: %v", r)
			}
		}()

		compressModel, cerr := service.ModelRepo.GetDefaultChatModel()
		if cerr != nil {
			freedom.Logger().Warnf("GetDefaultChatModel for compress: %v", cerr)
			compressModel = chatModel
		}
		mainAgent.SetCompressModel(compressModel)

		if sysCfg.VisualEnabled {

			if visualModel, verr := service.ModelRepo.GetVisualChatModel(); verr == nil && visualModel != nil {
				mainAgent.SetVisualModel(visualModel)
			}
		}

		runner, err := mainAgent.NewRunner(ctx)
		if err != nil {
			service.Worker.Logger().Error(err)
			return
		}

		runner.OnComplete(func(event *agent.RunnerCompleteEvent) {
			service.ChatComplete()
			skillService.RefreshUserSkills(userId)
		})

		effectiveContent := req.Content + attachmentTags
		if err = runner.Query(ctx, effectiveContent); err != nil {
			service.Worker.Logger().Error(err)
			return
		}

		agent.RegisterRunner(session.SessionId, runner)
	}()

	rsp := &vo.ChatRsp{
		SessionId:	session.SessionId,
		Session: &vo.SessionDetailRsp{
			SessionId:		session.SessionId,
			Title:			session.Title,
			Status:			session.Status,
			PersonaId:		session.PersonaId,
			Project:		session.Project,
			AIModelId:		session.AIModelId,
			PromptTokensCount:	session.PromptTokensCount,
			CompletionTokensCount:	session.CompletionTokensCount,
			McpIds:			mcpIds,
			SkillIds:		skillIds,
			LastChatTime:		session.LastChatTime,
		},
	}

	if aiModel != nil {
		rsp.Session.ModelName = aiModel.DisplayName
		rsp.Session.ContextLimit = aiModel.ContextLenInTokens()
	}

	if persona != nil {
		rsp.Session.PersonaName = persona.Name
		rsp.Session.PersonaIcon = persona.Icon
	}

	return rsp, nil
}

func (service *ChatService) GetRunner(sessionId string, userId string) (*agent.MainRunner, error) {

	if _, err := service.MsgSessionRepo.GetUserSession(sessionId, userId); err != nil {
		return nil, errs.ErrSessionNotFound
	}

	runner, ok := agent.GetRunner(sessionId)
	if !ok {
		return nil, errs.ErrChatRunnerNotFound
	}
	return runner, nil
}

func (service *ChatService) StopChat(sessionId string, userId string) error {

	if _, err := service.MsgSessionRepo.GetUserSession(sessionId, userId); err != nil {
		return errs.ErrSessionNotFound
	}

	runner, ok := agent.GetRunner(sessionId)
	if !ok {
		return errs.ErrChatRunnerNotFound
	}
	if runner.IsStopped() {
		return errs.ErrChatRunnerNotFound
	}

	runner.Terminat()
	return nil
}

func (service *ChatService) CompressChat(sessionId string, userId string) (*vo.ChatCompressRsp, error) {
	return &vo.ChatCompressRsp{}, nil
}

func (service *ChatService) PollCompress(taskId string) (*vo.ChatCompressPollRsp, error) {
	result, err := service.MsgSessionRepo.GetCompressTaskStatus(taskId)
	if err != nil {

		return nil, errs.ErrChatCompressTaskNotFound
	}

	rsp := &vo.ChatCompressPollRsp{Status: result}
	if result == vo.CompressTaskStatusFailed {
		rsp.Message = "compress failed"
	}
	return rsp, nil
}

func (service *ChatService) resolveSession(
	userId string,
	req *vo.ChatReq,
) (session *po.Session, persona *po.UserPersona, mcpIds []int, skillIds []int, userRole string, err error) {

	if req.SessionId != nil && *req.SessionId != "" {
		session, err = service.MsgSessionRepo.GetUserSession(*req.SessionId, userId)
		if err != nil {
			return nil, nil, nil, nil, "", errs.ErrSessionNotFound
		}

		if session.PersonaId > 0 {
			persona, mcpIds, skillIds, userRole, err = service.loadPersonaConfig(session.PersonaId, userId)
			if err != nil {
				return nil, nil, nil, nil, "", err
			}
			session.AIModelId = persona.AIModelId
		} else {
			mcpIds = parseJSONInts(session.McpIds)
			skillIds = parseJSONInts(session.SkillIds)
		}

		session.LastChatTime = time.Now()
		return
	}

	if req.AIModelId <= 0 && req.PersonaId == nil {
		return nil, nil, nil, nil, "", errs.ErrChatModelRequired
	}

	session = &po.Session{
		UserId:		userId,
		Title:		truncateRunes(req.Content, 30),
		AIModelId:	req.AIModelId,
		Project:	req.Project,
		LastChatTime:	time.Now(),
	}

	if req.PersonaId != nil && *req.PersonaId > 0 {

		persona, mcpIds, skillIds, userRole, err = service.loadPersonaConfig(*req.PersonaId, userId)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		session.PersonaId = *req.PersonaId

		session.McpIds = ""
		session.SkillIds = ""
		session.AIModelId = persona.AIModelId

	} else {

		mcpIds = req.McpIds
		skillIds = req.SkillIds

		if len(mcpIds) > 0 {
			mcpJSON, _ := json.Marshal(mcpIds)
			session.McpIds = string(mcpJSON)
		}
		if len(skillIds) > 0 {
			skillJSON, _ := json.Marshal(skillIds)
			session.SkillIds = string(skillJSON)
		}
	}

	return
}

func (service *ChatService) loadPersonaConfig(
	personaId int,
	userId string,
) (persona *po.UserPersona, mcpIds []int, skillIds []int, userRole string, err error) {
	persona, err = service.PersonaRepo.GetUserPersonaByID(personaId, userId)
	if err != nil {
		return nil, nil, nil, "", errs.ErrPersonaNotFound
	}
	userRole = persona.RoleInfo

	tools, _ := service.PersonaRepo.ListPersonaToolsByPersona(personaId)
	for _, t := range tools {
		switch t.ToolType {
		case "mcp":
			mcpIds = append(mcpIds, t.ToolId)
		case "skill":
			skillIds = append(skillIds, t.ToolId)
		}
	}
	return
}

func (service *ChatService) buildMCPTools(
	userId string,
	mcpIds []int,
	mcpService *McpService,
) (mcpObjects []tools.MCP, mcpNames []string, err error) {
	if len(mcpIds) == 0 {
		return nil, nil, nil
	}

	endpoints, err := service.McpRepo.GetMCPEndpointsByIDs(mcpIds)
	if err != nil {
		return nil, nil, err
	}
	mcpObjects = make([]tools.MCP, 0, len(endpoints))
	mcpNames = make([]string, 0, len(endpoints))

	for _, ep := range endpoints {
		if ep.Deleted == 1 || ep.Status == po.MCPEndpointStatusDisabled {
			continue
		}

		mcpObj := mcpService.BuildMCPObjectFromEndpoint(&ep)

		mcpObjects = append(mcpObjects, mcpObj)
		mcpNames = append(mcpNames, ep.Name)
	}
	return
}

var imageExtensions = map[string]bool{
	".jpg":	true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp":	true, ".webp": true, ".svg": true, ".tiff": true, ".ico": true,
}

func formatFileSize(size int64) string {
	switch {
	case size >= 1<<30:
		return fmt.Sprintf("%.1fGB", float64(size)/(1<<30))
	case size >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(size)/(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%dKB", size/(1<<10))
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func (service *ChatService) processAttachments(userId string, uploadIds []string) (string, error) {
	if len(uploadIds) == 0 {
		return "", nil
	}

	var tags []string
	for _, uploadId := range uploadIds {
		tag, err := service.processOneAttachment(userId, uploadId)
		if err != nil {
			return "", errs.NewFormatError(
				"process attachment %s failed: %v",
				"处理附件 %s 失败: %v",
				uploadId, err,
			)
		}
		tags = append(tags, tag)
	}
	return strings.Join(tags, ""), nil
}

func (service *ChatService) processOneAttachment(userId string, uploadId string) (string, error) {
	upload, err := service.HFSRepo.GetUploadByUploadId(uploadId)
	if err != nil {
		return "", errs.NewFormatError(
			"upload not found: %s",
			"上传任务不存在: %s",
			uploadId,
		)
	}
	if upload.Status == po.UploadStatusPending || upload.Status == po.UploadStatusCancelled {
		return "", errs.NewFormatError(
			"upload not completed: %s (status=%d)",
			"上传尚未完成: %s (status=%d)",
			uploadId, upload.Status,
		)
	}

	srcPath := filepath.Join(upload.TempDir, upload.FileName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", errs.NewFormatError(
			"attachment file not found: %s",
			"附件文件不存在: %s",
			srcPath,
		)
	}

	sb, err := sandbox.NewSandbox(infra.GetUserName(service.Worker))
	if err != nil {
		return "", errs.NewFormatError(
			"create sandbox failed: %v",
			"创建沙盒失败: %v",
			err,
		)
	}

	ext := strings.ToLower(filepath.Ext(upload.FileName))
	newUUID := util.UUID()
	var uploadSrcPath, dstRelPath string

	if imageExtensions[ext] {
		uploadSrcPath = srcPath
		dstRelPath = "/temp/" + newUUID + ext
	} else {
		convertedPath := filepath.Join(upload.TempDir, newUUID+".md")
		if cerr := knowledge.ConvertFile(srcPath, convertedPath, false); cerr != nil {
			freedom.Logger().Warnf(
				"docling ConvertFile failed for %s (%s), falling back to original: %v",
				upload.FileName, uploadId, cerr,
			)
			uploadSrcPath = srcPath
			dstRelPath = "/temp/" + newUUID + ext
		} else {
			uploadSrcPath = convertedPath
			dstRelPath = "/temp/" + newUUID + ".md"
		}
	}

	dstAbsPath := filepath.Join(sb.GetWorkspace(), dstRelPath)
	if err := sb.Upload(uploadSrcPath, dstAbsPath); err != nil {
		return "", errs.NewFormatError(
			"upload to sandbox failed: %v",
			"上传文件到沙盒失败: %v",
			err,
		)
	}

	if uploadSrcPath != srcPath {
		os.Remove(uploadSrcPath)
	}

	if err := service.HFSRepo.MarkUploadUsed(uploadId); err != nil {
		freedom.Logger().Warnf("MarkUploadUsed failed for %s: %v", uploadId, err)
	}

	return fmt.Sprintf(
		"\n<raven-upload size=\"%s\">\n  %s\n</raven-upload>",
		formatFileSize(upload.FileSize),
		dstRelPath,
	), nil
}

func (service *ChatService) ChatComplete() {
}

func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

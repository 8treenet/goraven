package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/agent"
	"goraven/core/knowledge"
	"goraven/core/sandbox"
	"goraven/core/tools"
	"goraven/util"
	"os"
	"path/filepath"
	"runtime/debug"
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
		initiator.BindBooting(func(bootManager freedom.BootManager) {
			freedom.ServiceLocator().Call(func(service *ChatService) error {
				service.Worker.DeferRecycle()
				service.VerifierTimer()
				return nil
			})
		})
	})
}

// ChatService 对话服务，管理聊天会话和 AI Agent 运行器
type ChatService struct {
	Worker             freedom.Worker
	MsgSessionRepo     *repository.MsgSessionRepository
	ModelRepo          *repository.ProviderRepository
	McpRepo            *repository.MCPRepository
	SkillRepo          *repository.SkillRepository
	PersonaRepo        *repository.PersonaRepository
	SysSettingRepo     *repository.SystemSettingRepository
	DailyStatsRepo     *repository.DailyStatsRepository
	HFSRepo            *repository.HFSRepository
	SharedProjectRepo  *repository.TeamProjectRepository
	UserRepo           *repository.UserRepository
	AutomationTaskRepo *repository.AutomationTaskRepository
	AutomationExecRepo *repository.AutomationExecutionRepository
}

// dailyTokenQuota 加载用户并返回每日 Token 限额（单位 M，0=不限）与今日已用量。
// err 非 nil 表示用户查询失败，调用方按各自语义处理（交互对话忽略、后台任务中止）。
func (service *ChatService) dailyTokenQuota(userId string) (user *po.User, limit, used int, err error) {
	user, err = service.UserRepo.FindByUserId(userId)
	if err != nil {
		return nil, 0, 0, err
	}
	limit = user.DailyTokenLimit
	if limit > 0 {
		if u, statsErr := service.DailyStatsRepo.GetTodayTokenUsage(userId); statsErr == nil {
			used = u
		}
	}
	return
}

// StartChat 创建或复用会话，构建 Agent 并启动运行器
// 跨 service 的依赖通过参数传入，由 controller 负责注入
func (service *ChatService) StartChat(
	ctx context.Context,
	userId string,
	req *vo.ChatReq,
	skillService *SkillService,
) (rsp *vo.ChatRsp, err error) {
	if req.Content == "" {
		return nil, errs.ErrChatContentRequired
	}

	isNewSession := req.SessionId == nil || *req.SessionId == ""

	// 1. 解析或创建会话
	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession(userId, req)
	if err != nil {
		return nil, err
	}

	// 2. 检查是否有活跃运行器
	if session.Status == 1 {
		return nil, errs.ErrChatSessionActive
	}
	if _, ok := agent.GetRunner(session.SessionId); ok {
		return nil, errs.ErrChatSessionActive
	}

	// 2.1 每日 Token 限额预检：额度不足时直接拒绝，不进入 Agent
	_, dailyTokenLimit, dailyTokenUsed, quotaErr := service.dailyTokenQuota(userId)
	if quotaErr == nil && dailyTokenLimit > 0 && dailyTokenUsed >= dailyTokenLimit*1_000_000 {
		return nil, errs.ErrDailyTokenLimitExceeded
	}

	// 3. 创建聊天模型（同时返回模型元数据 po.AIModel，避免后续重复查询）
	reasoning := req.Reasoning == 1
	chatModel, aiModel, aiModelId, err := service.ModelRepo.CreateChatModelFromID(session.AIModelId, reasoning)
	if err != nil {
		return nil, err
	}
	session.AIModelId = aiModelId

	// 4. 加载系统配置
	sysCfg, err := service.SysSettingRepo.LoadConfig()
	if err != nil {
		return nil, err
	}

	// 5. 构建 MCP 工具列表和过滤器
	// 5.1 合入始终启用的 MCP（去重），原始 mcpIds 仅用于响应回显
	alwaysOnMcpIds, err := service.McpRepo.FindAlwaysOnMcpIds()
	if err != nil {
		return nil, err
	}
	effectiveMcpIds := util.DedupInts(append(mcpIds, alwaysOnMcpIds...))
	mcpObjects, mcpNames, err := service.buildMCPTools(userId, effectiveMcpIds)
	if err != nil {
		return nil, err
	}

	// 6. 构建技能过滤器
	skillNames, err := service.SkillRepo.GetUserSkillNamesByIDs(userId, skillIds)
	if err != nil {
		return nil, err
	}

	// 6.1 合入始终启用的技能（去重）
	alwaysOnNames, err := service.SkillRepo.FindAlwaysOnSkillNames(userId)
	if err != nil {
		return nil, err
	}
	skillNames = util.DedupStrings(append(skillNames, alwaysOnNames...))

	// 7. 构建 AgentParam
	resolvedProject, resolvedProjectWorkspace, teamProjectInfo, err := service.resolveProjectFromSession(session)
	if err != nil {
		return nil, err
	}

	// 团队项目加锁：同一时刻只能有一个 Agent 操作该项目
	spId, spLocked, err := service.lockTeamProject(session, teamProjectInfo)
	if err != nil {
		return nil, err
	}
	// 加锁后若 StartChat 在 goroutine 启动前因任何步骤失败返回 error，立即释放锁避免 30min 残留。
	// 成功路径 err==nil 不触发；goroutine 内由 OnComplete 负责释放。
	defer func() {
		if err != nil && spLocked {
			service.SharedProjectRepo.UnlockTeamProject(spId)
		}
	}()

	param := agent.AgentParam{
		Session:             session,
		MsgRepo:             service.MsgSessionRepo,
		ChatModel:           chatModel,
		UserRole:            userRole,
		SystemSkillProvider: service.SkillRepo,
		SysCfg:              sysCfg,
		DailyStatsRepo:      service.DailyStatsRepo,
		Project:             resolvedProject,
		ProjectWorkspace:    resolvedProjectWorkspace,
		UserName:            infra.GetUserName(service.Worker),
		DailyTokenLimit:     dailyTokenLimit,
		DailyTokenUsed:      dailyTokenUsed,
		WikiWriteMode:       strings.Contains(req.Content, "goraven-llmwiki"),
		AutomationTaskRepo:  service.AutomationTaskRepo,
		TaskMcpIds:          mcpIds,
		TaskSkillIds:        skillIds,
	}

	// 处理附件（查库、文档转 markdown、上传到沙盒临时目录）
	// 必须在 SaveSession 之前，避免处理失败产生孤儿会话
	attachmentTags, err := service.processAttachments(userId, req.Attachments)
	if err != nil {
		return nil, err
	}

	// 持久化会话（所有准备工作完成后才入库，避免孤儿会话）
	if err = service.MsgSessionRepo.SaveSession(session); err != nil {
		return nil, err
	}
	// 8. 创建 MainAgent
	mainAgent, err := agent.NewMainAgent(param)
	if err != nil {
		return nil, err
	}
	// 添加 MCP 工具
	for _, mcpObj := range mcpObjects {
		mainAgent.AddMCP(mcpObj)
	}

	// 设置 MCP 过滤器（仅允许选中的 MCP）
	mainAgent.SetMCPFilter(agent.NewSimpleMCPFilter(mcpNames))

	// 设置技能过滤器
	simpleSkillFilter := agent.NewSimpleSkillFilter(skillNames, func(name string) {
		service.DailyStatsRepo.AddToolDailyStats(userId, "skill", name)
	})
	mainAgent.SetSkillAccessor(simpleSkillFilter)
	// CAS 抢占会话构建锁：同会话已有构建中的/活跃的 Runner 时直接拒绝，
	// 防止并发 StartChat 导致双 Runner 交错写消息。
	if !agent.HoldRunner(session.SessionId) {
		return nil, errs.ErrChatSessionActive
	}
	service.Worker.DeferRecycle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("StartChat panic: %v\n%s", r, debug.Stack())
				agent.ClearRunnerHold(session.SessionId)
				if spLocked {
					service.SharedProjectRepo.UnlockTeamProject(spId)
				}
			}
		}()

		// 设置 Flash 模型（复用默认模型，reasoning=false）
		flashModel, cerr := service.ModelRepo.GetDefaultChatModel()
		if cerr != nil {
			freedom.Logger().Warnf("GetDefaultChatModel for flash: %v", cerr)
			flashModel = chatModel
		}
		mainAgent.SetFlashModel(flashModel)

		if sysCfg.VisualEnabled {
			// 设置多模态识别模型（仅 isVisual=1 的模型，不降级）
			if visualModel, verr := service.ModelRepo.GetVisualChatModel(); verr == nil && visualModel != nil {
				mainAgent.SetVisualModel(visualModel)
			}
		}

		// 9. 创建 Runner 并启动
		runner, err := mainAgent.NewRunner(ctx)
		if err != nil {
			service.Worker.Logger().Error(err)
			agent.ClearRunnerHold(session.SessionId)
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			return
		}

		runner.OnComplete(func(event *agent.RunnerCompleteEvent) {
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			service.ChatComplete()
			skillService.RefreshUserSkills(userId)
			if isNewSession {
				titler := agent.NewSessionTitler(flashModel, service.MsgSessionRepo, service.DailyStatsRepo, session.SessionId, userId)
				service.generateAndSaveTitle(titler, event.Reply, session.SessionId, userId, req.Content)
			}
		})

		// 构建实际发送内容（附件标签 + 用户消息）
		effectiveContent := req.Content + attachmentTags
		if err = runner.Query(ctx, effectiveContent); err != nil {
			service.Worker.Logger().Error(err)
			agent.ClearRunnerHold(session.SessionId)
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			return
		}

		// 10. 注册运行器（StartFetch 由 Stream 控制器触发，Runner 结束时自清理）
		agent.RegisterRunner(session.SessionId, runner)
	}()

	// 构建会话详情随响应返回，直接使用 resolveSession 和 CreateChatModelFromID 已查到的数据，无需重复查询
	rsp = &vo.ChatRsp{
		SessionId: session.SessionId,
		Session: &vo.SessionDetailRsp{
			SessionId: session.SessionId,
			Title:     session.Title,
			// We have committed to generating: the runner goroutine flips the
			// DB status to 1 via saveQuery, but that runs asynchronously after
			// this response is built. Return 1 so the frontend can render the
			// generating/background state immediately instead of a stale 0.
			Status:                1,
			PersonaId:             session.PersonaId,
			Project:               session.Project,
			TeamProject:           teamProjectInfo,
			AIModelId:             session.AIModelId,
			PromptTokensCount:     session.PromptTokensCount,
			CompletionTokensCount: session.CompletionTokensCount,
			McpIds:                mcpIds,
			SkillIds:              skillIds,
			LastChatTime:          session.LastChatTime,
		},
	}

	// 关联模型信息（直接复用 CreateChatModelFromID 返回的 aiModel，不重复查询）
	if aiModel != nil {
		rsp.Session.ModelName = aiModel.DisplayName
		rsp.Session.ModelIcon = aiModel.Icon
		rsp.Session.ContextLimit = aiModel.ContextLenInTokens()
	}

	// 关联角色信息（直接复用 resolveSession 返回的 persona，不重复查询）
	if persona != nil {
		rsp.Session.PersonaName = persona.Name
		rsp.Session.PersonaIcon = persona.Icon
	}

	return rsp, nil
}

// GetRunner 获取会话的活跃运行器，由 Stream 控制器调用 StartFetch 获取 SSE 通道
func (service *ChatService) GetRunner(sessionId string, userId string) (*agent.MainRunner, error) {
	// 校验会话归属
	if _, err := service.MsgSessionRepo.GetUserSession(sessionId, userId); err != nil {
		return nil, errs.ErrSessionNotFound
	}

	runner, ok := agent.GetRunner(sessionId)
	if !ok {
		return nil, errs.ErrChatRunnerNotFound
	}
	return runner, nil
}

// StopChat 停止会话的活跃运行器
func (service *ChatService) StopChat(sessionId string, userId string) error {
	// 校验会话归属
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

	// Terminat 触发后台 goroutine 退出，defer 里会自动 DeleteRunner
	runner.Terminat()
	return nil
}

// CompressChat 手动压缩上下文，返回任务 ID 供轮询
// 标记 session.status=1 防止并发操作，异步执行压缩，完成后标记 status=0
func (service *ChatService) CompressChat(sessionId string, userId string) (*vo.ChatCompressRsp, error) {
	// 校验会话归属
	session, err := service.MsgSessionRepo.GetUserSession(sessionId, userId)
	if err != nil {
		return nil, errs.ErrSessionNotFound
	}

	// 会话进行中不允许压缩
	if session.Status == 1 {
		return nil, errs.ErrChatSessionActive
	}

	// 创建 Flash 模型（优先 isFlash=1 > session.AIModelId > 默认模型）
	flashModel, err := service.ModelRepo.GetFlashChatModel(session.AIModelId)
	if err != nil {
		return nil, err
	}

	// 加载系统配置
	sysCfg, err := service.SysSettingRepo.LoadConfig()
	if err != nil {
		return nil, err
	}

	// 加载历史消息并转为 schema.Message
	messages, err := service.MsgSessionRepo.GetChatMessages(session.SessionId)
	if err != nil {
		return nil, err
	}
	history := agent.BuildHistoryFromMessages(messages)

	// 构建 Compress 对象
	compress := agent.NewCompress(flashModel, service.MsgSessionRepo, service.DailyStatsRepo, session.SessionId, session.UserId, sysCfg)

	// 标记 session 为进行中，防止并发
	if err := service.MsgSessionRepo.UpdateSessionStatus(sessionId, 1); err != nil {
		return nil, err
	}

	// 生成任务 ID 并写入缓存
	taskId := util.UUID()
	ctx := context.Background()
	service.MsgSessionRepo.SetCompressTaskStatus(taskId, vo.CompressTaskStatusRunning)

	// 仅异步执行压缩（前置校验已完成，可提前返回错误）
	service.Worker.DeferRecycle()
	go func() {
		status := vo.CompressTaskStatusDone
		if _, cerr := compress.ForceCompress(ctx, history); cerr != nil {
			freedom.Logger().Errorf("CompressChat failed: sessionId=%s, err=%v", sessionId, cerr)
			status = vo.CompressTaskStatusFailed
		}
		// 更新缓存状态
		service.MsgSessionRepo.SetCompressTaskStatus(taskId, status)
		// 压缩结束，恢复 session 状态
		_ = service.MsgSessionRepo.UpdateSessionStatus(sessionId, 0)
	}()

	return &vo.ChatCompressRsp{TaskId: taskId}, nil
}

// PollCompress 轮询压缩任务状态
func (service *ChatService) PollCompress(taskId string) (*vo.ChatCompressPollRsp, error) {
	result, err := service.MsgSessionRepo.GetCompressTaskStatus(taskId)
	if err != nil {
		// 任务不存在或已过期
		return nil, errs.ErrChatCompressTaskNotFound
	}

	rsp := &vo.ChatCompressPollRsp{Status: result}
	if result == vo.CompressTaskStatusFailed {
		rsp.Message = "compress failed"
	}
	return rsp, nil
}

// resolveSession 解析或创建会话，返回会话、角色对象（无角色时 nil）、mcpIds、skillIds、userRole
func (service *ChatService) resolveSession(
	userId string,
	req *vo.ChatReq,
) (session *po.Session, persona *po.UserPersona, mcpIds []int, skillIds []int, userRole string, err error) {
	// 已有会话
	if req.SessionId != nil && *req.SessionId != "" {
		session, err = service.MsgSessionRepo.GetUserSession(*req.SessionId, userId)
		if err != nil {
			return nil, nil, nil, nil, "", errs.ErrSessionNotFound
		}
		// 从会话快照或 persona_tool 解析 mcpIds/skillIds
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

	// AIModelId=0 表示使用默认模型（后端从默认池随机选取），仅拒绝非法负值
	if req.AIModelId < 0 && req.PersonaId == nil {
		return nil, nil, nil, nil, "", errs.ErrChatModelRequired
	}

	// 新建会话
	if req.Project != "" && req.TeamProjectId != nil {
		return nil, nil, nil, nil, "", errs.ErrProjectMutualExclusive
	}

	session = &po.Session{
		UserId:       userId,
		Title:        truncateRunes(req.Content, 30),
		AIModelId:    req.AIModelId,
		LastChatTime: time.Now(),
	}

	if req.TeamProjectId != nil && *req.TeamProjectId > 0 {
		sp, err := service.SharedProjectRepo.GetByID(*req.TeamProjectId)
		if err != nil {
			return nil, nil, nil, nil, "", errs.ErrTeamProjectNotFound
		}
		session.SharedProjectId = sp.Id
		session.Project = sp.ProjectName
	} else {
		session.Project = req.Project
	}

	if req.PersonaId != nil && *req.PersonaId > 0 {
		// 角色模式
		persona, mcpIds, skillIds, userRole, err = service.loadPersonaConfig(*req.PersonaId, userId)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		session.PersonaId = *req.PersonaId
		// 角色模式下 mcpIds/skillIds 不存到 session（读 persona_tool 表）
		session.McpIds = ""
		session.SkillIds = ""
		session.AIModelId = persona.AIModelId

	} else {
		// 自由模式
		mcpIds = req.McpIds
		skillIds = req.SkillIds
		// 自由模式下存快照到 session
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

// loadPersonaConfig 加载角色配置，返回 persona、mcpIds、skillIds、roleInfo
func (service *ChatService) loadPersonaConfig(
	personaId int,
	userId string,
) (persona *po.UserPersona, mcpIds []int, skillIds []int, userRole string, err error) {
	persona, err = service.PersonaRepo.GetUserPersonaByID(personaId, userId)
	if err != nil {
		return nil, nil, nil, "", errs.ErrPersonaNotFound
	}
	userRole = persona.RoleInfo

	// 从 persona_tool 关联表读取 mcpIds/skillIds
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

// buildMCPTools 构建 MCP 工具对象和名称列表
func (service *ChatService) buildMCPTools(
	userId string,
	mcpIds []int,
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
	//toolOwnerMap := make(map[string]string) // toolName -> endpoint display name
	for _, ep := range endpoints {
		if ep.Deleted == 1 || ep.Status == po.MCPEndpointStatusDisabled {
			continue
		}

		mcpObj := buildMCPObjectFromEndpoint(&ep)
		// toolNames, validateErr := tools.ValidateMCP(service.Worker.Context(), mcpObj, box)
		// if validateErr != nil {
		// 	continue
		// }

		// for _, name := range toolNames {
		// 	if prev, exists := toolOwnerMap[name]; exists {
		// 		return nil, nil, errs.NewFormatError(
		// 			"MCP tool name conflict: '%s' is exposed by both '%s' and '%s'",
		// 			"MCP 工具名称冲突：'%s' 同时存在于 '%s' 和 '%s'",
		// 			name, prev, ep.DisplayName,
		// 		)
		// 	}
		// 	toolOwnerMap[name] = ep.DisplayName
		// }

		mcpObjects = append(mcpObjects, mcpObj)
		mcpNames = append(mcpNames, ep.Name)
	}
	return
}

// imageExtensions 图片文件扩展名，这些文件不需要通过 docling 转换
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp": true, ".webp": true, ".svg": true, ".tiff": true, ".ico": true,
}

// formatFileSize 将字节数转换为可读的大小字符串
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

// processAttachments 处理所有附件：查库验证、文档转 markdown、上传到沙盒临时目录
// 必须在 SaveSession 之前调用，避免处理失败产生孤儿会话
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

// processOneAttachment 处理单个附件：查库、转换、上传、返回 goraven-upload 标签
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
		if cerr := knowledge.ConvertFile(srcPath, convertedPath); cerr != nil {
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
		"\n<goraven-upload size=\"%s\">\n  %s\n</goraven-upload>",
		formatFileSize(upload.FileSize),
		dstAbsPath,
	), nil
}

func (service *ChatService) ChatComplete() {
}

// generateAndSaveTitle 在首轮对话结束后调用标题生成器写回 session.title
// 失败时仅记日志，保留 resolveSession 写入的默认标题；使用独立 background context
func (service *ChatService) generateAndSaveTitle(titler *agent.SessionTitler, assistantReply, sessionId, userId, userContent string) {
	if titler == nil {
		return
	}
	if strings.TrimSpace(assistantReply) == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	title, err := titler.Generate(ctx, userContent, assistantReply)
	if err != nil {
		freedom.Logger().Warnf("generateSessionTitle sessionId=%s err=%v", sessionId, err)
		return
	}
	if err := service.MsgSessionRepo.UpdateSession(sessionId, userId, map[string]interface{}{
		"title":   title,
		"updated": time.Now(),
	}); err != nil {
		freedom.Logger().Warnf("UpdateSession title sessionId=%s err=%v", sessionId, err)
	}
}

// truncateRunes 截取字符串前 n 个 rune
func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// lockTeamProject 对团队项目加锁，同一时刻只能有一个 Agent 操作该项目。
func (service *ChatService) lockTeamProject(session *po.Session, teamProjectInfo *vo.TeamProjectInfo) (spId int, spLocked bool, err error) {
	if session.SharedProjectId > 0 {
		spId = session.SharedProjectId
	}

	if spId == 0 {
		return 0, false, nil
	}

	ok, lockErr := service.SharedProjectRepo.LockTeamProject(spId, session.SessionId)
	if lockErr != nil {
		return 0, false, lockErr
	}
	if !ok {
		return 0, false, errs.ErrTeamProjectBusy
	}
	service.SharedProjectRepo.IncrementVisitAndUpdateLastActive(spId)
	return spId, true, nil
}

// resolveProjectFromSession 从 session 解析项目信息
// 个人项目 (SharedProjectId==0)：projectName=session.Project, projectWorkspace="", teamInfo=nil
// 团队项目 (SharedProjectId>0)：projectName=项目目录名, projectWorkspace=团队项目目录, teamInfo=团队项目信息
func (service *ChatService) resolveProjectFromSession(session *po.Session) (projectName, projectWorkspace string, teamInfo *vo.TeamProjectInfo, err error) {
	if session.SharedProjectId <= 0 {
		return session.Project, "", nil, nil
	}
	tp, dbErr := service.SharedProjectRepo.GetByID(session.SharedProjectId)
	if dbErr != nil {
		return session.Project, "", nil, nil
	}
	creator, userErr := service.SharedProjectRepo.GetUserByID(tp.CreatorId)
	creatorName := ""
	if userErr == nil {
		creatorName = creator.Nickname
		if creatorName == "" {
			creatorName = creator.Username
		}
	}
	projectWorkspace = config.Get().GetTeamProjectDir()
	teamInfo = &vo.TeamProjectInfo{
		Id:          tp.Id,
		CreatorId:   tp.CreatorId,
		CreatorName: creatorName,
		ProjectName: tp.ProjectName,
		Description: tp.Description,
	}
	return tp.ProjectName, projectWorkspace, teamInfo, nil
}

// automationScanLimit 每轮扫描最多处理多少个到点任务
const automationScanLimit = 60

// VerifierTimer 启动后台调度协程：每 60 秒扫描一次到点的自动化任务并触发执行。
// 静默后台运行，无 SSE 事件；未初始化或 Preview 模式下不启动。
func (service *ChatService) VerifierTimer() {
	if !config.Get().System.Initialized {
		return
	}
	if config.Get().Behavior.PreviewUser != "" {
		return
	}
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			service.scanAutomationTasks()
		}
	}()
}

// scanAutomationTasks 扫描到点任务（status=1 且 next_run_at <= now），逐个异步执行。
// NextRunAt 为唯一调度依据；重复触发由 AutomationTask 内的 SetNX 执行锁兜底。
func (service *ChatService) scanAutomationTasks() {
	defer func() {
		if r := recover(); r != nil {
			freedom.Logger().Errorf("scanAutomationTasks panic: %v\n%s", r, debug.Stack())
		}
	}()

	tasks, err := service.AutomationTaskRepo.ListDue(time.Now(), automationScanLimit)
	if err != nil {
		freedom.Logger().Errorf("scanAutomationTasks ListDue err: %v", err)
		return
	}
	for i := range tasks {
		task := &tasks[i]
		startedAt := time.Now()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					freedom.Logger().Errorf("AutomationTask %d panic: %v\n%s", task.Id, r, debug.Stack())
				}
			}()
			service.AutomationTask(task, startedAt)
		}()
	}
}

// buildAutomationSession 依据任务暂存字段组装自动化会话；
// PersonaId>0 走角色模式（MCP/技能读 persona_tool 表，模型以角色配置为准），
// 否则自由模式，McpIds/SkillIds 快照写入 session。AutomationTaskId 供侧边栏过滤。
func (service *ChatService) buildAutomationSession(task *po.AutomationTask) (
	session *po.Session, persona *po.UserPersona, mcpIds []int, skillIds []int, userRole string, err error,
) {
	session = &po.Session{
		UserId:           task.UserId,
		Title:            truncateRunes(task.Title, 30),
		AIModelId:        task.AIModelId,
		LastChatTime:     time.Now(),
		Project:          task.Project,
		SharedProjectId:  task.SharedProjectId,
		AutomationTaskId: task.Id,
	}

	if task.PersonaId > 0 {
		persona, mcpIds, skillIds, userRole, err = service.loadPersonaConfig(task.PersonaId, task.UserId)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		session.PersonaId = task.PersonaId
		session.AIModelId = persona.AIModelId
		// 角色模式下 mcpIds/skillIds 不存 session（读 persona_tool 表）
		return
	}

	// 自由模式：快照写入 session
	mcpIds = parseJSONInts(task.McpIds)
	skillIds = parseJSONInts(task.SkillIds)
	if len(mcpIds) > 0 {
		mcpJSON, _ := json.Marshal(mcpIds)
		session.McpIds = string(mcpJSON)
	}
	if len(skillIds) > 0 {
		skillJSON, _ := json.Marshal(skillIds)
		session.SkillIds = string(skillJSON)
	}
	return
}

// AutomationTask 静默后台执行自动化任务：依据任务暂存字段创建会话，
// 以 Requirement 作为首条消息驱动 Agent 运行。非 Streaming 模式，不给前端发 SSE 事件。
// 开始时用 SetNX 抢执行锁（TTL 5 分钟）：抢锁失败说明其他程序在执行，直接跳过；
// 锁在 OnComplete（Agent 执行结束）时释放，启动失败等未触发 OnComplete 的场景随 TTL 自然过期，
// 防重复由占位推演/MarkDone 保证。
func (service *ChatService) AutomationTask(task *po.AutomationTask, startedAt time.Time) {
	// 1. 执行锁：SetNX 5 分钟，失败说明其他程序在执行，直接跳过
	locked, lockErr := service.AutomationTaskRepo.LockTask(task.Id)
	if lockErr != nil {
		freedom.Logger().Errorf("AutomationTask %d LockTask err: %v", task.Id, lockErr)
		return
	}
	if !locked {
		return
	}

	// 2. 单次任务：加锁成功立即置为已完成，防止任何重复触发；失败不再重试
	if task.ExecType == po.AutomationExecTypeOnce {
		if err := service.AutomationTaskRepo.MarkDone(task.Id); err != nil {
			freedom.Logger().Errorf("AutomationTask %d MarkDone err: %v", task.Id, err)
			return
		}
	} else {
		// 3. 周期任务：先占位推演下一次执行时间（防执行耗时长被重复扫描），再执行
		next, nextErr := repository.CalcNextRunAt(task, startedAt)
		if nextErr != nil {
			freedom.Logger().Errorf("AutomationTask %d CalcNextRunAt err: %v", task.Id, nextErr)
			return
		}
		if err := service.AutomationTaskRepo.UpdateNextRunAt(task.Id, next); err != nil {
			freedom.Logger().Errorf("AutomationTask %d UpdateNextRunAt err: %v", task.Id, err)
			return
		}
	}

	// 4. 依据暂存字段组装会话
	session, _, mcpIds, skillIds, userRole, buildErr := service.buildAutomationSession(task)
	if buildErr != nil {
		if errors.Is(buildErr, errs.ErrPersonaNotFound) {
			// 角色已被删除，任务无法再执行，直接停用，避免周期任务反复报错
			if err := service.AutomationTaskRepo.UpdateStatus(task.Id, task.UserId, po.AutomationStatusDisabled, nil); err != nil {
				freedom.Logger().Errorf("AutomationTask %d disable err: %v", task.Id, err)
			}
			freedom.Logger().Warnf("AutomationTask %d disabled: persona %d not found", task.Id, task.PersonaId)
			return
		}
		freedom.Logger().Errorf("AutomationTask %d build session err: %v", task.Id, buildErr)
		return
	}

	// 5. 用户信息与每日 Token 限额预检（超限静默跳过，等下个周期）
	user, dailyTokenLimit, dailyTokenUsed, quotaErr := service.dailyTokenQuota(task.UserId)
	if quotaErr != nil {
		freedom.Logger().Errorf("AutomationTask %d find user err: %v", task.Id, quotaErr)
		return
	}
	if dailyTokenLimit > 0 && dailyTokenUsed >= dailyTokenLimit*1_000_000 {
		freedom.Logger().Warnf("AutomationTask %d skipped: user %s daily token limit exceeded", task.Id, task.UserId)
		return
	}

	// 6. 创建聊天模型（AIModelId=0 走默认模型池）
	chatModel, _, aiModelId, modelErr := service.ModelRepo.CreateChatModelFromID(session.AIModelId, false)
	if modelErr != nil {
		freedom.Logger().Errorf("AutomationTask %d create chat model err: %v", task.Id, modelErr)
		return
	}
	session.AIModelId = aiModelId

	// 7. 系统配置
	sysCfg, cfgErr := service.SysSettingRepo.LoadConfig()
	if cfgErr != nil {
		freedom.Logger().Errorf("AutomationTask %d load sys config err: %v", task.Id, cfgErr)
		return
	}

	// 8. MCP 工具与技能过滤器（合入始终启用项，与 StartChat 规则一致）
	alwaysOnMcpIds, mcpErr := service.McpRepo.FindAlwaysOnMcpIds()
	if mcpErr != nil {
		freedom.Logger().Errorf("AutomationTask %d find always-on mcps err: %v", task.Id, mcpErr)
		return
	}
	effectiveMcpIds := util.DedupInts(append(mcpIds, alwaysOnMcpIds...))
	mcpObjects, mcpNames, mcpErr := service.buildMCPTools(task.UserId, effectiveMcpIds)
	if mcpErr != nil {
		freedom.Logger().Errorf("AutomationTask %d build mcp tools err: %v", task.Id, mcpErr)
		return
	}
	skillNames, skillErr := service.SkillRepo.GetUserSkillNamesByIDs(task.UserId, skillIds)
	if skillErr != nil {
		freedom.Logger().Errorf("AutomationTask %d get skill names err: %v", task.Id, skillErr)
		return
	}
	alwaysOnSkillNames, skillErr := service.SkillRepo.FindAlwaysOnSkillNames(task.UserId)
	if skillErr != nil {
		freedom.Logger().Errorf("AutomationTask %d find always-on skills err: %v", task.Id, skillErr)
		return
	}
	skillNames = util.DedupStrings(append(skillNames, alwaysOnSkillNames...))

	// 9. 解析项目并对团队项目加锁（同一时刻只能有一个 Agent 操作该项目，忙时静默跳过）
	resolvedProject, resolvedProjectWorkspace, teamProjectInfo, projErr := service.resolveProjectFromSession(session)
	if projErr != nil {
		freedom.Logger().Errorf("AutomationTask %d resolve project err: %v", task.Id, projErr)
		return
	}
	spId, spLocked, lockErr := service.lockTeamProject(session, teamProjectInfo)
	if lockErr != nil {
		freedom.Logger().Warnf("AutomationTask %d lock team project err: %v", task.Id, lockErr)
		return
	}

	// 10. 持久化会话（含 AutomationTaskId，侧边栏据此过滤）
	if err := service.MsgSessionRepo.SaveSession(session); err != nil {
		freedom.Logger().Errorf("AutomationTask %d save session err: %v", task.Id, err)
		if spLocked {
			service.SharedProjectRepo.UnlockTeamProject(spId)
		}
		return
	}

	// 11. 构建 MainAgent
	param := agent.AgentParam{
		Session:             session,
		MsgRepo:             service.MsgSessionRepo,
		ChatModel:           chatModel,
		UserRole:            userRole,
		SystemSkillProvider: service.SkillRepo,
		SysCfg:              sysCfg,
		DailyStatsRepo:      service.DailyStatsRepo,
		Project:             resolvedProject,
		ProjectWorkspace:    resolvedProjectWorkspace,
		UserName:            user.Username,
		DailyTokenLimit:     dailyTokenLimit,
		DailyTokenUsed:      dailyTokenUsed,
		AutomationTaskRepo:  service.AutomationTaskRepo,
		TaskMcpIds:          mcpIds,
		TaskSkillIds:        skillIds,
	}
	mainAgent, agentErr := agent.NewMainAgent(param)
	if agentErr != nil {
		freedom.Logger().Errorf("AutomationTask %d create main agent err: %v", task.Id, agentErr)
		if spLocked {
			service.SharedProjectRepo.UnlockTeamProject(spId)
		}
		return
	}
	for _, mcpObj := range mcpObjects {
		mainAgent.AddMCP(mcpObj)
	}
	mainAgent.SetMCPFilter(agent.NewSimpleMCPFilter(mcpNames))
	simpleSkillFilter := agent.NewSimpleSkillFilter(skillNames, func(name string) {
		service.DailyStatsRepo.AddToolDailyStats(task.UserId, "skill", name)
	})
	mainAgent.SetSkillAccessor(simpleSkillFilter)
	// 静默后台运行：关闭流式输出，走非流式消息通道
	mainAgent.SetEnableStreaming(false)

	agent.RegisterRunnerHold(session.SessionId)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				freedom.Logger().Errorf("AutomationTask %d panic: %v\n%s", task.Id, r, debug.Stack())
				agent.ClearRunnerHold(session.SessionId)
				if spLocked {
					service.SharedProjectRepo.UnlockTeamProject(spId)
				}
			}
		}()

		// 设置 Flash 模型（复用默认模型）
		flashModel, flashErr := service.ModelRepo.GetDefaultChatModel()
		if flashErr != nil {
			freedom.Logger().Warnf("GetDefaultChatModel for flash: %v", flashErr)
			flashModel = chatModel
		}
		mainAgent.SetFlashModel(flashModel)

		if sysCfg.VisualEnabled {
			if visualModel, visualVerr := service.ModelRepo.GetVisualChatModel(); visualVerr == nil && visualModel != nil {
				mainAgent.SetVisualModel(visualModel)
			}
		}

		ctx := context.Background()
		runner, runErr := mainAgent.NewRunner(ctx)
		if runErr != nil {
			freedom.Logger().Errorf("AutomationTask %d create runner err: %v", task.Id, runErr)
			agent.ClearRunnerHold(session.SessionId)
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			return
		}

		runner.OnComplete(func(event *agent.RunnerCompleteEvent) {
			// Agent 执行结束（无论成败），释放任务锁
			service.releaseAutomationTaskLock(task.Id)
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			finishedAt := time.Now()
			// 执行成功（会话完成）才写入执行记录；失败不写，原因在 session 与聊天记录中自然可见
			if event.Terminated || event.Err != "" {
				freedom.Logger().Warnf("AutomationTask %d failed: sessionId=%s terminated=%v err=%s",
					task.Id, session.SessionId, event.Terminated, event.Err)
			} else if execErr := service.AutomationExecRepo.CreateExecution(&po.AutomationExecution{
				AutomationTaskId: task.Id,
				SessionId:        session.SessionId,
				StartedAt:        startedAt,
				FinishedAt:       finishedAt,
			}); execErr != nil {
				freedom.Logger().Errorf("AutomationTask %d create execution err: %v", task.Id, execErr)
			}
			// 间隔类型：以实际完成时间为基准精确推演（fixed-delay 语义，间隔从跑完才开始计）
			if task.ExecType == po.AutomationExecTypeInterval {
				next := finishedAt.Add(time.Duration(task.IntervalMinutes) * time.Minute)
				if nextErr := service.AutomationTaskRepo.UpdateNextRunAt(task.Id, next); nextErr != nil {
					freedom.Logger().Errorf("AutomationTask %d refine next run err: %v", task.Id, nextErr)
				}
			}
		})

		// 以任务需求作为首条消息，后台静默执行
		if queryErr := runner.Query(ctx, task.Requirement); queryErr != nil {
			freedom.Logger().Errorf("AutomationTask %d query err: %v", task.Id, queryErr)
			agent.ClearRunnerHold(session.SessionId)
			if spLocked {
				service.SharedProjectRepo.UnlockTeamProject(spId)
			}
			return
		}

		// 注册运行器（静默会话无 SSE 消费方，仅保持与 StartChat 一致的生命周期管理）
		agent.RegisterRunner(session.SessionId, runner)
	}()
}

// releaseAutomationTaskLock 释放自动化任务执行锁
func (service *ChatService) releaseAutomationTaskLock(taskId int) {
	if err := service.AutomationTaskRepo.UnlockTask(taskId); err != nil {
		freedom.Logger().Warnf("AutomationTask %d UnlockTask err: %v", taskId, err)
	}
}

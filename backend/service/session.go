package service

import (
	"encoding/json"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/agent"
	"slices"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *SessionService {
			return &SessionService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *SessionService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// SessionService 用户会话服务
type SessionService struct {
	Worker            freedom.Worker
	MsgSessionRepo    *repository.MsgSessionRepository
	ModelRepo         *repository.ProviderRepository
	PersonaRepo       *repository.PersonaRepository
	SharedProjectRepo *repository.TeamProjectRepository
}

// ListSessions 获取用户会话列表（侧边栏"所有对话"），分页
func (service *SessionService) ListSessions(userId string, req *vo.SessionListReq) (*infra.PageResponse, error) {
	sessions, pageResult, err := service.MsgSessionRepo.ListSessions(userId, req)
	if err != nil {
		return nil, err
	}

	sharedMap := service.resolveSharedProjectsBatch(sessions)

	items := make([]vo.SessionListItem, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, vo.SessionListItem{
			SessionId:     s.SessionId,
			Title:         s.Title,
			Status:        s.Status,
			PersonaId:     s.PersonaId,
			Project:       s.Project,
			TeamProject:   sharedMap[s.SessionId],
			LastChatTime:  s.LastChatTime,
			Created:       s.Created,
		})
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pageResult.TotalPage,
		TotalCount: pageResult.TotalCount,
		Page:       pageResult.Page,
		PageSize:   pageResult.PageSize,
	}, nil
}

// GetSession 获取会话详情（含 token 统计、模型、MCP/技能快照）
func (service *SessionService) GetSession(sessionId string, userId string) (*vo.SessionDetailRsp, error) {
	session, err := service.MsgSessionRepo.GetUserSession(sessionId, userId)
	if err != nil {
		return nil, errs.ErrSessionNotFound
	}

	// Runner-lost 安全检查：如果 session 标记为进行中（status=1）但内存中已无
	// 对应 Runner（例如服务器重启、进程崩溃），自动重置 status 为 0，避免前端
	// 轮询永远等不到完成信号。
	if session.Status == 1 {
		if _, ok := agent.GetRunner(session.SessionId); !ok {
			_ = service.MsgSessionRepo.UpdateSessionStatus(session.SessionId, 0)
			session.Status = 0
		}
	}

	_, _, sharedInfo, _ := service.resolveProjectFromSession(session)

	rsp := &vo.SessionDetailRsp{
		SessionId:             session.SessionId,
		Title:                 session.Title,
		Status:                session.Status,
		PersonaId:             session.PersonaId,
		Project:               session.Project,
		TeamProject:           sharedInfo,
		AIModelId:             session.AIModelId,
		ContextTokens:         session.ContextTokens,
		PromptTokensCount:     session.PromptTokensCount,
		CompletionTokensCount: session.CompletionTokensCount,
		McpIds:                parseJSONInts(session.McpIds),
		SkillIds:              parseJSONInts(session.SkillIds),
		LastChatTime:          session.LastChatTime,
		Created:               session.Created,
	}

	// 关联模型名称和上下文长度
	if session.AIModelId > 0 {
		if model, err := service.ModelRepo.GetModelByID(session.AIModelId); err == nil {
			rsp.ModelName = model.DisplayName
			rsp.ModelIcon = model.Icon
			rsp.ContextLimit = model.ContextLenInTokens()
		}
	}

	// 关联角色名称和图标
	if session.PersonaId > 0 {
		if persona, err := service.PersonaRepo.GetUserPersonaByID(session.PersonaId, userId); err == nil {
			rsp.PersonaName = persona.Name
			rsp.PersonaIcon = persona.Icon
		}
	}

	return rsp, nil
}

// GetMessages 获取会话历史消息
func (service *SessionService) GetMessages(sessionId string, userId string) ([]vo.MessageItem, error) {
	// 校验会话属于当前用户
	if _, err := service.MsgSessionRepo.GetUserSession(sessionId, userId); err != nil {
		return nil, errs.ErrSessionNotFound
	}

	messages, err := service.MsgSessionRepo.GetAllMessages(sessionId)
	if err != nil {
		return nil, err
	}

	items := make([]vo.MessageItem, 0, len(messages))
	for _, m := range messages {
		reasoningContent := service.buildReasoningContent(m)
		items = append(items, vo.MessageItem{
			MsgId:            m.MsgId,
			RoundId:          m.RoundId,
			ContextState:     m.ContextState,
			Content:          m.Content,
			ReasoningContent: reasoningContent,
			RoleType:         m.RoleType,
			Created:          m.Created.Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}

// DeleteSession 删除会话（软删除）
func (service *SessionService) DeleteSession(sessionId string, userId string) error {
	return service.MsgSessionRepo.SoftDeleteSession(sessionId, userId)
}

// UpdateSession 更新会话（标题、归档等，只更新传入字段）
func (service *SessionService) UpdateSession(sessionId string, userId string, req *vo.SessionUpdateReq) error {
	// 校验会话属于当前用户
	if _, err := service.MsgSessionRepo.GetUserSession(sessionId, userId); err != nil {
		return errs.ErrSessionNotFound
	}

	updates := map[string]interface{}{"updated": time.Now()}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}
	return service.MsgSessionRepo.UpdateSession(sessionId, userId, updates)
}

// parseJSONInts 将 JSON 数组字符串解析为 []int，空字符串返回 nil
func parseJSONInts(s string) []int {
	var ids []int
	if s == "" {
		return ids
	}

	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

// buildReasoningContent 构建当前消息的完整思考内容（自己的 + 同轮其他 tool 消息的思考）
func (service *SessionService) buildReasoningContent(m po.Message) []vo.MessageReasoningItem {
	result := make([]vo.MessageReasoningItem, 0)
	if m.ReasoningContent == "" {
		return result
	}

	list, _ := service.MsgSessionRepo.GetToolReasoningContent(m.SessionId, m.RoundId, m.MsgId)
	slices.Reverse(list)
	for _, v := range list {
		result = append(result, vo.MessageReasoningItem{
			EventType: agent.SSEEventTypeReasoning,
			Content:   v.ReasoningContent,
		})

		if v.ToolCallsInfo == "" {
			continue
		}

		toolinfo := vo.MessageReasoningItem{
			EventType: agent.SSEEventTypeTool,
			Tool:      &vo.MessageReasoningToolCallsInfo{},
		}

		if err := json.Unmarshal([]byte(v.ToolCallsInfo), toolinfo.Tool); err == nil {
			result = append(result, toolinfo)
		}
	}

	result = append(result, vo.MessageReasoningItem{EventType: agent.SSEEventTypeReasoning, Content: m.ReasoningContent})
	return result
}

// resolveProjectFromSession 从 session 解析项目信息
func (service *SessionService) resolveProjectFromSession(session *po.Session) (projectName, projectWorkspace string, teamInfo *vo.TeamProjectInfo, err error) {
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

// resolveSharedProjectsBatch 批量解析多个 session 的团队项目信息，避免 N+1 查询。
func (service *SessionService) resolveSharedProjectsBatch(sessions []po.Session) map[string]*vo.TeamProjectInfo {
	spIds := make([]int, 0, len(sessions))
	for i := range sessions {
		if sessions[i].SharedProjectId > 0 {
			spIds = append(spIds, sessions[i].SharedProjectId)
		}
	}
	if len(spIds) == 0 {
		return map[string]*vo.TeamProjectInfo{}
	}
	spMap, err := service.SharedProjectRepo.ListByIDs(spIds)
	if err != nil {
		return map[string]*vo.TeamProjectInfo{}
	}
	creatorIds := make([]string, 0, len(spMap))
	seen := map[string]bool{}
	for i := range spMap {
		cid := spMap[i].CreatorId
		if !seen[cid] {
			seen[cid] = true
			creatorIds = append(creatorIds, cid)
		}
	}
	userMap, _ := service.SharedProjectRepo.GetUsersByIDs(creatorIds)

	result := make(map[string]*vo.TeamProjectInfo, len(spIds))
	for i := range sessions {
		s := &sessions[i]
		if s.SharedProjectId <= 0 {
			continue
		}
		tp, ok := spMap[s.SharedProjectId]
		if !ok {
			continue
		}
		creatorName := ""
		if u, ok := userMap[tp.CreatorId]; ok {
			creatorName = u.Nickname
			if creatorName == "" {
				creatorName = u.Username
			}
		}
		result[s.SessionId] = &vo.TeamProjectInfo{
			Id:          tp.Id,
			CreatorId:   tp.CreatorId,
			CreatorName: creatorName,
			ProjectName: tp.ProjectName,
			Description: tp.Description,
		}
	}
	return result
}

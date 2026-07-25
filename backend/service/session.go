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

type SessionService struct {
	Worker            freedom.Worker
	MsgSessionRepo    *repository.MsgSessionRepository
	ModelRepo         *repository.ProviderRepository
	PersonaRepo       *repository.PersonaRepository
	SharedProjectRepo *repository.SharedProjectRepository
}

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
			SharedProject: sharedMap[s.SessionId],
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

func (service *SessionService) GetSession(sessionId string, userId string) (*vo.SessionDetailRsp, error) {
	session, err := service.MsgSessionRepo.GetUserSession(sessionId, userId)
	if err != nil {
		return nil, errs.ErrSessionNotFound
	}

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
		SharedProject:         sharedInfo,
		AIModelId:             session.AIModelId,
		ContextTokens:         session.ContextTokens,
		PromptTokensCount:     session.PromptTokensCount,
		CompletionTokensCount: session.CompletionTokensCount,
		McpIds:                parseJSONInts(session.McpIds),
		SkillIds:              parseJSONInts(session.SkillIds),
		LastChatTime:          session.LastChatTime,
		Created:               session.Created,
	}

	if session.AIModelId > 0 {
		if model, err := service.ModelRepo.GetModelByID(session.AIModelId); err == nil {
			rsp.ModelName = model.DisplayName
			rsp.ContextLimit = model.ContextLenInTokens()
		}
	}

	if session.PersonaId > 0 {
		if persona, err := service.PersonaRepo.GetUserPersonaByID(session.PersonaId, userId); err == nil {
			rsp.PersonaName = persona.Name
			rsp.PersonaIcon = persona.Icon
		}
	}

	return rsp, nil
}

func (service *SessionService) GetMessages(sessionId string, userId string) ([]vo.MessageItem, error) {

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

func (service *SessionService) DeleteSession(sessionId string, userId string) error {
	return service.MsgSessionRepo.SoftDeleteSession(sessionId, userId)
}

func (service *SessionService) UpdateSession(sessionId string, userId string, req *vo.SessionUpdateReq) error {

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

func (service *SessionService) resolveProjectFromSession(session *po.Session) (projectName, projectWorkspace string, sharedInfo *vo.SharedProjectInfo, err error) {
	if session.SharedProjectId <= 0 {
		return session.Project, "", nil, nil
	}
	sp, dbErr := service.SharedProjectRepo.GetByID(session.SharedProjectId)
	if dbErr != nil {
		return session.Project, "", nil, nil
	}
	owner, userErr := service.SharedProjectRepo.GetUserByID(sp.OwnerId)
	if userErr != nil {
		return sp.ProjectName, "", nil, nil
	}
	ownerName := owner.Nickname
	if ownerName == "" {
		ownerName = owner.Username
	}
	projectWorkspace = config.Get().GetUserSpace(owner.Username)
	sharedInfo = &vo.SharedProjectInfo{
		Id:          sp.Id,
		OwnerId:     sp.OwnerId,
		OwnerName:   ownerName,
		ProjectName: sp.ProjectName,
		Description: sp.Description,
	}
	return sp.ProjectName, projectWorkspace, sharedInfo, nil
}

func (service *SessionService) resolveSharedProjectsBatch(sessions []po.Session) map[string]*vo.SharedProjectInfo {
	spIds := make([]int, 0, len(sessions))
	for i := range sessions {
		if sessions[i].SharedProjectId > 0 {
			spIds = append(spIds, sessions[i].SharedProjectId)
		}
	}
	if len(spIds) == 0 {
		return map[string]*vo.SharedProjectInfo{}
	}
	spMap, err := service.SharedProjectRepo.ListByIDs(spIds)
	if err != nil {
		return map[string]*vo.SharedProjectInfo{}
	}
	ownerIds := make([]string, 0, len(spMap))
	seen := map[string]bool{}
	for i := range spMap {
		oid := spMap[i].OwnerId
		if !seen[oid] {
			seen[oid] = true
			ownerIds = append(ownerIds, oid)
		}
	}
	userMap, _ := service.SharedProjectRepo.GetUsersByIDs(ownerIds)

	result := make(map[string]*vo.SharedProjectInfo, len(spIds))
	for i := range sessions {
		s := &sessions[i]
		if s.SharedProjectId <= 0 {
			continue
		}
		sp, ok := spMap[s.SharedProjectId]
		if !ok {
			continue
		}
		ownerName := ""
		if u, ok := userMap[sp.OwnerId]; ok {
			ownerName = u.Nickname
			if ownerName == "" {
				ownerName = u.Username
			}
		}
		result[s.SessionId] = &vo.SharedProjectInfo{
			Id:          sp.Id,
			OwnerId:     sp.OwnerId,
			OwnerName:   ownerName,
			ProjectName: sp.ProjectName,
			Description: sp.Description,
		}
	}
	return result
}

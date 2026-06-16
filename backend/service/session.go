package service

import (
	"encoding/json"
	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/backend/vo/errs"
	"raven/core/agent"
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
	Worker		freedom.Worker
	MsgSessionRepo	*repository.MsgSessionRepository
	ModelRepo	*repository.ProviderRepository
	PersonaRepo	*repository.PersonaRepository
}

func (service *SessionService) ListSessions(userId string, req *vo.SessionListReq) (*infra.PageResponse, error) {
	sessions, pageResult, err := service.MsgSessionRepo.ListSessions(userId, req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.SessionListItem, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, vo.SessionListItem{
			SessionId:	s.SessionId,
			Title:		s.Title,
			Status:		s.Status,
			PersonaId:	s.PersonaId,
			Project:	s.Project,
			LastChatTime:	s.LastChatTime,
			Created:	s.Created,
		})
	}
	return &infra.PageResponse{
		List:		items,
		TotalPage:	pageResult.TotalPage,
		TotalCount:	pageResult.TotalCount,
		Page:		pageResult.Page,
		PageSize:	pageResult.PageSize,
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

	rsp := &vo.SessionDetailRsp{
		SessionId:		session.SessionId,
		Title:			session.Title,
		Status:			session.Status,
		PersonaId:		session.PersonaId,
		Project:		session.Project,
		AIModelId:		session.AIModelId,
		ContextTokens:		session.ContextTokens,
		PromptTokensCount:	session.PromptTokensCount,
		CompletionTokensCount:	session.CompletionTokensCount,
		McpIds:			parseJSONInts(session.McpIds),
		SkillIds:		parseJSONInts(session.SkillIds),
		LastChatTime:		session.LastChatTime,
		Created:		session.Created,
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
			MsgId:			m.MsgId,
			RoundId:		m.RoundId,
			ContextState:		m.ContextState,
			Content:		m.Content,
			ReasoningContent:	reasoningContent,
			RoleType:		m.RoleType,
			Created:		m.Created.Format("2006-01-02 15:04:05"),
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
			EventType:	agent.SSEEventTypeReasoning,
			Content:	v.ReasoningContent,
		})

		if v.ToolCallsInfo == "" {
			continue
		}

		toolinfo := vo.MessageReasoningItem{
			EventType:	agent.SSEEventTypeTool,
			Tool:		&vo.MessageReasoningToolCallsInfo{},
		}

		if err := json.Unmarshal([]byte(v.ToolCallsInfo), toolinfo.Tool); err == nil {
			result = append(result, toolinfo)
		}
	}

	result = append(result, vo.MessageReasoningItem{EventType: agent.SSEEventTypeReasoning, Content: m.ReasoningContent})
	return result
}

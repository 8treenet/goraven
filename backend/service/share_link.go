package service

import (
	"raven/backend/infra"
	"raven/backend/po"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/backend/vo/errs"
	"raven/util"
	"strings"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *ShareLinkService {
			return &ShareLinkService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *ShareLinkService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type ShareLinkService struct {
	SessionService
	ShareLinkRepo	*repository.ShareLinkRepository
	UserRepo	*repository.UserRepository
}

var expiresInMapping = map[string]time.Duration{
	"1h":	time.Hour,
	"24h":	24 * time.Hour,
	"7d":	7 * 24 * time.Hour,
	"30d":	30 * 24 * time.Hour,
}

func (service *ShareLinkService) buildShareTitle(title string, session *po.Session) string {
	if strings.TrimSpace(title) != "" {
		title = strings.TrimSpace(title)
		if len([]rune(title)) > 100 {
			title = string([]rune(title)[:100])
		}
		return title
	}

	if strings.TrimSpace(session.Title) != "" {
		return session.Title
	}

	firstMsg, err := service.MsgSessionRepo.GetFirstUserMessage(session.SessionId)
	if err == nil && firstMsg.Content != "" {
		content := firstMsg.Content
		runes := []rune(content)
		if len(runes) > 30 {
			content = string(runes[:30])
		}
		return content
	}

	return "Raven 对话分享"
}

func (service *ShareLinkService) CreateShare(userId string, sessionId string, req *vo.CreateShareReq) (*vo.ShareLinkRsp, error) {

	session, err := service.MsgSessionRepo.GetUserSession(sessionId, userId)
	if err != nil {
		return nil, errs.ErrShareSessionNotFound
	}

	existing, err := service.ShareLinkRepo.GetSessionShare(sessionId, userId)
	if err == nil && existing != nil && !existing.IsExpired() {
		return &vo.ShareLinkRsp{
			ShareId:	existing.ShareId,
			SessionId:	existing.SessionId,
			Title:		existing.Title,
			ExpiresAt:	existing.ExpiresAt,
			ViewCount:	existing.ViewCount,
			IsExpired:	existing.IsExpired(),
			Created:	existing.Created,
		}, errs.ErrShareAlreadyExists
	}

	expiresIn := req.ExpiresIn
	if expiresIn == "" {
		expiresIn = "24h"
	}
	duration, ok := expiresInMapping[expiresIn]
	if !ok {
		duration = 24 * time.Hour
	}

	title := service.buildShareTitle(req.Title, session)

	shareLink := &po.ShareLink{
		ShareId:	util.UUID(),
		UserId:		userId,
		SessionId:	sessionId,
		Title:		title,
		ExpiresAt:	time.Now().Add(duration),
	}
	if err := service.ShareLinkRepo.CreateShareLink(shareLink); err != nil {
		return nil, err
	}

	return &vo.ShareLinkRsp{
		ShareId:	shareLink.ShareId,
		SessionId:	shareLink.SessionId,
		Title:		shareLink.Title,
		ExpiresAt:	shareLink.ExpiresAt,
		ViewCount:	shareLink.ViewCount,
		IsExpired:	false,
		Created:	shareLink.Created,
	}, nil
}

func (service *ShareLinkService) GetSessionShare(sessionId string, userId string) (*vo.ShareLinkRsp, error) {
	shareLink, err := service.ShareLinkRepo.GetSessionShare(sessionId, userId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}
	return &vo.ShareLinkRsp{
		ShareId:	shareLink.ShareId,
		SessionId:	shareLink.SessionId,
		Title:		shareLink.Title,
		ExpiresAt:	shareLink.ExpiresAt,
		ViewCount:	shareLink.ViewCount,
		IsExpired:	shareLink.IsExpired(),
		Created:	shareLink.Created,
	}, nil
}

func (service *ShareLinkService) DeleteShare(sessionId string, userId string) error {
	return service.ShareLinkRepo.DeleteShareLink(sessionId, userId)
}

func (service *ShareLinkService) ListUserShares(userId string, req *vo.UserShareListReq) (*infra.PageResponse, error) {
	shareLinks, pr, err := service.ShareLinkRepo.ListUserShareLinks(userId, req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserShareListItem, 0, len(shareLinks))
	for _, s := range shareLinks {
		items = append(items, vo.UserShareListItem{
			ShareId:	s.ShareId,
			SessionId:	s.SessionId,
			Title:		s.Title,
			ExpiresAt:	s.ExpiresAt,
			ViewCount:	s.ViewCount,
			IsExpired:	s.IsExpired(),
			Created:	s.Created,
		})
	}
	return &infra.PageResponse{
		List:		items,
		TotalPage:	pr.TotalPage,
		TotalCount:	pr.TotalCount,
		Page:		pr.Page,
		PageSize:	pr.PageSize,
	}, nil
}

func (service *ShareLinkService) GetPublicShare(shareId string) (*vo.PublicShareRsp, error) {
	shareLink, err := service.ShareLinkRepo.GetShareLink(shareId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}

	if shareLink.IsExpired() {
		return &vo.PublicShareRsp{
			ShareId:	shareLink.ShareId,
			Title:		shareLink.Title,
			ExpiresAt:	shareLink.ExpiresAt,
			ViewCount:	shareLink.ViewCount,
			IsExpired:	true,
		}, errs.ErrShareExpired
	}

	creator := ""
	user, err := service.UserRepo.FindByUserId(shareLink.UserId)
	if err == nil {
		creator = user.Nickname
		if creator == "" {
			creator = user.Username
		}
	}

	go func() {
		_ = service.ShareLinkRepo.IncrementViewCount(shareId)
	}()

	messages, err := service.MsgSessionRepo.GetAllMessages(shareLink.SessionId)
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

	return &vo.PublicShareRsp{
		ShareId:	shareLink.ShareId,
		Title:		shareLink.Title,
		Creator:	creator,
		Created:	shareLink.Created,
		ExpiresAt:	shareLink.ExpiresAt,
		ViewCount:	shareLink.ViewCount + 1,
		IsExpired:	false,
		Messages:	items,
	}, nil
}

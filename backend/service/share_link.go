package service

import (
	"fmt"
	"html"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/iris/v12"
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

// ShareLinkService 对话分享链接服务
type ShareLinkService struct {
	SessionService
	ShareLinkRepo *repository.ShareLinkRepository
	UserRepo      *repository.UserRepository
	SysCfgRepo    *repository.SystemSettingRepository
}

// expiresInMapping 过期时间选项到 Duration 的映射
var expiresInMapping = map[string]time.Duration{
	"1h":  time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

// buildShareTitle 生成分享标题
// 优先使用传入的 title，其次取 session.title，最后取用户第一条消息前 30 字
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

	// 取用户第一条消息前 30 字
	firstMsg, err := service.MsgSessionRepo.GetFirstUserMessage(session.SessionId)
	if err == nil && firstMsg.Content != "" {
		content := firstMsg.Content
		runes := []rune(content)
		if len(runes) > 30 {
			content = string(runes[:30])
		}
		return content
	}

	return "GoRaven 对话分享"
}

// CreateShare 创建分享链接
func (service *ShareLinkService) CreateShare(userId string, sessionId string, req *vo.CreateShareReq) (*vo.ShareLinkRsp, error) {
	// 校验会话归属
	session, err := service.MsgSessionRepo.GetUserSession(sessionId, userId)
	if err != nil {
		return nil, errs.ErrShareSessionNotFound
	}

	// 计算过期时间
	expiresIn := req.ExpiresIn
	if expiresIn == "" {
		expiresIn = "24h"
	}
	duration, ok := expiresInMapping[expiresIn]
	if !ok {
		duration = 24 * time.Hour
	}

	shareType := req.ShareType
	if shareType != "internal" {
		shareType = "public"
	}

	title := service.buildShareTitle(req.Title, session)

	// 如果系统域名未配置，尝试用前端传过来的域名自动补全
	service.autoConfigureDomain(req.Domain)

	// 若已有有效分享，更新它
	existing, err := service.ShareLinkRepo.GetSessionShare(sessionId, userId)
	if err == nil && existing != nil && !existing.IsExpired() {
		updated, err := service.ShareLinkRepo.UpdateShareLink(existing.ShareId, title, shareType, time.Now().Add(duration))
		if err != nil {
			return nil, err
		}
		return &vo.ShareLinkRsp{
			ShareId:   updated.ShareId,
			SessionId: updated.SessionId,
			Title:     updated.Title,
			ShareType: updated.ShareType,
			ExpiresAt: updated.ExpiresAt,
			ViewCount: updated.ViewCount,
			IsExpired: false,
			Created:   updated.Created,
		}, nil
	}

	shareLink := &po.ShareLink{
		ShareId:   util.UUID(),
		UserId:    userId,
		SessionId: sessionId,
		Title:     title,
		ShareType: shareType,
		ExpiresAt: time.Now().Add(duration),
	}
	if err := service.ShareLinkRepo.CreateShareLink(shareLink); err != nil {
		return nil, err
	}

	return &vo.ShareLinkRsp{
		ShareId:   shareLink.ShareId,
		SessionId: shareLink.SessionId,
		Title:     shareLink.Title,
		ShareType: shareLink.ShareType,
		ExpiresAt: shareLink.ExpiresAt,
		ViewCount: shareLink.ViewCount,
		IsExpired: false,
		Created:   shareLink.Created,
	}, nil
}

// GetSessionShare 获取会话的分享链接
func (service *ShareLinkService) GetSessionShare(sessionId string, userId string) (*vo.ShareLinkRsp, error) {
	shareLink, err := service.ShareLinkRepo.GetSessionShare(sessionId, userId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}
	return &vo.ShareLinkRsp{
		ShareId:   shareLink.ShareId,
		SessionId: shareLink.SessionId,
		Title:     shareLink.Title,
		ShareType: shareLink.ShareType,
		ExpiresAt: shareLink.ExpiresAt,
		ViewCount: shareLink.ViewCount,
		IsExpired: shareLink.IsExpired(),
		Created:   shareLink.Created,
	}, nil
}

// DeleteShare 删除会话的分享链接
func (service *ShareLinkService) DeleteShare(sessionId string, userId string) error {
	return service.ShareLinkRepo.DeleteShareLink(sessionId, userId)
}

// ListUserShares 获取用户分享链接分页列表
func (service *ShareLinkService) ListUserShares(userId string, req *vo.UserShareListReq) (*infra.PageResponse, error) {
	shareLinks, pr, err := service.ShareLinkRepo.ListUserShareLinks(userId, req)
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserShareListItem, 0, len(shareLinks))
	for _, s := range shareLinks {
		items = append(items, vo.UserShareListItem{
			ShareId:   s.ShareId,
			SessionId: s.SessionId,
			Title:     s.Title,
			ShareType: s.ShareType,
			ExpiresAt: s.ExpiresAt,
			ViewCount: s.ViewCount,
			IsExpired: s.IsExpired(),
			Created:   s.Created,
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

// GetShareInfo 获取分享信息（无鉴权）
func (service *ShareLinkService) GetShareInfo(shareId string) (*vo.PublicShareRsp, error) {
	shareLink, err := service.ShareLinkRepo.GetShareLink(shareId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}

	if shareLink.IsExpired() {
		return &vo.PublicShareRsp{
			ShareId:   shareLink.ShareId,
			Title:     shareLink.Title,
			ShareType: shareLink.ShareType,
			ExpiresAt: shareLink.ExpiresAt,
			ViewCount: shareLink.ViewCount,
			IsExpired: true,
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

	return &vo.PublicShareRsp{
		ShareId:   shareLink.ShareId,
		Title:     shareLink.Title,
		Creator:   creator,
		ShareType: shareLink.ShareType,
		Created:   shareLink.Created,
		ExpiresAt: shareLink.ExpiresAt,
		ViewCount: shareLink.ViewCount + 1,
		IsExpired: false,
	}, nil
}

// buildShareMessages 获取并构建分享的消息列表
func (service *ShareLinkService) buildShareMessages(shareLink *po.ShareLink) ([]vo.MessageItem, error) {
	messages, err := service.MsgSessionRepo.GetAllMessages(shareLink.SessionId)
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

// GetShareMessages 获取分享的消息
// userId 为空表示未登录；如果是内部分享且未登录，直接报错
func (service *ShareLinkService) GetShareMessages(shareId string, userId string) ([]vo.MessageItem, error) {
	shareLink, err := service.ShareLinkRepo.GetShareLink(shareId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}

	if shareLink.IsExpired() {
		return nil, errs.ErrShareExpired
	}

	if shareLink.ShareType == "internal" && userId == "" {
		return nil, errs.ErrShareNotFound
	}

	return service.buildShareMessages(shareLink)
}

// GetShareOGData 获取分享页 Open Graph 元数据（供社交平台/爬虫预览卡片）。
// 与 GetShareInfo 的区别：不增量浏览量；不返回消息正文；从传入的 ctx 取请求上下文拼绝对 URL。
func (service *ShareLinkService) GetShareOGData(shareId string, ctx iris.Context) (*vo.ShareOGData, error) {
	shareLink, err := service.ShareLinkRepo.GetShareLink(shareId)
	if err != nil {
		return nil, errs.ErrShareNotFound
	}
	if shareLink.IsExpired() {
		return nil, errs.ErrShareExpired
	}

	creator := ""
	if user, e := service.UserRepo.FindByUserId(shareLink.UserId); e == nil {
		creator = user.Nickname
		if creator == "" {
			creator = user.Username
		}
	}

	host := ctx.Host()
	var baseURL string
	if host != "" && !util.IsLocalOrIPHost(host) {
		baseURL = requestScheme(ctx) + "://" + host
	} else if sysconf, e := service.SysCfgRepo.LoadConfig(); e == nil {
		if d := strings.TrimSuffix(strings.TrimSpace(sysconf.GeneralDomain), "/"); d != "" && !util.IsLocalOrIPURL(d) {
			baseURL = d
		}
	}

	lang := config.Get().GetLanguage()
	cover := "og-cover_en.png"
	if lang == "zh" {
		cover = "og-cover_cn.png"
	}

	title := strings.TrimSpace(shareLink.Title)
	if creator == "" {
		creator = "GoRaven"
	}
	var description string
	if lang == "zh" {
		if title == "" {
			title = "GoRaven 对话分享"
		}
		description = fmt.Sprintf("%s 在 GoRaven 的 AI 对话分享 · %s", creator, shareLink.Created.Format("2006-01-02 15:04"))
	} else {
		if title == "" {
			title = "GoRaven Conversation Share"
		}
		description = fmt.Sprintf("%s's AI conversation on GoRaven · %s", creator, shareLink.Created.Format("2006-01-02 15:04"))
	}

	var imageURL, pageURL string
	if baseURL != "" {
		imageURL = baseURL + "/logos/" + cover
		pageURL = baseURL + "/share/" + shareLink.ShareId
	}

	return &vo.ShareOGData{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		PageURL:     pageURL,
	}, nil
}

// BuildOGHTML 基于 ShareOGData 生成 Open Graph / Twitter Card 的 <meta> 标签块。
// ImageURL / PageURL 为空时跳过对应标签，避免输出空值 meta（社交平台会视为无效）。
func BuildOGHTML(d *vo.ShareOGData) string {
	var b strings.Builder
	t := html.EscapeString(d.Title)
	desc := html.EscapeString(d.Description)
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\"/>\n", t)
	fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\"/>\n", desc)
	if d.ImageURL != "" {
		i := html.EscapeString(d.ImageURL)
		fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"%s\"/>\n", i)
		fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"%s\"/>\n", i)
	}
	if d.PageURL != "" {
		u := html.EscapeString(d.PageURL)
		fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"%s\"/>\n", u)
	}
	b.WriteString("<meta property=\"og:type\" content=\"website\"/>\n")
	b.WriteString("<meta property=\"og:site_name\" content=\"GoRaven\"/>\n")
	b.WriteString("<meta name=\"twitter:card\" content=\"summary_large_image\"/>\n")
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\"/>\n", t)
	fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\"/>\n", desc)
	return b.String()
}

// requestScheme 从当前请求推导协议，兼容反向代理（X-Forwarded-Proto）与直连 TLS。
func requestScheme(ctx iris.Context) string {
	if ctx.Request().TLS != nil {
		return "https"
	}
	if proto := ctx.GetHeader("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(proto)
	}
	return "http"
}

// autoConfigureDomain 若系统 GeneralDomain 为空，尝试用前端传的 domain 自动补全。
// 仅当 domain 为非 IP/非 localhost 的合法 URL 时才写入。
func (service *ShareLinkService) autoConfigureDomain(domain string) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return
	}
	sysconf, err := service.SysCfgRepo.LoadConfig()
	if err != nil {
		return
	}
	if strings.TrimSpace(sysconf.GeneralDomain) != "" {
		return
	}
	if util.IsLocalOrIPURL(domain) {
		return
	}
	_ = service.SysCfgRepo.Update("general.domain", domain)
}

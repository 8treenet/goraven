package vo

import "time"

// CreateShareReq 创建分享链接请求
type CreateShareReq struct {
	Title     string `json:"title"`     // 分享标题，为空时取 session.title
	ExpiresIn string `json:"expiresIn"` // 过期时间选项：1h, 24h, 7d, 30d
	ShareType string `json:"shareType"` // 分享类型：public(公开) / internal(内部)，默认 public
	Domain    string `json:"domain"`    // 前端页面域名（window.location.origin），用于后端自动补全 GeneralDomain
}

// ShareLinkRsp 分享链接响应
type ShareLinkRsp struct {
	ShareId   string    `json:"shareId"`   // 分享唯一标识
	SessionId string    `json:"sessionId"` // 被分享的会话ID
	Title     string    `json:"title"`     // 分享标题
	ShareType string    `json:"shareType"` // 分享类型：public(公开) / internal(内部)
	ExpiresAt time.Time `json:"expiresAt"` // 过期时间
	ViewCount int       `json:"viewCount"` // 浏览次数
	IsExpired bool      `json:"isExpired"` // 是否已过期
	Created   time.Time `json:"created"`   // 创建时间
}

// UserShareListItem 用户分享列表条目
type UserShareListItem struct {
	ShareId    string    `json:"shareId"`    // 分享唯一标识
	SessionId  string    `json:"sessionId"`  // 被分享的会话ID
	Title      string    `json:"title"`      // 分享标题
	ShareType  string    `json:"shareType"`  // 分享类型：public(公开) / internal(内部)
	ExpiresAt  time.Time `json:"expiresAt"`  // 过期时间
	ViewCount  int       `json:"viewCount"`  // 浏览次数
	IsExpired  bool      `json:"isExpired"`  // 是否已过期
	Created    time.Time `json:"created"`    // 创建时间
}

// UserShareListReq 用户分享列表请求
type UserShareListReq struct {
	Page     int `url:"page"`     // 页码
	PageSize int `url:"pageSize"` // 每页条数
}

// PublicShareRsp 分享信息响应（公开和内部共用）
type PublicShareRsp struct {
	ShareId   string    `json:"shareId"`   // 分享唯一标识
	Title     string    `json:"title"`     // 分享标题
	Creator   string    `json:"creator"`   // 创建者显示名（Nickname 或 Username）
	ShareType string    `json:"shareType"` // 分享类型：public(公开) / internal(内部)
	Created   time.Time `json:"created"`   // 创建时间
	ExpiresAt time.Time `json:"expiresAt"` // 过期时间
	ViewCount int       `json:"viewCount"` // 浏览次数
	IsExpired bool      `json:"isExpired"` // 是否已过期
}

// ShareOGData 分享页 Open Graph 元数据，供爬虫/社交平台预览卡片使用。
// 由 ShareLinkService.GetShareOGData 产出，不包含会话正文，也不计入浏览量。
type ShareOGData struct {
	Title       string // og:title 与 twitter:title
	Description string // og:description 与 twitter:description
	ImageURL    string // og:image 与 twitter:image，绝对 URL
	PageURL     string // og:url，绝对 URL
}

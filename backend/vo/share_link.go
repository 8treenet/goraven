package vo

import "time"

type CreateShareReq struct {
	Title     string `json:"title"`
	ExpiresIn string `json:"expiresIn"`
	ShareType string `json:"shareType"`
	Domain    string `json:"domain"`
}

type ShareLinkRsp struct {
	ShareId   string    `json:"shareId"`
	SessionId string    `json:"sessionId"`
	Title     string    `json:"title"`
	ExpiresAt time.Time `json:"expiresAt"`
	ViewCount int       `json:"viewCount"`
	IsExpired bool      `json:"isExpired"`
	ShareType string    `json:"shareType"`
	Created   time.Time `json:"created"`
}

type UserShareListItem struct {
	ShareId   string    `json:"shareId"`
	SessionId string    `json:"sessionId"`
	Title     string    `json:"title"`
	ExpiresAt time.Time `json:"expiresAt"`
	ViewCount int       `json:"viewCount"`
	IsExpired bool      `json:"isExpired"`
	Created   time.Time `json:"created"`
	ShareType string    `json:"shareType"`
}

type UserShareListReq struct {
	Page     int `url:"page"`
	PageSize int `url:"pageSize"`
}

type PublicShareRsp struct {
	ShareId   string        `json:"shareId"`
	Title     string        `json:"title"`
	Creator   string        `json:"creator"`
	Created   time.Time     `json:"created"`
	ExpiresAt time.Time     `json:"expiresAt"`
	ViewCount int           `json:"viewCount"`
	IsExpired bool          `json:"isExpired"`
	Messages  []MessageItem `json:"messages"`
	ShareType string        `json:"shareType"`
}

type ShareOGData struct {
	Title       string
	Description string
	ImageURL    string
	PageURL     string
}

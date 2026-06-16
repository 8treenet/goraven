package iface

import "raven/backend/po"

type MessageRepo interface {
	SaveChatMessage(sessionId string, msg *po.Message) error

	GetChatMessages(sessionId string) ([]*po.Message, error)

	AddSessionTokens(sessionId string, promptTokens, completionTokens, promptCachedTokens int) error

	SetContextTokens(sessionId string, tokens int) error

	UpdateSessionStatus(sessionId string, status int) error

	MarkSessionCompressed(sessionId string, roundIds []string) error
}

type DailyStatsRepo interface {
	AddDailyStats(userId string, promptTokens, completionTokens, promptCachedTokens int) error
	AddToolDailyStats(userId, toolType, toolName string) error
}

type SystemTask interface {
	Complete(replyContent string, err error)
	GetUserId() string
}

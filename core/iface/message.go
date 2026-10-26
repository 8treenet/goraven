package iface

import "goraven/backend/po"

// MessageRepo 消息持久化接口
type MessageRepo interface {
	//存储消息
	SaveChatMessage(sessionId string, msg *po.Message) error
	//获取 正序的记录
	GetChatMessages(sessionId string) ([]*po.Message, error)
	//累加session的token统计
	AddSessionTokens(sessionId string, promptTokens, completionTokens, promptCachedTokens int) error
	//设置当前上下文长度
	SetContextTokens(sessionId string, tokens int) error
	//更新session的状态
	UpdateSessionStatus(sessionId string, status int) error
	//标记session中指定roundIds的消息为已压缩
	MarkSessionCompressed(sessionId string, roundIds []string) error
}

// DailyStatsRepo token统计接口
type DailyStatsRepo interface {
	AddDailyStats(userId string, promptTokens, completionTokens, promptCachedTokens int) error
	// AddToolDailyStats 记录工具/技能调用次数 (toolType: mcp / skill / tool)
	AddToolDailyStats(userId, toolType, toolName string) error
}

// SystemTask
type SystemTask interface {
	//SystemAgent执行完成后的回调通知，replyContent为回复内容，err为执行过程中的错误
	Complete(replyContent string, err error)
	//返回用户Id
	GetUserId() string
}

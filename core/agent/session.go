package agent

import (
	"context"

	"goraven/backend/repository"
	"goraven/core/iface"
)

type SessionTitler struct {
}

func NewSessionTitler(flashModel iface.BaseChatModel, msgRepo iface.MessageRepo, statsRepo iface.DailyStatsRepo, sessionId, userId string) *SessionTitler {
	return &SessionTitler{}
}

func (t *SessionTitler) Generate(ctx context.Context, userContent, assistantReply string) (string, error) {
	return "", nil
}

type Compress struct {
}

func NewCompress(flashModel iface.BaseChatModel, msgRepo iface.MessageRepo, statsRepo iface.DailyStatsRepo, sessionId, userId string, sysCfg *repository.SystemConfig) *Compress {
	return &Compress{}
}

func (c *Compress) ForceCompress(ctx context.Context, history interface{}) (string, error) {
	return "", nil
}

func BuildHistoryFromMessages(messages interface{}) interface{} {
	return nil
}

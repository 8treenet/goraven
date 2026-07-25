package agent

import (
	"errors"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/iface"
)

type AgentParam struct {
	Session             *po.Session
	MsgRepo             iface.MessageRepo
	DailyStatsRepo      iface.DailyStatsRepo
	ChatModel           iface.BaseChatModel
	userSpace           string
	UserRole            string
	SystemSkillProvider iface.SystemSkillProvider
	SysCfg              *repository.SystemConfig
	Project             string
	UserName            string
}

func (p *AgentParam) UserId() string {
	if p.Session == nil {
		return ""
	}
	return p.Session.UserId
}

func (p *AgentParam) SessionId() string {
	if p.Session == nil {
		return ""
	}
	return p.Session.SessionId
}

func (p *AgentParam) Validate() error {
	if p.MsgRepo == nil {
		return errors.New("AgentParam.MsgRepo is required")
	}
	if p.ChatModel == nil {
		return errors.New("AgentParam.ChatModel is required")
	}
	if p.Session == nil || p.Session.UserId == "" {
		return errors.New("AgentParam.Session is required")
	}
	return nil
}

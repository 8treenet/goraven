package agent

import (
	"errors"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/iface"
)

type AgentParam struct {
	Session             *po.Session               // 会话信息
	MsgRepo             iface.MessageRepo         // 消息存储
	DailyStatsRepo      iface.DailyStatsRepo      // token统计
	ChatModel           iface.BaseChatModel       // 聊天模型
	userSpace           string                    // 用户的根目录
	UserRole            string                    // 用户设定的角色信息
	SystemSkillProvider iface.SystemSkillProvider // 系统技能提供者
	SysCfg              *repository.SystemConfig  // 系统配置（从 DB 加载）
	Project             string                    // 工作项目名称
	ProjectWorkspace    string                    // 工作项目所在的工作空间根路径（团队项目时为团队项目目录，个人项目为空默认使用userSpace）
	UserName            string
	DailyTokenLimit     int                      // 每日 Token 限额（单位 M，0=不限制）
	DailyTokenUsed      int                      // 今日已用 token（StartChat 时快照）
	WikiWriteMode       bool                     // 用户消息涉及 llmwiki 写入操作时放开只读限制
	AutomationTaskRepo  iface.AutomationTaskRepo // 自动化任务持久化，非 nil 时挂载创建自动化任务工具
	TaskMcpIds          []int                    // 当前会话的 MCP ID 列表（解析后原始值，供自动化任务暂存）
	TaskSkillIds        []int                    // 当前会话的技能 ID 列表（解析后原始值，供自动化任务暂存）
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

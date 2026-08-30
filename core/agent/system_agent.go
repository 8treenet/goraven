package agent

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"goraven/backend/repository"
	"goraven/config"
	"goraven/core/iface"
	"goraven/core/sandbox"
)

const systemAgentName = "system_agent"

type SystemAgentParam struct {
	ChatModel      iface.BaseChatModel      // 聊天模型
	DailyStatsRepo iface.DailyStatsRepo     // token统计
	SysCfg         *repository.SystemConfig // 系统配置
	UserId         string
	UserName       string
}

func (p *SystemAgentParam) Validate() error {
	if p.ChatModel == nil {
		return errors.New("SystemAgentParam.ChatModel is required")
	}
	return nil
}

func NewSystemAgent(param SystemAgentParam) (*SystemAgent, error) {
	box, boxerr := sandbox.NewSandbox(param.UserName)
	if boxerr != nil {
		return nil, boxerr
	}
	err := param.Validate()
	if err != nil {
		return nil, err
	}
	baseAgent := NewBaseAgent(systemAgentName, getSysInstructionPrompt(), getSystemAgentDescription(), param.ChatModel, param.SysCfg, box)
	baseAgent.SetValidateCommand(ValidateCommandForSystem)

	return &SystemAgent{
		BaseAgent:      *baseAgent,
		dailyStatsRepo: param.DailyStatsRepo,
		box:            box,
		userId:         param.UserId,
	}, nil
}

type SystemAgent struct {
	BaseAgent
	dailyStatsRepo iface.DailyStatsRepo
	box            sandbox.Sandbox
	userId         string
}

func (sa *SystemAgent) NewRunner(ctx context.Context) (*SystemRunner, error) {
	agent, err := sa.GetChatModelAgent()
	if err != nil {
		return nil, err
	}

	return newSystemRunner(sa, agent), nil
}

func (sa *SystemAgent) BuildInstallSkillMessage(skillName string) string {
	userSkillDir := filepath.Join(sa.box.GetWorkspace(), "skills")
	lang := config.Get().GetLanguage()
	if lang == "en" {
		return fmt.Sprintf(`Install dependencies for skill "%s". SKILL.md path: %s/%s/SKILL.md

Please report the result concisely, do not output detailed installation logs.`, skillName, userSkillDir, skillName)
	}
	return fmt.Sprintf(`安装技能 "%s" 的依赖。SKILL.md 路径: %s/%s/SKILL.md

请简洁汇报结果，不要输出安装过程中的详细日志。`, skillName, userSkillDir, skillName)
}

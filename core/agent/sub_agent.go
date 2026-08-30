package agent

import (
	"goraven/core/iface"
	"goraven/core/sandbox"
)

const subAgentName = "sub_agent"

func NewSubAgent(param AgentParam, chatmodel iface.BaseChatModel, sandbox sandbox.Sandbox) *SubAgent {
	baseAgent := NewBaseAgent(subAgentName, getSubAgentInstructionPrompt(param), getSubAgentDescription(), chatmodel, param.SysCfg, sandbox)
	return &SubAgent{BaseAgent: *baseAgent}
}

type SubAgent struct {
	BaseAgent
}

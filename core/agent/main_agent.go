package agent

import (
	"context"
	"goraven/core/iface"
	"goraven/core/tools"
)

func NewMainAgent(param AgentParam) (*MainAgent, error) {

	return &MainAgent{}, nil
}

type MainAgent struct {
}

func (main *MainAgent) SetCompressModel(model iface.BaseChatModel) {

}

func (main *MainAgent) SetVisualModel(model iface.BaseChatModel) {

}

func (main *MainAgent) NewRunner(ctx context.Context) (runner *MainRunner, e error) {
	return
}

func (main *MainAgent) AddMCP(tools.MCP) {

}

func (main *MainAgent) SetMCPFilter(filter *SimpleMCPFilter) {

}

func (main *MainAgent) SetSkillAccessor(filter *SimpleSkillFilter) {

}

func (main *MainAgent) SetFlashModel(model iface.BaseChatModel) {

}

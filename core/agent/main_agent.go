package agent

import (
	"context"
	"raven/core/iface"
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

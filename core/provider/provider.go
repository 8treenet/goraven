package provider

import (
	"errors"

	"raven/core/iface"
)

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

type Provider interface {
	Name() string
	Models() ([]ModelInfo, error)
	CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error)
	SetProxy(url string) error
}

type ProviderConfig struct {
	APIKey      string
	BaseURL     string
	ExtraFields string
}

func GetProviderByName(name string, cfg ProviderConfig) (Provider, error) {
	return nil, errors.New("unknown provider name: " + name)
}

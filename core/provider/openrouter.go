package provider

import (
	"context"
	"fmt"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/openrouter"
)

const OpenrouterProviderName = "open_router"

func NewOpenrouterProvider(apiKey string) *OpenrouterProvider {
	return &OpenrouterProvider{
		APIKey: apiKey,
	}
}

type OpenrouterProvider struct {
	APIKey   string
	proxyURL string
}

// CreateModel
func (provider *OpenrouterProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &openrouter.Config{
		APIKey:     provider.APIKey,
		Model:      modelName,
		HTTPClient: httpclient,
	}
	if reasoning {
		config.Reasoning = &openrouter.Reasoning{
			Effort: openrouter.EffortOfHigh,
		}
	}

	chatModel, err := openrouter.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &OpenrouterModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *OpenrouterProvider) Name() string {
	return OpenrouterProviderName
}

// Models returns models from OpenRouter API
func (provider *OpenrouterProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	err := requests.NewHTTPRequest("https://openrouter.ai/api/v1/models").Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

func (provider *OpenrouterProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type OpenrouterModel struct {
	openrouter.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName  string
	contextLen int
}

func (m *OpenrouterModel) ModelName() string {
	return m.modelName
}

func (m *OpenrouterModel) Provider() string {
	return OpenrouterProviderName
}

func (m *OpenrouterModel) ContextLength() int {
	return m.contextLen
}

func (m *OpenrouterModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

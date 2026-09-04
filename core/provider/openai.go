package provider

import (
	"context"
	"fmt"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

const OpenAIProviderName = "openai"

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		APIKey: apiKey,
	}
}

type OpenAIProvider struct {
	APIKey   string
	proxyURL string
}

func (provider *OpenAIProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &openai.ChatModelConfig{
		APIKey:     provider.APIKey,
		Model:      modelName,
		HTTPClient: httpclient,
	}

	if reasoning {
		config.ReasoningEffort = openai.ReasoningEffortLevelHigh
	}

	chatModel, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &OpenAIModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

func (provider *OpenAIProvider) Name() string {
	return OpenAIProviderName
}

func (provider *OpenAIProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	err := requests.NewHTTPRequest("https://api.openai.com/v1/models").Get().
		SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

func (provider *OpenAIProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type OpenAIModel struct {
	openai.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName  string
	contextLen int
}

func (m *OpenAIModel) ModelName() string {
	return m.modelName
}

func (m *OpenAIModel) Provider() string {
	return OpenAIProviderName
}

func (m *OpenAIModel) ContextLength() int {
	return m.contextLen
}

func (m *OpenAIModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

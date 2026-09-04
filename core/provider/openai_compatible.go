package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

const OpenAICompatibleProviderName = "openai_compatible"

func NewOpenAICompatibleProvider(providerName, apiKey string, baseURL string, extraFields string) *OpenAICompatibleProvider {
	url := baseURL
	mapdata := map[string]any{}
	//extraFields = `{"thinking":{"type":"enabled"}}`
	if extraFields != "" {
		json.Unmarshal([]byte(extraFields), &mapdata)
	}

	return &OpenAICompatibleProvider{
		APIKey:       apiKey,
		BaseURL:      url,
		providerName: providerName,
		extraFields:  mapdata,
	}
}

type OpenAICompatibleProvider struct {
	APIKey       string
	BaseURL      string
	providerName string
	extraFields  map[string]any
	proxyURL     string
}

func (provider *OpenAICompatibleProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &openai.ChatModelConfig{
		APIKey:     provider.APIKey,
		Model:      modelName,
		BaseURL:    provider.BaseURL,
		HTTPClient: httpclient,
	}

	if reasoning {
		config.ReasoningEffort = openai.ReasoningEffortLevelHigh
	}

	if len(provider.extraFields) > 0 {
		config.ExtraFields = provider.extraFields
	}

	chatModel, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &OpenAICompatibleModel{
		ChatModel:    *chatModel,
		modelName:    modelName,
		providerName: provider.providerName,
		contextLen:   contextLen,
	}, nil
}

func (provider *OpenAICompatibleProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

func (provider *OpenAICompatibleProvider) Name() string {
	return provider.providerName
}

func (provider *OpenAICompatibleProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	err := requests.NewHTTPRequest(provider.BaseURL+"/models").Get().
		SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

type OpenAICompatibleModel struct {
	openai.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName    string
	providerName string
	contextLen   int
}

func (m *OpenAICompatibleModel) ModelName() string {
	return m.modelName
}

func (m *OpenAICompatibleModel) Provider() string {
	return m.providerName
}

func (m *OpenAICompatibleModel) ContextLength() int {
	return m.contextLen
}

func (m *OpenAICompatibleModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

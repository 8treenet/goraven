package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"goraven/core/iface"
	"goraven/util"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

const VolcanoProviderName = "volcano"

func NewVolcanoProvider(apiKey string, baseURL string) *VolcanoProvider {
	return &VolcanoProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

type VolcanoProvider struct {
	APIKey   string
	BaseURL  string
	proxyURL string
}

// CreateModel
func (provider *VolcanoProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &openai.ChatModelConfig{
		APIKey:     provider.APIKey,
		Model:      modelName,
		BaseURL:    baseURL,
		HTTPClient: httpclient,
	}

	if reasoning {
		config.ExtraFields = map[string]any{}
		json.Unmarshal([]byte(`{"thinking":{"type":"enabled"}}`), &config.ExtraFields)
	}

	chatModel, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &VolcanoModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *VolcanoProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

func (provider *VolcanoProvider) Name() string {
	return VolcanoProviderName
}

// Models returns models from Volcano Ark API (OpenAI compatible)
func (provider *VolcanoProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	url := fmt.Sprintf("%s/models", baseURL)
	err := requests.NewHTTPRequest(url).Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

type VolcanoModel struct {
	openai.ChatModel
	modelName  string
	contextLen int
}

func (m *VolcanoModel) ModelName() string {
	return m.modelName
}

func (m *VolcanoModel) Provider() string {
	return VolcanoProviderName
}

func (m *VolcanoModel) ContextLength() int {
	return m.contextLen
}

func (m *VolcanoModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

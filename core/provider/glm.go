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

const GLMProviderName = "glm"

func NewGLMProvider(apiKey string, baseURL string) *GLMProvider {
	return &GLMProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

type GLMProvider struct {
	APIKey   string
	BaseURL  string
	proxyURL string
}

// CreateModel
func (provider *GLMProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
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

	return &GLMModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *GLMProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

func (provider *GLMProvider) Name() string {
	return GLMProviderName
}

// Models returns models from Anthropic API
func (provider *GLMProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}
	url := fmt.Sprintf("%s/models", baseURL)
	err := requests.NewHTTPRequest(url).Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

type GLMModel struct {
	openai.ChatModel
	modelName  string
	contextLen int
}

func (m *GLMModel) ModelName() string {
	return m.modelName
}

func (m *GLMModel) Provider() string {
	return GLMProviderName
}

func (m *GLMModel) ContextLength() int {
	return m.contextLen
}

func (m *GLMModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

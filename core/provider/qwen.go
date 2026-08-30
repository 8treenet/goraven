package provider

import (
	"context"
	"fmt"
	"net/url"
	"goraven/core/iface"
	"goraven/util"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/qwen"
)

const QwenProviderName = "bailian"

func NewQwenProvider(apiKey string, baseURL string) *QwenProvider {
	return &QwenProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

type QwenProvider struct {
	APIKey   string
	BaseURL  string
	proxyURL string
}

// CreateModel
func (provider *QwenProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &qwen.ChatModelConfig{
		APIKey:     provider.APIKey,
		BaseURL:    baseURL,
		Model:      modelName,
		HTTPClient: httpclient,
	}

	if reasoning {
		config.EnableThinking = &reasoning
	}

	chatModel, err := qwen.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &QwenModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *QwenProvider) Name() string {
	return QwenProviderName
}

// Models returns models from Qwen API (OpenAI compatible)
func (provider *QwenProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

func (provider *QwenProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	url := fmt.Sprintf("%s/models", baseURL)
	err := requests.NewHTTPRequest(url).Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

type QwenModel struct {
	qwen.ChatModel
	modelName  string
	contextLen int
}

func (m *QwenModel) ModelName() string {
	return m.modelName
}

func (m *QwenModel) Provider() string {
	return QwenProviderName
}

func (m *QwenModel) ContextLength() int {
	return m.contextLen
}

func (m *QwenModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

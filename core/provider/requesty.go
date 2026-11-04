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

const RequestyProviderName = "requesty"

const requestyBaseURL = "https://router.requesty.ai/v1"

func NewRequestyProvider(apiKey string) *RequestyProvider {
	return &RequestyProvider{
		APIKey: apiKey,
	}
}

type RequestyProvider struct {
	APIKey   string
	proxyURL string
}

// CreateModel
func (provider *RequestyProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &openai.ChatModelConfig{
		APIKey:     provider.APIKey,
		Model:      modelName,
		BaseURL:    requestyBaseURL,
		HTTPClient: httpclient,
	}
	if reasoning {
		config.ReasoningEffort = openai.ReasoningEffortLevelHigh
	}

	chatModel, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &RequestyModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *RequestyProvider) Name() string {
	return RequestyProviderName
}

// Models returns the Requesty managed models, falling back to the full catalog
func (provider *RequestyProvider) Models() ([]ModelInfo, error) {
	models, err := provider.fetchModels(requestyBaseURL + "/models/managed")
	if err == nil && len(models) > 0 {
		return models, nil
	}
	return provider.fetchModels(requestyBaseURL + "/models")
}

func (provider *RequestyProvider) fetchModels(endpoint string) ([]ModelInfo, error) {
	var body struct {
		Data []struct {
			ModelInfo
			API string `json:"api"`
		} `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.APIKey)
	err := requests.NewHTTPRequest(endpoint).Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	models := make([]ModelInfo, 0, len(body.Data))
	for _, m := range body.Data {
		if m.API != "" && m.API != "chat" {
			continue
		}
		models = append(models, m.ModelInfo)
	}
	return models, nil
}

func (provider *RequestyProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type RequestyModel struct {
	openai.ChatModel
	iface.ConversationHeaderKeyHolder
	iface.VisualSupportHolder
	modelName  string
	contextLen int
}

func (m *RequestyModel) ModelName() string {
	return m.modelName
}

func (m *RequestyModel) Provider() string {
	return RequestyProviderName
}

func (m *RequestyModel) ContextLength() int {
	return m.contextLen
}

func (m *RequestyModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

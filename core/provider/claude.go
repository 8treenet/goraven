package provider

import (
	"context"
	"goraven/core/iface"
	"goraven/util"
	"net/url"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cloudwego/eino-ext/components/model/claude"
)

const ClaudeProviderName = "claude"

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		APIKey: apiKey,
	}
}

type ClaudeProvider struct {
	APIKey   string
	proxyURL string
}

func (provider *ClaudeProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config := &claude.Config{
		APIKey:     provider.APIKey,
		Model:      modelName,
		HTTPClient: httpclient,
	}
	if reasoning {
		config.MaxTokens = 64000
		config.ThinkingConfig = &anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 32000,
			},
		}
	}

	chatModel, err := claude.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &ClaudeModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

func (provider *ClaudeProvider) Name() string {
	return ClaudeProviderName
}

func (provider *ClaudeProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Data []struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			DisplayName string `json:"display_name"`
			CreatedAt   string `json:"created_at"`
		} `json:"data"`
	}
	err := requests.NewHTTPRequest("https://api.anthropic.com/v1/models?limit=1000").Get().
		SetHeaderValue("x-api-key", provider.APIKey).
		SetHeaderValue("anthropic-version", "2023-06-01").ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	result := make([]ModelInfo, len(body.Data))
	for i, m := range body.Data {
		result[i] = ModelInfo{ID: m.ID, Object: m.Type, OwnedBy: "anthropic"}
	}
	return result, nil
}

func (provider *ClaudeProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type ClaudeModel struct {
	claude.ChatModel
	iface.ConversationHeaderKeyHolder
	modelName  string
	contextLen int
}

func (m *ClaudeModel) ModelName() string {
	return m.modelName
}

func (m *ClaudeModel) Provider() string {
	return ClaudeProviderName
}

func (m *ClaudeModel) ContextLength() int {
	return m.contextLen
}

func (m *ClaudeModel) Format() iface.APIFormat {
	return iface.APIFormatClaude
}

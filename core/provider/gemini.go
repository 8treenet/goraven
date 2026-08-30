package provider

import (
	"context"
	"fmt"
	"net/url"
	"goraven/core/iface"
	"goraven/util"
	"strings"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"google.golang.org/genai"
)

const GeminiProviderName = "gemini"

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		APIKey: apiKey,
	}
}

type GeminiProvider struct {
	APIKey   string
	proxyURL string
}

func (provider *GeminiProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:     provider.APIKey,
		HTTPClient: httpclient,
	})
	if err != nil {
		return nil, err
	}

	config := &gemini.Config{
		Client: client,
		Model:  modelName,
	}

	if reasoning {
		config.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingLevel:   genai.ThinkingLevelHigh,
		}
	}

	chatModel, err := gemini.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}

	return &GeminiModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

func (provider *GeminiProvider) Name() string {
	return GeminiProviderName
}

func (provider *GeminiProvider) Models() ([]ModelInfo, error) {
	var body struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", provider.APIKey)
	err := requests.NewHTTPRequest(url).Get().ToJSON(&body).Error
	if err != nil {
		return nil, err
	}
	var result []ModelInfo
	for _, m := range body.Models {
		id := strings.TrimPrefix(m.Name, "models/")
		if id == m.Name || id == "" {
			continue
		}
		result = append(result, ModelInfo{ID: id, Object: "model", OwnedBy: "google"})
	}
	return result, nil
}

func (provider *GeminiProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type GeminiModel struct {
	gemini.ChatModel
	modelName  string
	contextLen int
}

func (m *GeminiModel) ModelName() string {
	return m.modelName
}

func (m *GeminiModel) Provider() string {
	return GeminiProviderName
}

func (m *GeminiModel) ContextLength() int {
	return m.contextLen
}

func (m *GeminiModel) Format() iface.APIFormat {
	return iface.APIFormatGemini
}

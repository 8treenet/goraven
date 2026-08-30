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

const DeepseekProviderName = "deepseek"

func NewDeepSeekProvider(key string, baseURL string) *DeepSeekProvider {
	mapdata := map[string]any{}
	json.Unmarshal([]byte(`{"thinking":{"type":"enabled"}}`), &mapdata)
	return &DeepSeekProvider{Key: key, BaseURL: baseURL, extraFields: mapdata}
}

type DeepSeekProvider struct {
	Key         string
	BaseURL     string
	extraFields map[string]any
	proxyURL    string
}

// CreateModel
func (provider *DeepSeekProvider) CreateModel(modelName string, reasoning bool, contextLen int) (iface.BaseChatModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	config := &openai.ChatModelConfig{
		APIKey:  provider.Key,
		Model:   modelName,
		BaseURL: baseURL,
	}

	if reasoning {
		config.ReasoningEffort = "xhigh"
		config.ExtraFields = provider.extraFields
	}

	if len(provider.extraFields) > 0 {
		config.ExtraFields = provider.extraFields
	}

	httpclient := util.GetGoRavenHTTPClient()
	if provider.proxyURL != "" {
		httpclient = util.GetProxyHTTPClient(provider.proxyURL)
	}
	config.HTTPClient = httpclient

	chatModel, err := openai.NewChatModel(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return &DeepseekModel{
		ChatModel:  *chatModel,
		modelName:  modelName,
		contextLen: contextLen,
	}, nil
}

// Name
func (provider *DeepSeekProvider) Name() string {
	return DeepseekProviderName
}

func (provider *DeepSeekProvider) Models() (result []ModelInfo, e error) {
	var body struct {
		Data []ModelInfo `json:"data"`
	}
	key := fmt.Sprintf("Bearer %s", provider.Key)
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	e = requests.NewHTTPRequest(baseURL+"/models").Get().SetHeaderValue("Authorization", key).ToJSON(&body).Error
	result = body.Data
	return
}

func (provider *DeepSeekProvider) SetProxy(addr string) error {
	_, err := url.Parse(addr)
	provider.proxyURL = addr
	return err
}

type DeepseekModel struct {
	openai.ChatModel
	modelName  string
	contextLen int
}

func (dm *DeepseekModel) ModelName() string {
	return dm.modelName
}

func (dm *DeepseekModel) Provider() string {
	return DeepseekProviderName
}

func (dm *DeepseekModel) ContextLength() int {
	return dm.contextLen
}

func (dm *DeepseekModel) Format() iface.APIFormat {
	return iface.APIFormatOpenAI
}

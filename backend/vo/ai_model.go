package vo

import "time"

// AdminModelListReq 管理员模型列表请求
type AdminModelListReq struct {
	ProviderID string `url:"providerId"` // 供应商标识筛选，如 "deepseek"
	Search     string `url:"search"`     // 模型名模糊搜索
	Page       int    `url:"page"`       // 页码
	PageSize   int    `url:"pageSize"`   // 每页条数
}

// AdminModelItem 管理员模型列表条目（apiKey脱敏）
type AdminModelItem struct {
	AIModelId             int       `json:"aiModelId"`             // 模型ID
	ProviderDisplayName   string    `json:"providerDisplayName"`   // 供应商显示名称，如 "DeepSeek"、"百炼"
	DisplayName           string    `json:"displayName"`           // 模型显示名称，如 "DeepSeek V3"
	ProviderID            string    `json:"providerId"`            // 供应商标识，如 "deepseek"、"bailian"
	ModelName             string    `json:"modelName"`             // 模型名称，如 "deepseek-chat"
	Icon                  string    `json:"icon"`                  // 图标URL或Base64
	APIKeyMasked          string    `json:"apiKeyMasked"`          // API Key脱敏，如 ["sk-x1****bc1"]，仅展示不可复制
	BaseURL               string    `json:"baseUrl"`               // 基础URL
	ProxyURL              string    `json:"proxyUrl"`              // 代理地址
	ConversationHeaderKey string    `json:"conversationHeaderKey"` // 会话归并 header 名，非必填，空则不注入
	ContextLen            int       `json:"contextLen"`            // 上下文长度，单位KB
	ExtraFields           string    `json:"extraFields"`           // 额外配置JSON
	IsDefault             uint8     `json:"isDefault"`             // 是否默认模型: 0否 1是
	IsFlash               int       `json:"isFlash"`               // 是否 Flash 模型: 0否 1是
	IsVisual              int       `json:"isVisual"`              // 是否多模态识别模型: 0否 1是
	Status                uint8     `json:"status"`                // 状态: 0禁用 1启用
	Access                uint8     `json:"access"`                // 访问权限：0全员开放 1仅成员可见
	Remark                string    `json:"remark"`                // 备注
	Created               time.Time `json:"created"`               // 创建时间
	Updated               time.Time `json:"updated"`               // 更新时间
}

// AdminCreateModelReq 管理员创建模型请求
type AdminCreateModelReq struct {
	ProviderDisplayName   string `json:"providerDisplayName" validate:"required"` // 供应商显示名称，前端自定义
	DisplayName           string `json:"displayName"`                             // 模型显示名称，如 "DeepSeek V3"
	ProviderID            string `json:"providerId" validate:"required"`          // 供应商标识，对应provider包常量
	ModelName             string `json:"modelName" validate:"required"`           // 模型名称
	Icon                  string `json:"icon"`                                    // 图标URL或Base64，非必填
	APIKey                string `json:"apiKey"`                                  // API密钥；ollama可空；openai_compatible/claude_compatible不可空；其他必填
	BaseURL               string `json:"baseUrl"`                                 // 基础URL，ollama/openai_compatible/claude_compatible必填
	ExtraFields           string `json:"extraFields"`                             // 额外配置JSON，如 {"thinking":{"type":"enabled"}}
	ProxyURL              string `json:"proxyUrl"`                                // 代理地址，非必填，填写后创建时测试连通性
	ConversationHeaderKey string `json:"conversationHeaderKey"`                   // 会话归并 header 名，非必填，空则不注入
	ContextLen            int    `json:"contextLen"`                              // 上下文长度KB，默认200
	IsDefault             uint8  `json:"isDefault"`                               // 是否默认模型: 0否 1是，设置1时自动取消其他默认
	IsFlash               int    `json:"isFlash"`                                 // 是否 Flash 模型: 0否 1是，设置1时自动取消其他 Flash
	IsVisual              int    `json:"isVisual"`                                // 是否多模态识别模型: 0否 1是，设置1时自动取消其他多模态
	Remark                string `json:"remark"`                                  // 备注
}

// AdminUpdateModelReq 管理员编辑模型请求（仅传需要修改的字段，providerId不可改）
type AdminUpdateModelReq struct {
	ProviderDisplayName   string  `json:"providerDisplayName"`   // 供应商显示名称
	DisplayName           string  `json:"displayName"`           // 模型显示名称，如 "DeepSeek V3"
	ModelName             string  `json:"modelName"`             // 模型名称
	Icon                  string  `json:"icon"`                  // 图标URL或Base64
	APIKey                string  `json:"apiKey"`                // API密钥，变化时触发连通性测试
	BaseURL               string  `json:"baseUrl"`               // 基础URL，变化时触发连通性测试
	ExtraFields           string  `json:"extraFields"`           // 额外配置JSON
	ProxyURL              string  `json:"proxyUrl"`              // 代理地址，变化时触发连通性测试
	ConversationHeaderKey *string `json:"conversationHeaderKey"` // 会话归并 header 名，nil不修改，空串清除
	ContextLen            int     `json:"contextLen"`            // 上下文长度KB
	IsDefault             *uint8  `json:"isDefault"`             // 是否默认模型，nil不修改
	IsFlash               *int    `json:"isFlash"`               // 是否 Flash 模型，nil不修改
	IsVisual              *int    `json:"isVisual"`              // 是否多模态识别模型，nil不修改
	Status                *uint8  `json:"status"`                // 状态: 0禁用 1启用，nil不修改
	Remark                string  `json:"remark"`                // 备注
}

// AdminModelDetailRsp 管理员模型详情响应（含原始apiKey，用于编辑回填和Key复制）
type AdminModelDetailRsp struct {
	AIModelId             int       `json:"aiModelId"`             // 模型ID
	ProviderDisplayName   string    `json:"providerDisplayName"`   // 供应商显示名称
	DisplayName           string    `json:"displayName"`           // 模型显示名称，如 "DeepSeek V3"
	ProviderID            string    `json:"providerId"`            // 供应商标识
	ModelName             string    `json:"modelName"`             // 模型名称
	Icon                  string    `json:"icon"`                  // 图标URL或Base64
	APIKey                string    `json:"apiKey"`                // 原始API密钥，前端可复制
	BaseURL               string    `json:"baseUrl"`               // 基础URL
	ExtraFields           string    `json:"extraFields"`           // 额外配置JSON
	ProxyURL              string    `json:"proxyUrl"`              // 代理地址
	ConversationHeaderKey string    `json:"conversationHeaderKey"` // 会话归并 header 名，非必填，空则不注入
	ContextLen            int       `json:"contextLen"`            // 上下文长度KB
	IsDefault             uint8     `json:"isDefault"`             // 是否默认模型: 0否 1是
	IsFlash               int       `json:"isFlash"`               // 是否 Flash 模型: 0否 1是
	IsVisual              int       `json:"isVisual"`              // 是否多模态识别模型: 0否 1是
	Status                uint8     `json:"status"`                // 状态: 0禁用 1启用
	Access                uint8     `json:"access"`                // 访问权限：0全员开放 1仅成员可见
	Remark                string    `json:"remark"`                // 备注
	Created               time.Time `json:"created"`               // 创建时间
	Updated               time.Time `json:"updated"`               // 更新时间
}

// ProviderItem 供应商列表条目
type ProviderItem struct {
	ProviderID            string `json:"providerId"`            // 供应商标识，对应provider包常量，如 "deepseek"、"bailian"
	ProviderDisplayNameZh string `json:"providerDisplayNameZh"` // 中文显示名称，如 "DeepSeek"、"百炼"、"自定义 OpenAI"
	ProviderDisplayNameEn string `json:"providerDisplayNameEn"` // 英文显示名称，如 "DeepSeek"、"Bailian"、"Custom OpenAI"
	Icon                  string `json:"icon"`                  // 默认图标URL，如 "/logos/deepseek.svg"，前端用作新增模型时的默认图标
	DefaultBaseURL        string `json:"defaultBaseUrl"`        // 按当前语言预设的默认 BaseURL（用户可改），如 "https://api.deepseek.com/v1"；空表示该 provider 不预设
	RequireAPIKey         bool   `json:"requireApiKey"`         // 是否必填API Key（ollama为false，其余为true）
	RequireBaseURL        bool   `json:"requireBaseUrl"`        // 是否必填Base URL（ollama/openai_compatible/claude_compatible/国内5家为true）
}

// RecommendModelItem 推荐模型条目，来源于provider.Models()接口
type RecommendModelItem struct {
	ID      string `json:"id"`      // 模型ID，如 "deepseek-chat"
	Object  string `json:"object"`  // 对象类型，通常为 "model"
	OwnedBy string `json:"ownedBy"` // 模型归属，如 "deepseek"
}

// UserModelItem 用户可选模型列表条目
type UserModelItem struct {
	AIModelId           int    `json:"aiModelId"`           // 模型ID
	ProviderDisplayName string `json:"providerDisplayName"` // 供应商显示名称，如 "DeepSeek"
	DisplayName         string `json:"displayName"`         // 模型显示名称，如 "DeepSeek V3"
	ModelName           string `json:"modelName"`           // 模型名称，如 "deepseek-chat"
	Icon                string `json:"icon"`                // 图标URL或Base64
	ContextLen          int    `json:"contextLen"`          // 上下文长度，单位KB
	IsDefault           uint8  `json:"isDefault"`           // 是否默认模型: 0否 1是
	IsFlash             int    `json:"isFlash"`             // 是否 Flash 模型: 0否 1是
	IsVisual            int    `json:"isVisual"`            // 是否多模态模型: 0否 1是
}

// AIModelMembersRsp 模型成员列表响应
type AIModelMembersRsp struct {
	MemberIds []string `json:"memberIds"` // 成员 userId 列表
}

// AIModelMemberUpdateReq 编辑成员请求 PUT /api/admin/models/:id/members
type AIModelMemberUpdateReq struct {
	AddUserIds    []string `json:"addUserIds"`    // 要添加的用户 ID 列表
	RemoveUserIds []string `json:"removeUserIds"` // 要移除的用户 ID 列表
}

// AIModelAccessUpdateReq 设置访问权限请求 PUT /api/admin/models/:id/access
type AIModelAccessUpdateReq struct {
	Access uint8 `json:"access"` // 访问权限：0全员开放 1仅成员可见
}

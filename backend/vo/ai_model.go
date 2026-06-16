package vo

import "time"

type AdminModelListReq struct {
	ProviderID	string	`url:"providerId"`
	Search		string	`url:"search"`
	Page		int	`url:"page"`
	PageSize	int	`url:"pageSize"`
}

type AdminModelItem struct {
	AIModelId		int		`json:"aiModelId"`
	ProviderDisplayName	string		`json:"providerDisplayName"`
	DisplayName		string		`json:"displayName"`
	ProviderID		string		`json:"providerId"`
	ModelName		string		`json:"modelName"`
	Icon			string		`json:"icon"`
	APIKeyMasked		string		`json:"apiKeyMasked"`
	BaseURL			string		`json:"baseUrl"`
	ProxyURL		string		`json:"proxyUrl"`
	ContextLen		int		`json:"contextLen"`
	ExtraFields		string		`json:"extraFields"`
	IsDefault		uint8		`json:"isDefault"`
	IsCompress		int		`json:"isCompress"`
	IsVisual		int		`json:"isVisual"`
	Status			uint8		`json:"status"`
	Remark			string		`json:"remark"`
	Created			time.Time	`json:"created"`
	Updated			time.Time	`json:"updated"`
}

type AdminCreateModelReq struct {
	ProviderDisplayName	string	`json:"providerDisplayName" validate:"required"`
	DisplayName		string	`json:"displayName"`
	ProviderID		string	`json:"providerId" validate:"required"`
	ModelName		string	`json:"modelName" validate:"required"`
	Icon			string	`json:"icon"`
	APIKey			string	`json:"apiKey"`
	BaseURL			string	`json:"baseUrl"`
	ExtraFields		string	`json:"extraFields"`
	ProxyURL		string	`json:"proxyUrl"`
	ContextLen		int	`json:"contextLen"`
	IsDefault		uint8	`json:"isDefault"`
	IsCompress		int	`json:"isCompress"`
	IsVisual		int	`json:"isVisual"`
	Remark			string	`json:"remark"`
}

type AdminUpdateModelReq struct {
	ProviderDisplayName	string	`json:"providerDisplayName"`
	DisplayName		string	`json:"displayName"`
	ModelName		string	`json:"modelName"`
	Icon			string	`json:"icon"`
	APIKey			string	`json:"apiKey"`
	BaseURL			string	`json:"baseUrl"`
	ExtraFields		string	`json:"extraFields"`
	ProxyURL		string	`json:"proxyUrl"`
	ContextLen		int	`json:"contextLen"`
	IsDefault		*uint8	`json:"isDefault"`
	IsCompress		*int	`json:"isCompress"`
	IsVisual		*int	`json:"isVisual"`
	Status			*uint8	`json:"status"`
	Remark			string	`json:"remark"`
}

type AdminModelDetailRsp struct {
	AIModelId		int		`json:"aiModelId"`
	ProviderDisplayName	string		`json:"providerDisplayName"`
	DisplayName		string		`json:"displayName"`
	ProviderID		string		`json:"providerId"`
	ModelName		string		`json:"modelName"`
	Icon			string		`json:"icon"`
	APIKey			string		`json:"apiKey"`
	BaseURL			string		`json:"baseUrl"`
	ExtraFields		string		`json:"extraFields"`
	ProxyURL		string		`json:"proxyUrl"`
	ContextLen		int		`json:"contextLen"`
	IsDefault		uint8		`json:"isDefault"`
	IsCompress		int		`json:"isCompress"`
	IsVisual		int		`json:"isVisual"`
	Status			uint8		`json:"status"`
	Remark			string		`json:"remark"`
	Created			time.Time	`json:"created"`
	Updated			time.Time	`json:"updated"`
}

type ProviderItem struct {
	ProviderID		string	`json:"providerId"`
	ProviderDisplayNameZh	string	`json:"providerDisplayNameZh"`
	ProviderDisplayNameEn	string	`json:"providerDisplayNameEn"`
	Icon			string	`json:"icon"`
	DefaultBaseURL		string	`json:"defaultBaseUrl"`
	RequireAPIKey		bool	`json:"requireApiKey"`
	RequireBaseURL		bool	`json:"requireBaseUrl"`
}

type RecommendModelItem struct {
	ID	string	`json:"id"`
	Object	string	`json:"object"`
	OwnedBy	string	`json:"ownedBy"`
}

type UserModelItem struct {
	AIModelId		int	`json:"aiModelId"`
	ProviderDisplayName	string	`json:"providerDisplayName"`
	DisplayName		string	`json:"displayName"`
	ModelName		string	`json:"modelName"`
	Icon			string	`json:"icon"`
	ContextLen		int	`json:"contextLen"`
	IsDefault		uint8	`json:"isDefault"`
	IsCompress		int	`json:"isCompress"`
	IsVisual		int	`json:"isVisual"`
}

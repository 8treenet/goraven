package seed

import (
	"goraven/config"
)

type ProviderDef struct {
	ID               string
	DisplayNameZh    string
	DisplayNameEn    string
	Icon             string
	DefaultBaseURLZh string
	DefaultBaseURLEn string
	RequireAPIKey    bool
	RequireBaseURL   bool
}

func (def *ProviderDef) CurrentDefaultBaseURL() string {
	if config.Get().GetLanguage() == "en" {
		return def.DefaultBaseURLEn
	}
	return def.DefaultBaseURLZh
}

var ProviderDefs = []ProviderDef{}

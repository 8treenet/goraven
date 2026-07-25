package vo

type AdminSettingGroup struct {
	Name         string             `json:"name"`
	DisplayName  string             `json:"displayName"`
	DisplayOrder int                `json:"displayOrder"`
	Settings     []AdminSettingItem `json:"settings"`
}

type AdminSettingItem struct {
	Key          string   `json:"key"`
	Value        string   `json:"value"`
	ValueType    string   `json:"valueType"`
	DefaultValue string   `json:"defaultValue"`
	DisplayName  string   `json:"displayName"`
	Description  string   `json:"description"`
	InputType    string   `json:"inputType"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Placeholder  string   `json:"placeholder,omitempty"`
	DisplayOrder int      `json:"displayOrder"`
}

type AdminUpdateSettingsReq struct {
	Settings []AdminSettingUpdateItem `json:"settings" validate:"required,dive"`
}

type AdminSettingUpdateItem struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type AdminUpdateSettingsRsp struct {
	Updated int `json:"updated"`
}

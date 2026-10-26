package vo

// AdminSettingGroup 系统设置分组响应
type AdminSettingGroup struct {
	Name         string              `json:"name"`         // 分组标识，如 "agent"、"general"
	DisplayName  string              `json:"displayName"`  // 分组显示名（跟随系统语言）
	DisplayOrder int                 `json:"displayOrder"` // 排序权重，从小到大
	Settings     []AdminSettingItem  `json:"settings"`     // 分组内的设置项列表
}

// AdminSettingItem 单个设置项（含 UI 元数据，前端据此动态渲染控件）
type AdminSettingItem struct {
	Key          string  `json:"key"`          // 设置项唯一标识，对应 DB configKey
	Value        string  `json:"value"`        // 当前值（统一 string 存储）
	ValueType    string  `json:"valueType"`    // 值类型：string/int/float/bool/date/datetime
	DefaultValue string  `json:"defaultValue"` // 默认值
	DisplayName  string  `json:"displayName"`  // 显示名（跟随系统语言）
	Description  string  `json:"description"`  // 描述说明（跟随系统语言）
	InputType    string  `json:"inputType"`    // 控件类型：text/number/slider/switch/password/textarea/date/datetime/select/jsonEditor
	Min          *float64 `json:"min,omitempty"` // 最小值，仅 number/slider 有效
	Max          *float64 `json:"max,omitempty"` // 最大值，仅 number/slider 有效
	Placeholder  string   `json:"placeholder,omitempty"` // 输入框占位符，仅 ValueType=string 有效，可为空
	DisplayOrder int      `json:"displayOrder"` // 排序权重，同组内从小到大
}

// AdminUpdateSettingsReq 批量更新系统设置请求
type AdminUpdateSettingsReq struct {
	Settings []AdminSettingUpdateItem `json:"settings" validate:"required,dive"` // 待更新的设置项列表
}

// AdminSettingUpdateItem 单个设置项更新
type AdminSettingUpdateItem struct {
	Key   string `json:"key" validate:"required"`   // 设置项标识
	Value string `json:"value"` // 新值
}

// AdminUpdateSettingsRsp 批量更新系统设置响应
type AdminUpdateSettingsRsp struct {
	Updated int `json:"updated"` // 成功更新的条数
}

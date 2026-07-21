package vo

type AdminDashboardRsp struct {
	Overview       AdminDashboardOverview `json:"overview"`
	SkillUsageRank []ToolUsageRankItem    `json:"skillUsageRank"`
	McpUsageRank   []ToolUsageRankItem    `json:"mcpUsageRank"`
	ToolUsageRank  []ToolUsageRankItem    `json:"toolUsageRank"`
}

type AdminDashboardOverview struct {
	ActiveUsers     int             `json:"activeUsers"`
	ActiveUsersDiff float64         `json:"activeUsersDiff"`
	TotalSessions   int64           `json:"totalSessions"`
	NewSessions     int64           `json:"newSessions"`
	WeekTokens      int64           `json:"weekTokens"`
	TodayTokens     int64           `json:"todayTokens"`
	EnabledModels   int64           `json:"enabledModels"`
	Sparkline       []SparklineItem `json:"sparkline"`
}

type UserDashboardRsp struct {
	Overview       UserDashboardOverview `json:"overview"`
	SkillUsageRank []ToolUsageRankItem   `json:"skillUsageRank"`
	McpUsageRank   []ToolUsageRankItem   `json:"mcpUsageRank"`
	ToolUsageRank  []ToolUsageRankItem   `json:"toolUsageRank"`
	StorageStats   UserStorageStats      `json:"storageStats"`
}

type UserDashboardOverview struct {
	TodayTokens   int64           `json:"todayTokens"`
	WeekTokens    int64           `json:"weekTokens"`
	TotalTokens   int64           `json:"totalTokens"`
	TotalSessions int64           `json:"totalSessions"`
	NewSessions   int64           `json:"newSessions"`
	Sparkline     []SparklineItem `json:"sparkline"`
}

type SparklineItem struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

type TokenTrendItem struct {
	Date             string `json:"date"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
}

type ModelUsageItem struct {
	ModelName        string  `json:"modelName"`
	TokenCount       int64   `json:"tokenCount"`
	Percentage       float64 `json:"percentage"`
	PromptTokens     int64   `json:"promptTokens"`
	CompletionTokens int64   `json:"completionTokens"`
}

type UserTokenRankItem struct {
	UserId     string  `json:"userId"`
	Username   string  `json:"username"`
	TokenCount int64   `json:"tokenCount"`
	Percentage float64 `json:"percentage"`
}

type ActiveUserTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ToolUsageRankItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type UserStorageStats struct {
	UsedBytes  int64              `json:"usedBytes"`
	FreeBytes  int64              `json:"freeBytes"`
	TotalBytes int64              `json:"totalBytes"`
	Items      []StorageUsageItem `json:"items"`
}

type StorageUsageItem struct {
	Name       string  `json:"name"`
	BytesSize  int64   `json:"bytesSize"`
	Percentage float64 `json:"percentage"`
}

type TokenTrendRsp struct {
	Items []TokenTrendItem `json:"items"`
}

type ActiveUserTrendRsp struct {
	Items []ActiveUserTrendItem `json:"items"`
}

type ModelUsageRsp struct {
	Items []ModelUsageItem `json:"items"`
}

type UserTokenRankRsp struct {
	Items []UserTokenRankItem `json:"items"`
}

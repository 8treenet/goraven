package vo

// ──────────────────────────────────────────────
// 管理员仪表盘
// ──────────────────────────────────────────────

// AdminDashboardRsp 管理员仪表盘聚合响应
// 趋势图表和排行榜通过独立端点按需加载，此处仅返回脉搏指标和工具排行
type AdminDashboardRsp struct {
	Overview       AdminDashboardOverview `json:"overview"`       // 系统脉搏核心指标
	SkillUsageRank []ToolUsageRankItem    `json:"skillUsageRank"` // 近一周技能使用排行
	McpUsageRank   []ToolUsageRankItem    `json:"mcpUsageRank"`   // 近一周 MCP 使用排行
	ToolUsageRank  []ToolUsageRankItem    `json:"toolUsageRank"`  // 近一周工具使用排行
}

// AdminDashboardOverview 系统脉搏核心指标
// 管理员第一眼看到的内容，整合活跃、会话、Token、模型四大指标
type AdminDashboardOverview struct {
	ActiveUsers     int             `json:"activeUsers"`     // 近7日活跃用户数
	ActiveUsersDiff float64         `json:"activeUsersDiff"` // 环比上周变化率（正数上升，负数下降）
	TotalSessions   int64           `json:"totalSessions"`   // 会话总数
	NewSessions     int64           `json:"newSessions"`     // 本周新增会话
	WeekTokens      int64           `json:"weekTokens"`      // 本周 Token 消耗
	TodayTokens     int64           `json:"todayTokens"`     // 今日 Token 消耗
	EnabledModels   int64           `json:"enabledModels"`   // 启用模型数
	Sparkline       []SparklineItem `json:"sparkline"`       // 7日迷你趋势数据
}

// ──────────────────────────────────────────────
// 用户仪表盘
// ──────────────────────────────────────────────

// UserDashboardRsp 用户仪表盘响应
// 趋势图表通过独立端点按需加载，此处返回脉搏指标、存储统计和工具排行
type UserDashboardRsp struct {
	Overview        UserDashboardOverview `json:"overview"`        // 个人脉搏指标
	SkillUsageRank  []ToolUsageRankItem   `json:"skillUsageRank"`  // 近一周技能使用排行
	McpUsageRank    []ToolUsageRankItem   `json:"mcpUsageRank"`    // 近一周 MCP 使用排行
	ToolUsageRank   []ToolUsageRankItem   `json:"toolUsageRank"`   // 近一周工具使用排行
	StorageStats    UserStorageStats      `json:"storageStats"`    // 存储空间使用统计
}

// UserDashboardOverview 用户仪表盘脉搏指标
type UserDashboardOverview struct {
	TodayTokens     int64           `json:"todayTokens"`     // 今日 Token 消耗
	WeekTokens      int64           `json:"weekTokens"`      // 本周 Token 消耗
	TotalTokens     int64           `json:"totalTokens"`     // 历史累计 Token 消耗
	DailyTokenLimit int             `json:"dailyTokenLimit"` // 每日 Token 限额（单位 M，0=不限制）
	TotalSessions   int64           `json:"totalSessions"`   // 个人会话总数
	NewSessions     int64           `json:"newSessions"`     // 本周新增会话
	Sparkline       []SparklineItem `json:"sparkline"`       // 7日迷你趋势数据
}

// ──────────────────────────────────────────────
// 通用组件
// ──────────────────────────────────────────────

// SparklineItem 迷你趋势图数据点
// 无坐标轴、无网格线，仅用于前端绘制迷你面积图
type SparklineItem struct {
	Date   string `json:"date"`   // 日期 YYYY-MM-DD
	Tokens int64  `json:"tokens"` // 当日 Token 消耗
}

// TokenTrendItem Token 消耗趋势单日数据
// 按天聚合 prompt + completion token，支撑堆叠柱状图
type TokenTrendItem struct {
	Date             string `json:"date"`             // 日期 YYYY-MM-DD
	PromptTokens     int64  `json:"promptTokens"`     // 当日 prompt token
	CompletionTokens int64  `json:"completionTokens"` // 当日 completion token
}

// ModelUsageItem 模型使用分布项
// 按模型聚合 Token 消耗，支撑水平条形图
type ModelUsageItem struct {
	ModelName        string  `json:"modelName"`        // 模型名称（providerDisplayName - modelName）
	TokenCount       int64   `json:"tokenCount"`       // Token 消耗数
	Percentage       float64 `json:"percentage"`       // 占比百分比（0-100）
	PromptTokens     int64   `json:"promptTokens"`     // Prompt token 消耗数
	CompletionTokens int64   `json:"completionTokens"` // Completion token 消耗数
}

// UserTokenRankItem 用户 Token 消耗排行项
// 按用户聚合 Token 消耗，支撑水平条形图
type UserTokenRankItem struct {
	UserId     string  `json:"userId"`     // 用户 ID
	Username   string  `json:"username"`   // 用户名
	TokenCount int64   `json:"tokenCount"` // Token 消耗数
	Percentage float64 `json:"percentage"` // 占比百分比（0-100）
}

// ActiveUserTrendItem 活跃用户趋势单日数据
// 每日去重活跃用户数，支撑折线图
type ActiveUserTrendItem struct {
	Date  string `json:"date"`  // 日期 YYYY-MM-DD
	Count int64  `json:"count"` // 当日活跃用户数
}

// ToolUsageRankItem 工具/技能/MCP 使用排行项
// 按调用次数排序，支撑水平条形图
type ToolUsageRankItem struct {
	Name  string `json:"name"`  // 名称（技能名/MCP名/工具名）
	Count int64  `json:"count"` // 近7日调用次数
}

// UserStorageStats 用户存储空间使用统计
// usedBytes 为已用总量，freeBytes 为磁盘剩余可用空间，totalBytes 为磁盘总容量，items 为各子目录明细
type UserStorageStats struct {
	UsedBytes  int64              `json:"usedBytes"`  // 已用总量（字节）
	FreeBytes  int64              `json:"freeBytes"`  // 磁盘剩余可用空间（字节）
	TotalBytes int64              `json:"totalBytes"` // 磁盘总容量（字节）
	Items      []StorageUsageItem `json:"items"`      // 各子目录使用明细
}

// StorageUsageItem 存储空间使用统计项
// 按用户空间子目录分类，支撑存储占比展示
type StorageUsageItem struct {
	Name       string  `json:"name"`       // 目录名，如 documents、images
	BytesSize  int64   `json:"bytesSize"`  // 目录大小（字节）
	Percentage float64 `json:"percentage"` // 占磁盘总容量百分比（0-100）
}


// TokenTrendRsp Token 趋势响应
// 供前端切换时间粒度（7/30/90天）使用，admin 和 user 共用
type TokenTrendRsp struct {
	Items []TokenTrendItem `json:"items"` // 趋势数据
}

// ActiveUserTrendRsp 活跃用户趋势响应
// 供前端切换时间粒度（7/30/90天）使用
type ActiveUserTrendRsp struct {
	Items []ActiveUserTrendItem `json:"items"` // 趋势数据
}

// ModelUsageRsp 模型使用分布响应
// 供前端切换时间粒度（7/30/90天）使用，admin 和 user 共用
type ModelUsageRsp struct {
	Items []ModelUsageItem `json:"items"` // 模型使用分布数据
}

// UserTokenRankRsp 用户 Token 消耗排行响应
// 供前端切换时间粒度（7/30/90天）使用
type UserTokenRankRsp struct {
	Items []UserTokenRankItem `json:"items"` // 用户 Token 消耗排行
}

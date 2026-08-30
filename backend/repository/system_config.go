package repository

// SystemConfig 系统配置结构体，对应 system_setting 表的全量快照。
// 调用 repo.LoadConfig() 一次查全表填充此结构体，避免逐 key 读取。
type SystemConfig struct {
	GeneralDomain string // 本系统域名，用于生成文件外链和分享链接等

	ClawHubAPIURL string // ClawHub API 地址
	ClawHubToken  string // token（留空则不走加速）

	CompressThresholdPercent   int // 压缩触发的上下文阈值百分比，默认80(即模型上下文长度的80%)
	CompressKeepRounds         int // 压缩时保留最近几轮对话不压缩，默认3
	MaxIterations              int // agent最大步骤
	PruningTokenThreshold      int // 剪枝策略：总token阈值（单位K），默认96
	PruningMaxToolResultLength int // 工具结果最大长度（字符数），超过截断，默认2000
	PruningHeadTruncateLength  int // 截断时保留头部字符数，默认1000
	PruningTailTruncateLength  int // 截断时保留尾部字符数，默认1000

	LLMRequestDelayMs    int // LLM请求延迟（毫秒），用于避免限流，默认500
	FileLinkExpiresHours int // 文件分享外链有效期（小时），默认72

	KnowledgeEnableOCR bool // 是否启用 OCR 解析，默认 false

	WebFetchEnabled bool // 工具组：是否启用网页抓取，默认 true
	VisualEnabled   bool // 工具组：是否启用多模态识别，默认 false

	ShellTimeoutMinutes int // 工具组：Shell 命令执行超时（分钟），默认 5

	ModelMaxRetries       int // agent组：LLM 调用最大重试次数，默认 3
	ModelRateLimitWaitSec int // agent组：遇到 429 限流时的固定等待秒数，默认 8
	ModelBackoffBaseSec   int // agent组：退避重试的基秒数，第 N 次重试等待 N×BackoffBaseSec 秒，默认 3

	MainAgentTimeoutMinutes int // agent组：MainAgent 单次查询超时时间（分钟），默认 20
}

func NewDefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		CompressThresholdPercent:   80,
		CompressKeepRounds:         3,
		MaxIterations:              120,
		PruningTokenThreshold:      96,
		LLMRequestDelayMs:          500,
		PruningMaxToolResultLength: 2000,
		PruningHeadTruncateLength:  1000,
		PruningTailTruncateLength:  1000,
		FileLinkExpiresHours:       72,
		KnowledgeEnableOCR:         false,
		WebFetchEnabled:            true,
		VisualEnabled:              false,
		ShellTimeoutMinutes:        5,
		ModelMaxRetries:            3,
		ModelRateLimitWaitSec:      8,
		ModelBackoffBaseSec:        3,
		MainAgentTimeoutMinutes:  20,
	}
}

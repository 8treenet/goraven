package vo

import "time"

// SystemInfoRsp 系统信息完整快照响应
// 聚合返回概览、数据库、磁盘、MCP 健康状态、生态统计、插件等运维关键数据
type SystemInfoRsp struct {
	Overview    OverviewInfo    `json:"overview"`    // 系统概览（版本 + 语言 + 缓存 + 存储）
	Database    DatabaseInfo    `json:"database"`    // 数据库状态
	Disks       []DiskInfo      `json:"disks"`       // 系统磁盘挂载信息
	MCPHealth   []MCPHealthItem `json:"mcpHealth"`   // MCP 服务健康状态
	Ecosystem   EcosystemInfo   `json:"ecosystem"`   // 生态实体统计
	Plugins     []PluginInfo    `json:"plugins"`     // 已加载插件列表
	CollectedAt time.Time       `json:"collectedAt"` // 数据采集时间
}

// OverviewInfo 系统概览信息（版本 + 语言 + 缓存 + 存储占用）
type OverviewInfo struct {
	Version        string `json:"version"`        // GoRaven 版本号
	Language       string `json:"language"`       // 系统语言: "zh" / "en"
	CacheType      string `json:"cacheType"`      // 缓存类型: "local" / "redis"
	CacheMemory    string `json:"cacheMemory"`    // 缓存内存占用（Redis: used_memory_human, local: N items）
	Timezone       string `json:"timezone"`       // 系统时区，如 "Asia/Shanghai"
	UploadBytes    int64  `json:"uploadBytes"`    // 上传文件目录总大小（字节）
	TempBytes      int64  `json:"tempBytes"`      // 临时文件总大小（字节）
}

// DatabaseInfo 数据库运行状态
type DatabaseInfo struct {
	Type          string     `json:"type"`          // 数据库类型: sqlite / mysql / postgres
	Version       string     `json:"version"`       // 数据库版本号，如 "3.45.0" / "8.0.36" / "16.2"
	Name          string     `json:"name"`          // 数据库名（SQLite 为文件路径）
	DataSizeBytes int64      `json:"dataSizeBytes"` // 数据占用大小（字节）
	Pool          DBPoolInfo `json:"pool"`          // 连接池状态
}

// DBPoolInfo 数据库连接池统计
type DBPoolInfo struct {
	MaxOpenConnections int   `json:"maxOpenConnections"` // 最大连接数配置
	OpenConnections    int   `json:"openConnections"`    // 当前打开连接数（含 inUse + idle）
	InUse              int   `json:"inUse"`              // 正在使用的连接数
	Idle               int   `json:"idle"`               // 空闲连接数
	WaitCount          int64 `json:"waitCount"`          // 累计等待连接次数
	WaitDurationMs     int64 `json:"waitDurationMs"`     // 累计等待时长（毫秒）
	MaxIdleClosed      int64 `json:"maxIdleClosed"`      // 因超过最大空闲数而关闭的连接数
	MaxLifetimeClosed  int64 `json:"maxLifetimeClosed"`  // 因超过最大生命周期而关闭的连接数
}

// DiskInfo 系统磁盘挂载信息
type DiskInfo struct {
	MountPoint  string  `json:"mountPoint"`  // 挂载点路径
	FSType      string  `json:"fsType"`      // 文件系统类型
	Device      string  `json:"device"`      // 设备名，如 /dev/sda1
	TotalBytes  int64   `json:"totalBytes"`  // 总容量（字节）
	UsedBytes   int64   `json:"usedBytes"`   // 已用容量（字节）
	FreeBytes   int64   `json:"freeBytes"`   // 可用容量（字节）
	UsedPercent float64 `json:"usedPercent"` // 使用百分比
}

// EcosystemInfo 平台生态实体快速统计
type EcosystemInfo struct {
	TotalUsers       int64 `json:"totalUsers"`       // 用户总数（deleted=0）
	ActiveUsers      int64 `json:"activeUsers"`      // 启用用户数（status=1 AND deleted=0）
	AdminUsers       int64 `json:"adminUsers"`       // 管理员数（role=1 AND deleted=0）
	TotalModels      int64 `json:"totalModels"`      // 模型总数（deleted=0）
	EnabledModels    int64 `json:"enabledModels"`    // 启用模型数（status=1 AND deleted=0）
	TotalMcps        int64 `json:"totalMcps"`        // MCP 端点总数（deleted=0）
	EnabledMcps      int64 `json:"enabledMcps"`      // 启用 MCP 数（status=1 AND deleted=0）
	SystemSkills     int64 `json:"systemSkills"`     // 系统技能数（deleted=0）
	MarketSkills     int64 `json:"marketSkills"`     // 市场技能上架数（status=1 AND deleted=0）
	PersonaTemplates int64 `json:"personaTemplates"` // 角色模板数（deleted=0）
	TotalSessions    int64 `json:"totalSessions"`    // 会话总数（deleted=0）
	TotalMessages    int64 `json:"totalMessages"`    // 消息总数
	TotalTeamProjects int64 `json:"totalSharedProjects"` // 团队项目数
	// 分享统计
	TotalShareLinks  int64 `json:"totalShareLinks"`  // 分享链接总数（deleted=0）
	ActiveShareLinks int64 `json:"activeShareLinks"` // 有效分享链接数（未过期）
	TotalShareViews  int64 `json:"totalShareViews"`  // 分享链接总浏览次数
}

// PluginInfo 已加载插件信息
type PluginInfo struct {
	Name    string `json:"name"`    // 插件名称
	Version string `json:"version"` // 插件版本
}

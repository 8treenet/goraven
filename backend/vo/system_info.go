package vo

import "time"

type SystemInfoRsp struct {
	Overview    OverviewInfo    `json:"overview"`
	Database    DatabaseInfo    `json:"database"`
	Disks       []DiskInfo      `json:"disks"`
	MCPHealth   []MCPHealthItem `json:"mcpHealth"`
	Ecosystem   EcosystemInfo   `json:"ecosystem"`
	Plugins     []PluginInfo    `json:"plugins"`
	CollectedAt time.Time       `json:"collectedAt"`
}

type OverviewInfo struct {
	Version     string `json:"version"`
	Language    string `json:"language"`
	CacheType   string `json:"cacheType"`
	CacheMemory string `json:"cacheMemory"`
	Timezone    string `json:"timezone"`
	UploadBytes int64  `json:"uploadBytes"`
	TempBytes   int64  `json:"tempBytes"`
}

type DatabaseInfo struct {
	Type          string     `json:"type"`
	Version       string     `json:"version"`
	Name          string     `json:"name"`
	DataSizeBytes int64      `json:"dataSizeBytes"`
	Pool          DBPoolInfo `json:"pool"`
}

type DBPoolInfo struct {
	MaxOpenConnections int   `json:"maxOpenConnections"`
	OpenConnections    int   `json:"openConnections"`
	InUse              int   `json:"inUse"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"waitCount"`
	WaitDurationMs     int64 `json:"waitDurationMs"`
	MaxIdleClosed      int64 `json:"maxIdleClosed"`
	MaxLifetimeClosed  int64 `json:"maxLifetimeClosed"`
}

type DiskInfo struct {
	MountPoint  string  `json:"mountPoint"`
	FSType      string  `json:"fsType"`
	Device      string  `json:"device"`
	TotalBytes  int64   `json:"totalBytes"`
	UsedBytes   int64   `json:"usedBytes"`
	FreeBytes   int64   `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type EcosystemInfo struct {
	TotalUsers          int64 `json:"totalUsers"`
	ActiveUsers         int64 `json:"activeUsers"`
	AdminUsers          int64 `json:"adminUsers"`
	TotalModels         int64 `json:"totalModels"`
	EnabledModels       int64 `json:"enabledModels"`
	TotalMcps           int64 `json:"totalMcps"`
	EnabledMcps         int64 `json:"enabledMcps"`
	SystemSkills        int64 `json:"systemSkills"`
	MarketSkills        int64 `json:"marketSkills"`
	PersonaTemplates    int64 `json:"personaTemplates"`
	TotalSessions       int64 `json:"totalSessions"`
	TotalMessages       int64 `json:"totalMessages"`
	TotalSharedProjects int64 `json:"totalSharedProjects"`

	TotalShareLinks  int64 `json:"totalShareLinks"`
	ActiveShareLinks int64 `json:"activeShareLinks"`
	TotalShareViews  int64 `json:"totalShareViews"`
}

type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

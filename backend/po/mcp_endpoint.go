package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	MCPEndpointStatusDisabled = 0
	MCPEndpointStatusEnabled  = 1
)

// MCPEndpoint MCP 服务器端点配置表
type MCPEndpoint struct {
	McpId       int    `gorm:"primaryKey;column:mcp_id;type:int;autoIncrement"` // 主键ID
	Name        string `gorm:"uniqueIndex;column:name;type:varchar(64)"`        // 内部标识名,必须英文
	DisplayName string `gorm:"column:display_name;type:varchar(128)"`           // 展示名称
	Icon        string `gorm:"column:icon;type:varchar(256)"`                   // 图标：Lucide 图标名称或 URL
	Description string `gorm:"column:description;type:varchar(512)"`            // 功能描述

	Transport    string `gorm:"column:transport;type:varchar(16)"`  // 传输类型: Stdio / SSE / StreamableHttp
	HttpUrl      string `gorm:"column:http_url;type:varchar(512)"`  // StreamableHttp/SSE 服务地址
	HttpHeader   string `gorm:"column:http_header;type:text"`       // 请求头(JSON字符串)
	HttpProxyURL string `gorm:"column:proxy_url;type:varchar(256)"` // 代理地址，非必填，如 http://127.0.0.1:7890
	StdioType    string `gorm:"column:stdio_type;type:varchar(16)"` // Stdio 执行器类型: npx / uvx
	StdioEnv     string `gorm:"column:stdio_env;type:text"`         // 环境变量(JSON数组字符串)
	StdioArgs    string `gorm:"column:stdio_args;type:text"`        // 启动参数(JSON数组字符串)

	Status          uint8      `gorm:"column:status;default:1"`         // 状态: 0禁用 1启用
	AlwaysOn        int        `gorm:"column:always_on;default:0"`      // 始终启用：0关闭 1开启（管理员设置）
	Deleted         uint8      `gorm:"column:deleted;default:0"`        // 软删除: 0正常 1删除
	HealthLatency   int        `gorm:"column:health_latency;default:0"` // 健康检测延迟（毫秒），离线为0
	HealthCheckedAt *time.Time `gorm:"column:health_checked_at"`        // 最后健康检测时间
	Remark          string     `gorm:"column:remark;type:varchar(512)"` // 管理员备注
	Created         time.Time  `gorm:"not null;column:created"`
	Updated         time.Time  `gorm:"not null;column:updated"`
}

func (m *MCPEndpoint) TableName() string {
	return "mcp_endpoint"
}

func (m *MCPEndpoint) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.Created = now
	m.Updated = now
	return nil
}

func (m *MCPEndpoint) BeforeSave(tx *gorm.DB) error {
	m.Updated = time.Now()
	return nil
}

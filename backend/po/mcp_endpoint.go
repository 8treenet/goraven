package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	MCPEndpointStatusDisabled	= 0
	MCPEndpointStatusEnabled	= 1
)

type MCPEndpoint struct {
	McpId		int	`gorm:"primaryKey;column:mcp_id;type:int;autoIncrement"`
	Name		string	`gorm:"uniqueIndex;column:name;type:varchar(64)"`
	DisplayName	string	`gorm:"column:display_name;type:varchar(128)"`
	Icon		string	`gorm:"column:icon;type:varchar(256)"`
	Description	string	`gorm:"column:description;type:varchar(512)"`

	Transport	string	`gorm:"column:transport;type:varchar(16)"`
	HttpUrl		string	`gorm:"column:http_url;type:varchar(512)"`
	HttpHeader	string	`gorm:"column:http_header;type:text"`
	HttpProxyURL	string	`gorm:"column:proxy_url;type:varchar(256)"`
	StdioType	string	`gorm:"column:stdio_type;type:varchar(16)"`
	StdioEnv	string	`gorm:"column:stdio_env;type:text"`
	StdioArgs	string	`gorm:"column:stdio_args;type:text"`

	Status		uint8		`gorm:"column:status;default:1"`
	Deleted		uint8		`gorm:"column:deleted;default:0"`
	HealthLatency	int		`gorm:"column:health_latency;default:0"`
	HealthCheckedAt	*time.Time	`gorm:"column:health_checked_at"`
	Remark		string		`gorm:"column:remark;type:varchar(512)"`
	Created		time.Time	`gorm:"not null;column:created"`
	Updated		time.Time	`gorm:"not null;column:updated"`
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

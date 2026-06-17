package vo

import "time"

type AdminMCPListReq struct {
	Search    string `url:"search"`
	Transport string `url:"transport"`
	Status    *uint8 `url:"status"`
	Page      int    `url:"page"`
	PageSize  int    `url:"pageSize"`
}

type AdminMCPItem struct {
	McpId           int        `json:"mcpId"`
	Name            string     `json:"name"`
	DisplayName     string     `json:"displayName"`
	Icon            string     `json:"icon"`
	Description     string     `json:"description"`
	Transport       string     `json:"transport"`
	HttpUrl         string     `json:"httpUrl"`
	HttpHeader      string     `json:"httpHeader"`
	HttpProxyUrl    string     `json:"httpProxyUrl"`
	StdioType       string     `json:"stdioType"`
	StdioEnv        string     `json:"stdioEnv"`
	StdioArgs       string     `json:"stdioArgs"`
	Status          uint8      `json:"status"`
	HealthLatency   int        `json:"healthLatency"`
	HealthCheckedAt *time.Time `json:"healthCheckedAt"`
	Remark          string     `json:"remark"`
	Created         time.Time  `json:"created"`
	Updated         time.Time  `json:"updated"`
}

type AdminMCPDetailRsp struct {
	McpId           int        `json:"mcpId"`
	Name            string     `json:"name"`
	DisplayName     string     `json:"displayName"`
	Icon            string     `json:"icon"`
	Description     string     `json:"description"`
	Transport       string     `json:"transport"`
	HttpUrl         string     `json:"httpUrl"`
	HttpHeader      string     `json:"httpHeader"`
	HttpProxyUrl    string     `json:"httpProxyUrl"`
	StdioType       string     `json:"stdioType"`
	StdioEnv        string     `json:"stdioEnv"`
	StdioArgs       string     `json:"stdioArgs"`
	Status          uint8      `json:"status"`
	HealthLatency   int        `json:"healthLatency"`
	HealthCheckedAt *time.Time `json:"healthCheckedAt"`
	Remark          string     `json:"remark"`
	Created         time.Time  `json:"created"`
	Updated         time.Time  `json:"updated"`
}

type AdminCreateMCPReq struct {
	Name         string            `json:"name" validate:"required"`
	DisplayName  string            `json:"displayName" validate:"required"`
	Icon         string            `json:"icon"`
	Description  string            `json:"description"`
	Transport    string            `json:"transport" validate:"required"`
	HttpUrl      string            `json:"httpUrl"`
	HttpHeader   map[string]string `json:"httpHeader"`
	HttpProxyUrl string            `json:"httpProxyUrl"`
	StdioType    string            `json:"stdioType"`
	StdioEnv     map[string]string `json:"stdioEnv"`
	StdioArgs    []string          `json:"stdioArgs"`
	Remark       string            `json:"remark"`
}

// AdminMCPToggleAlwaysOnReq 切换 MCP 始终启用请求
type AdminMCPToggleAlwaysOnReq struct {
	AlwaysOn int `json:"alwaysOn" validate:"oneof=0 1"` // 始终启用：0关闭 1开启
}

type AdminUpdateMCPReq struct {
	DisplayName  string            `json:"displayName"`
	Icon         string            `json:"icon"`
	Description  string            `json:"description"`
	HttpUrl      string            `json:"httpUrl"`
	HttpHeader   map[string]string `json:"httpHeader"`
	HttpProxyUrl string            `json:"httpProxyUrl"`
	StdioType    string            `json:"stdioType"`
	StdioEnv     map[string]string `json:"stdioEnv"`
	StdioArgs    []string          `json:"stdioArgs"`
	Status       *uint8            `json:"status"`
	Remark       string            `json:"remark"`
}

type MCPRecommendItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Transport   string `json:"transport"`
	HttpUrl     string `json:"httpUrl"`
	HttpHeader  string `json:"httpHeader"`
	StdioType   string `json:"stdioType"`
	StdioArgs   string `json:"stdioArgs"`
	StdioEnv    string `json:"stdioEnv"`
	Installed   bool   `json:"installed"`
	McpId       int    `json:"mcpId"`
	McpStatus   uint8  `json:"mcpStatus"`
}

type MCPCodePreview struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Env     map[string]string `json:"env"`
}

type MCPRecommendDef struct {
	Name        string
	DisplayName string
	Icon        string
	Description string
	Transport   string
	HttpUrl     string
	StdioType   string
	StdioArgs   []string
}

type UserMCPItem struct {
	McpId       int    `json:"mcpId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

type MCPHealthItem struct {
	McpId           int        `json:"mcpId"`
	Name            string     `json:"name"`
	DisplayName     string     `json:"displayName"`
	Icon            string     `json:"icon"`
	HealthLatency   int        `json:"healthLatency"`
	HealthCheckedAt *time.Time `json:"healthCheckedAt"`
}

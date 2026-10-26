package vo

import "time"

// AdminMCPListReq 管理员 MCP 列表请求
type AdminMCPListReq struct {
	Search    string `url:"search"`    // 名称模糊搜索
	Transport string `url:"transport"` // 传输类型筛选: Stdio / SSE / StreamableHttp
	Status    *uint8 `url:"status"`    // 状态筛选: 0禁用 1启用，nil不筛选
	Page      int    `url:"page"`      // 页码
	PageSize  int    `url:"pageSize"`  // 每页条数
}

// AdminMCPItem 管理员 MCP 列表条目（敏感字段脱敏）
type AdminMCPItem struct {
	McpId           int        `json:"mcpId"`           // 主键ID
	Name            string     `json:"name"`            // 内部标识名，英文
	DisplayName     string     `json:"displayName"`     // 展示名称
	Icon            string     `json:"icon"`            // 图标
	Description     string     `json:"description"`     // 功能描述
	Transport       string     `json:"transport"`       // 传输类型: Stdio / SSE / StreamableHttp
	HttpUrl         string     `json:"httpUrl"`         // SSE/HTTP 服务地址
	HttpHeader      string     `json:"httpHeader"`      // 请求头（脱敏）
	HttpProxyUrl    string     `json:"httpProxyUrl"`    // 代理地址
	StdioType       string     `json:"stdioType"`       // Stdio 执行器类型: npx / uvx
	StdioEnv        string     `json:"stdioEnv"`        // 环境变量（脱敏）
	StdioArgs       string     `json:"stdioArgs"`       // 启动参数
	Status          uint8      `json:"status"`          // 状态: 0禁用 1启用
	AlwaysOn        int        `json:"alwaysOn"`        // 始终启用：0关闭 1开启
	HealthLatency   int        `json:"healthLatency"`   // 健康检测延迟（毫秒），离线为0
	HealthCheckedAt *time.Time `json:"healthCheckedAt"` // 最后健康检测时间
	Remark          string     `json:"remark"`          // 管理员备注
	Created         time.Time  `json:"created"`         // 创建时间
	Updated         time.Time  `json:"updated"`         // 更新时间
}

// AdminMCPDetailRsp 管理员 MCP 详情响应（包含原始敏感字段，用于编辑回填）
type AdminMCPDetailRsp struct {
	McpId           int        `json:"mcpId"`           // 主键ID
	Name            string     `json:"name"`            // 内部标识名，英文，不可修改
	DisplayName     string     `json:"displayName"`     // 展示名称
	Icon            string     `json:"icon"`            // 图标
	Description     string     `json:"description"`     // 功能描述
	Transport       string     `json:"transport"`       // 传输类型: Stdio / SSE / StreamableHttp，不可修改
	HttpUrl         string     `json:"httpUrl"`         // SSE/HTTP 服务地址
	HttpHeader      string     `json:"httpHeader"`      // 请求头（原始值）
	HttpProxyUrl    string     `json:"httpProxyUrl"`    // 代理地址
	StdioType       string     `json:"stdioType"`       // Stdio 执行器类型: npx / uvx
	StdioEnv        string     `json:"stdioEnv"`        // 环境变量（原始值）
	StdioArgs       string     `json:"stdioArgs"`       // 启动参数
	Status          uint8      `json:"status"`          // 状态: 0禁用 1启用
	AlwaysOn        int        `json:"alwaysOn"`        // 始终启用：0关闭 1开启
	HealthLatency   int        `json:"healthLatency"`   // 健康检测延迟（毫秒），离线为0
	HealthCheckedAt *time.Time `json:"healthCheckedAt"` // 最后健康检测时间
	Remark          string     `json:"remark"`          // 管理员备注
	Created         time.Time  `json:"created"`         // 创建时间
	Updated         time.Time  `json:"updated"`         // 更新时间
}

// AdminCreateMCPReq 管理员创建 MCP 请求
// 前端 httpHeader/stdioEnv 传 map[string]string，后端转为存储格式
type AdminCreateMCPReq struct {
	Name         string            `json:"name" validate:"required"`        // 内部标识名，英文，全局唯一
	DisplayName  string            `json:"displayName" validate:"required"` // 展示名称
	Icon         string            `json:"icon"`                            // 图标
	Description  string            `json:"description"`                     // 功能描述
	Transport    string            `json:"transport" validate:"required"`   // 传输类型: Stdio / SSE / StreamableHttp
	HttpUrl      string            `json:"httpUrl"`                         // SSE/HTTP 服务地址
	HttpHeader   map[string]string `json:"httpHeader"`                      // 请求头，前端传 map
	HttpProxyUrl string            `json:"httpProxyUrl"`                    // 代理地址，非必填
	StdioType    string            `json:"stdioType"`                       // Stdio 执行器类型: npx / uvx
	StdioEnv     map[string]string `json:"stdioEnv"`                        // 环境变量，前端传 map
	StdioArgs    []string          `json:"stdioArgs"`                       // 启动参数，如 ["@modelcontextprotocol/server-filesystem", "/tmp"]
	Remark       string            `json:"remark"`                          // 管理员备注
}

// AdminUpdateMCPReq 管理员编辑 MCP 请求（仅传需要修改的字段，name/transport 不可改）
type AdminUpdateMCPReq struct {
	DisplayName  string            `json:"displayName"`  // 展示名称
	Icon         string            `json:"icon"`         // 图标
	Description  string            `json:"description"`  // 功能描述
	HttpUrl      string            `json:"httpUrl"`      // SSE/HTTP 服务地址
	HttpHeader   map[string]string `json:"httpHeader"`   // 请求头，前端传 map
	HttpProxyUrl string            `json:"httpProxyUrl"` // 代理地址
	StdioType    string            `json:"stdioType"`    // Stdio 执行器类型: npx / uvx
	StdioEnv     map[string]string `json:"stdioEnv"`     // 环境变量，前端传 map
	StdioArgs    []string          `json:"stdioArgs"`    // 启动参数
	Status       *uint8            `json:"status"`       // 状态: 0禁用 1启用，nil不修改
	Remark       string            `json:"remark"`       // 管理员备注
}

// AdminMCPToggleAlwaysOnReq 切换 MCP 始终启用请求
type AdminMCPToggleAlwaysOnReq struct {
	AlwaysOn int `json:"alwaysOn" validate:"oneof=0 1"` // 始终启用：0关闭 1开启
}

// MCPRecommendItem 推荐 MCP 条目
type MCPRecommendItem struct {
	Name        string `json:"name"`        // 内部标识名，用于判断安装状态
	DisplayName string `json:"displayName"` // 展示名称
	Icon        string `json:"icon"`        // 图标
	Description string `json:"description"` // 功能描述
	Transport   string `json:"transport"`   // 传输类型
	HttpUrl     string `json:"httpUrl"`     // SSE/HTTP 服务地址（推荐模板）
	HttpHeader  string `json:"httpHeader"`  // SSE/HTTP 请求头（JSON 字符串）
	StdioType   string `json:"stdioType"`   // Stdio 执行器类型
	StdioArgs   string `json:"stdioArgs"`   // 启动参数
	StdioEnv    string `json:"stdioEnv"`    // 环境变量（KEY=VALUE 数组字符串）
	Installed   bool   `json:"installed"`   // 是否已安装（name 匹配 deleted=0 的记录）
	McpId       int    `json:"mcpId"`       // 已安装时的 mcpId，0表示未安装
	McpStatus   uint8  `json:"mcpStatus"`   // 已安装时的状态: 0禁用 1启用
}

// MCPCodePreview MCP 代码配置预览，前端展示用
type MCPCodePreview struct {
	Type    string            `json:"type"`    // 传输类型小写: stdio / sse / http
	Command string            `json:"command"` // Stdio 命令: npx 或 uvx
	Args    []string          `json:"args"`    // Stdio 参数
	Url     string            `json:"url"`     // SSE/HTTP 服务地址
	Headers map[string]string `json:"headers"` // SSE/HTTP 请求头
	Env     map[string]string `json:"env"`     // 环境变量
}

// MCPRecommendDef 推荐 MCP 定义
type MCPRecommendDef struct {
	Name        string   // 内部标识名
	DisplayName string   // 展示名称
	Icon        string   // 图标
	Description string   // 功能描述
	Transport   string   // 传输类型: Stdio / SSE / StreamableHttp
	HttpUrl     string   // SSE/HTTP 服务地址
	StdioType   string   // Stdio 执行器类型: npx / uvx
	StdioArgs   []string // 启动参数
}

// UserMCPItem 用户可选 MCP 列表条目
type UserMCPItem struct {
	McpId       int    `json:"mcpId"`       // 主键ID
	Name        string `json:"name"`        // 内部标识名，英文
	DisplayName string `json:"displayName"` // 展示名称
	Icon        string `json:"icon"`        // 图标
	Description string `json:"description"` // 功能描述
}

// MCPHealthItem MCP 服务健康状态条目，用于系统信息页展示
type MCPHealthItem struct {
	McpId           int        `json:"mcpId"`           // 主键ID
	Name            string     `json:"name"`            // 内部标识名
	DisplayName     string     `json:"displayName"`     // 展示名称
	Icon            string     `json:"icon"`            // 图标
	HealthLatency   int        `json:"healthLatency"`   // 健康检测延迟（毫秒），离线为0
	HealthCheckedAt *time.Time `json:"healthCheckedAt"` // 最后健康检测时间
}

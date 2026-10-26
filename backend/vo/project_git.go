package vo

import "time"

// --- 项目 Git 集成 ---

// 新建项目 Git 来源模式
const (
	GitModeClone = "clone" // 从远程仓库克隆（缺省值，向后兼容）
	GitModeLocal = "local" // 本地仓库：仅本地版本管理，无远程
)

// GitCloneReq 新建项目时的 Git 来源参数（可选）
type GitCloneReq struct {
	Mode          string `json:"mode"`          // clone=克隆远程（缺省）；local=本地仓库
	RemoteUrl     string `json:"remoteUrl"`     // 远程仓库地址（不允许内嵌凭据），local 模式必须为空
	AuthType      uint8  `json:"authType"`      // 0=公开 1=SSH 2=HTTPS
	SshPrivateKey string `json:"sshPrivateKey"` // SSH 私钥（只写不读）
	HttpsUsername string `json:"httpsUsername"` // HTTPS 用户名
	HttpsSecret   string `json:"httpsSecret"`   // HTTPS 密码/token（只写不读）
	Shallow       *bool  `json:"shallow"`       // 浅克隆，默认 true；仅 clone 模式生效
}

// IsLocal 是否本地仓库模式
func (req *GitCloneReq) IsLocal() bool {
	return req != nil && req.Mode == GitModeLocal
}

// ShallowEnabled 浅克隆开关，未传默认开启。
func (req *GitCloneReq) ShallowEnabled() bool {
	if req == nil || req.Shallow == nil {
		return true
	}
	return *req.Shallow
}

// GitRemoteInfo 远程仓库回显（绝不包含密钥）
type GitRemoteInfo struct {
	Url           string `json:"url"`
	AuthType      uint8  `json:"authType"`
	HttpsUsername string `json:"httpsUsername"`
	HasCredential bool   `json:"hasCredential"`
}

// GitChange 工作区文件变更
type GitChange struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M/A/D/U/R/C
	Size   int64  `json:"size,omitempty"`
}

// GitStatusRsp 状态聚合响应
type GitStatusRsp struct {
	Enabled      bool          `json:"enabled"` // 是否 Git 项目（创建时决定，不可变）
	Initialized  bool          `json:"initialized"`
	CloneState   uint8         `json:"cloneState"`
	CloneMessage string        `json:"cloneMessage"`
	Remote       GitRemoteInfo `json:"remote"`
	Branch       string        `json:"branch"`
	Ahead        int           `json:"ahead"`
	Behind       int           `json:"behind"`
	Changes      []GitChange   `json:"changes"`
	Skipped      []GitChange   `json:"skipped"`
}

// GitTestReq 测试连接请求，参数可覆盖已保存配置（测试连接不落库）
type GitTestReq struct {
	RemoteUrl     string  `json:"remoteUrl"`
	AuthType      *uint8  `json:"authType"`
	SshPrivateKey *string `json:"sshPrivateKey"`
	HttpsUsername *string `json:"httpsUsername"`
	HttpsSecret   *string `json:"httpsSecret"`
}

// GitTestRsp 测试连接响应
type GitTestRsp struct {
	Head string `json:"head"` // 远端 HEAD 提交
}

// GitCommitReq 手动提交请求
type GitCommitReq struct {
	Message string `json:"message"`
	Push    bool   `json:"push"` // 提交后是否直接推送
}

// GitCommitRsp 手动提交响应
type GitCommitRsp struct {
	Hash    string      `json:"hash"`
	Pushed  bool        `json:"pushed"`
	Skipped []GitChange `json:"skipped"`
}

// GitUnrelatedReq 历史无关处理选择
type GitUnrelatedReq struct {
	Action string `json:"action"` // merge | local_only
}

// GitLogReq 提交历史请求
type GitLogReq struct {
	Limit  int `url:"limit"`
	Offset int `url:"offset"`
}

// GitCommitInfo 提交历史项
type GitCommitInfo struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Time      time.Time `json:"time"`
}

// GitLogRsp 提交历史响应
type GitLogRsp struct {
	Items []GitCommitInfo `json:"items"`
}

// GitDiffReq diff 请求（无 commit 时为工作区 vs HEAD）
type GitDiffReq struct {
	Path   string `url:"path"`
	Commit string `url:"commit"`
}

// GitDiffRsp diff 响应
type GitDiffRsp struct {
	Path   string `json:"path"`
	Commit string `json:"commit"`
	Diff   string `json:"diff"`
}

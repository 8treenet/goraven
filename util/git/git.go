// Package git 提供项目 Git 集成的底层能力：命令执行、受控环境构建、
// 仓库读写操作与克隆取消注册表。
// 本包不依赖数据库与 Web 层，业务编排（权限、配置持久化、定时调度）在 backend/service。
package git

import (
	"os"
	"path/filepath"
	"time"
)

const (
	LargeFileLimit    = int64(50 * 1024 * 1024) // 单文件提交阈值 50MB
	OpTimeout         = 5 * time.Minute         // 常规 git 操作超时
	CloneTimeout      = 30 * time.Minute        // 克隆超时
	StatusMaxChanges  = 500                     // 状态列表截断行数
	DiffMaxChars      = 200000                  // diff 输出上限
	DefaultBotName    = "goraven-bot"
	DefaultBotEmail   = "bot@goraven.dev"
	InitCommitMessage = "chore: initialize repository"
)

// DefaultGitignoreContent 不存在时生成的保守默认 .gitignore 内容。
const DefaultGitignoreContent = `node_modules/
__pycache__/
.venv/
venv/
*.log
.DS_Store
.env
.env.*
*.pem
*.key
`

// 认证方式常量，与 backend/po 的 GitAuth* 数值保持一致。
const (
	AuthNone  uint8 = 0
	AuthSSH   uint8 = 1
	AuthHTTPS uint8 = 2
)

// Change 工作区文件变更
type Change struct {
	Path   string
	Status string // M/A/D/U/R/C
	Size   int64
	Source string // 重命名/复制的源路径
}

// Identity 提交身份
type Identity struct {
	AuthorName     string
	AuthorEmail    string
	CommitterName  string
	CommitterEmail string
}

// BotIdentity 机器人提交身份（固定常量）
var BotIdentity = Identity{
	AuthorName:     DefaultBotName,
	AuthorEmail:    DefaultBotEmail,
	CommitterName:  DefaultBotName,
	CommitterEmail: DefaultBotEmail,
}

// CommitInfo 提交历史项
type CommitInfo struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Time      time.Time
	Message   string
}

// AuthSetting 一次 git 操作的认证参数，由调用方从持久化配置转换而来。
type AuthSetting struct {
	AuthType      uint8
	SshPrivateKey string
	HttpsUsername string
	HttpsSecret   string
	OwnerType     uint8
	OwnerId       int
}

// IsRepository 判断目录是否已是 git 仓库（Agent 手动 clone 的场景直接采纳）。
func IsRepository(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// DirOrTemp 目录缺失时退化为临时目录，保证只读远程命令可执行。
func DirOrTemp(dir string) string {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return os.TempDir()
}

// EnsureDefaultGitignore 不存在时生成保守默认 .gitignore，已存在绝不覆盖。
func EnsureDefaultGitignore(dir string) (bool, error) {
	path := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(path, []byte(DefaultGitignoreContent), 0644); err != nil {
		return false, err
	}
	return true, nil
}

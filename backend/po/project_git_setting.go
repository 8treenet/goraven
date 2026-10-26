package po

import (
	"time"

	"gorm.io/gorm"
)

// Git owner 类型常量
const (
	GitOwnerUserProject uint8 = 1 // 个人项目
	GitOwnerTeamProject uint8 = 2 // 团队项目
)

// Git 认证方式常量
const (
	GitAuthNone  uint8 = 0 // 公开仓库，不注入凭据
	GitAuthSSH   uint8 = 1 // SSH 私钥
	GitAuthHTTPS uint8 = 2 // HTTPS 用户名 + 密码/token
)

// Git 克隆状态常量
const (
	GitCloneNone    uint8 = 0 // 非克隆创建
	GitCloneRunning uint8 = 1 // 克隆中
	GitCloneSuccess uint8 = 2 // 克隆完成
	GitCloneFailed  uint8 = 3 // 克隆失败
)

// ProjectGitSetting 项目 Git 配置（个人项目与团队项目共用一张表，owner_type 区分）。
// 是否存在配置行在项目创建时确定，即"是否为 Git 项目"，此后不可开启或关闭。
// 凭据明文存储，与平台现有约定一致（ai_model.api_key、mcp_endpoint.http_header 等均为明文），
// 仅在接口层脱敏、永不回显。
type ProjectGitSetting struct {
	Id            int       `gorm:"primaryKey;column:id;type:int;autoIncrement"`                               // 主键ID
	OwnerType     uint8     `gorm:"column:owner_type;type:tinyint;uniqueIndex:idx_project_git_owner;not null"` // owner类型: 1个人项目 2团队项目
	OwnerId       int       `gorm:"column:owner_id;type:int;uniqueIndex:idx_project_git_owner;not null"`       // owner ID: owner_type=1为user_project.id, owner_type=2为team_project.id
	RemoteUrl     string    `gorm:"column:remote_url;type:varchar(2048)"`                                      // 远程仓库地址，空表示仅本地版本管理
	AuthType      uint8     `gorm:"column:auth_type;type:tinyint;default:0;not null"`                          // 认证方式: 0公开仓库不注入凭据 1SSH私钥 2HTTPS用户名+密码/token
	SshPrivateKey string    `gorm:"column:ssh_private_key;type:text"`                                          // SSH私钥，auth_type=1时使用
	HttpsUsername string    `gorm:"column:https_username;type:varchar(255)"`                                   // HTTPS用户名，auth_type=2时使用
	HttpsSecret   string    `gorm:"column:https_secret;type:text"`                                             // HTTPS密码/token，auth_type=2时使用
	CloneState    uint8     `gorm:"column:clone_state;type:tinyint;default:0;not null"`                        // 克隆状态: 0非克隆创建 1克隆中 2克隆完成 3克隆失败
	CloneMessage  string    `gorm:"column:clone_message;type:text"`                                            // 克隆结果信息，失败时记录错误原因
	Created       time.Time `gorm:"not null;column:created"`                                                   // 创建时间
	Updated       time.Time `gorm:"not null;column:updated"`                                                   // 更新时间
}

// TableName 项目 Git 配置表名
func (p *ProjectGitSetting) TableName() string {
	return "project_git_setting"
}

// BeforeCreate 写入创建与更新时间
func (p *ProjectGitSetting) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.Created = now
	p.Updated = now
	return nil
}

// BeforeSave 刷新更新时间
func (p *ProjectGitSetting) BeforeSave(tx *gorm.DB) error {
	p.Updated = time.Now()
	return nil
}

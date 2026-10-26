package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthEnv 一次 git 操作使用的受控环境与清理函数
type AuthEnv struct {
	Env     []string
	Cleanup func()
}

const askpassFilename = "askpass.sh"

// askpassScript 静态 askpass 脚本：内容固定、无用户输入、无注入面。
// 用户名与密码通过环境变量传入，不落 argv、不落 remote URL、不写 .git-credentials。
const askpassScript = `#!/bin/sh
case "$1" in
*Username*|*username*) printf '%s' "$GORAVEN_GIT_USERNAME" ;;
*) printf '%s' "$GORAVEN_GIT_PASSWORD" ;;
esac
`

// baseEnv 构造与宿主环境隔离的受控环境：忽略系统/全局 git 配置，HOME 指向一次性目录。
func baseEnv(home string) []string {
	return []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + home,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C",
		"LANG=C",
	}
}

// BuildBaseEnv 创建仅含基础隔离环境（无凭据）的操作环境，调用方必须执行 Cleanup。
func BuildBaseEnv() (*AuthEnv, error) {
	tmpDir, err := os.MkdirTemp("", "goraven-git-")
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(tmpDir, 0700); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}
	return &AuthEnv{
		Env:     baseEnv(tmpDir),
		Cleanup: func() { os.RemoveAll(tmpDir) },
	}, nil
}

// BuildAuthEnv 按认证方式注入凭据环境；SSH 私钥写入本次操作的一次性目录（0600），
// known_hosts 按项目持久化，操作结束清理临时目录。
func BuildAuthEnv(secretDir string, setting *AuthSetting) (*AuthEnv, error) {
	auth, err := BuildBaseEnv()
	if err != nil {
		return nil, err
	}
	homeDir := ""
	for _, entry := range auth.Env {
		if strings.HasPrefix(entry, "HOME=") {
			homeDir = strings.TrimPrefix(entry, "HOME=")
		}
	}

	fail := func(err error) (*AuthEnv, error) {
		auth.Cleanup()
		return nil, err
	}

	switch setting.AuthType {
	case AuthNone:
		knownHosts, err := prepareKnownHosts(secretDir, setting)
		if err != nil {
			return fail(err)
		}
		auth.Env = append(auth.Env, "GIT_SSH_COMMAND="+buildSSHCommand("", knownHosts))
	case AuthSSH:
		key := normalizePrivateKey(setting.SshPrivateKey)
		if key == "" {
			return fail(ErrNotConfigured)
		}
		keyPath := filepath.Join(homeDir, "id_key")
		if err := os.WriteFile(keyPath, []byte(key), 0600); err != nil {
			return fail(err)
		}
		knownHosts, err := prepareKnownHosts(secretDir, setting)
		if err != nil {
			return fail(err)
		}
		auth.Env = append(auth.Env, "GIT_SSH_COMMAND="+buildSSHCommand(keyPath, knownHosts))
	case AuthHTTPS:
		if strings.TrimSpace(setting.HttpsSecret) == "" {
			return fail(ErrNotConfigured)
		}
		script, err := ensureAskpassScript(secretDir)
		if err != nil {
			return fail(err)
		}
		auth.Env = append(auth.Env,
			"GIT_ASKPASS="+script,
			"GORAVEN_GIT_USERNAME="+setting.HttpsUsername,
			"GORAVEN_GIT_PASSWORD="+setting.HttpsSecret,
		)
	default:
		return fail(ErrNotConfigured)
	}

	// 注入机器人身份：rebase/merge 等隐式产生提交的操作需要 committer 身份，
	// 手动提交时由 WithEnv 覆盖 author。
	auth.Env = append(auth.Env,
		"GIT_AUTHOR_NAME="+DefaultBotName,
		"GIT_AUTHOR_EMAIL="+DefaultBotEmail,
		"GIT_COMMITTER_NAME="+DefaultBotName,
		"GIT_COMMITTER_EMAIL="+DefaultBotEmail,
	)
	return auth, nil
}

// WithEnv 返回覆盖/追加环境变量的副本；同名 key 原地替换，保证语义确定。
func WithEnv(env []string, kv ...string) []string {
	result := make([]string, len(env))
	copy(result, env)
	for i := 0; i+1 < len(kv); i += 2 {
		key, value := kv[i], kv[i+1]
		replaced := false
		for j, entry := range result {
			if strings.HasPrefix(entry, key+"=") {
				result[j] = key + "=" + value
				replaced = true
				break
			}
		}
		if !replaced {
			result = append(result, key+"="+value)
		}
	}
	return result
}

// prepareKnownHosts 返回项目级持久化 known_hosts 路径并确保文件存在。
func prepareKnownHosts(secretDir string, setting *AuthSetting) (string, error) {
	path := ProjectKnownHostsPath(secretDir, setting.OwnerType, setting.OwnerId)
	if err := ensureKnownHostsFile(path); err != nil {
		return "", err
	}
	return path, nil
}

// buildSSHCommand 构造与宿主隔离的 ssh 命令。
// -F /dev/null 忽略系统与用户 ssh 配置；IdentityFile=none + IdentityAgent=none + IdentitiesOnly=yes
// 确保不会误用宿主 ~/.ssh 或 ssh-agent 中的密钥（OpenSSH 按系统账户 home 解析 ~/.ssh，仅隔离 HOME 无效）。
// keyPath 为空表示公开仓库匿名访问；非空时通过 -i 使用本次凭据。
func buildSSHCommand(keyPath, knownHosts string) string {
	args := []string{"ssh", "-F", os.DevNull}
	if keyPath != "" {
		args = append(args, "-i", shellQuote(keyPath))
	}
	args = append(args,
		"-o", "IdentitiesOnly=yes",
		"-o", "IdentityFile=none",
		"-o", "IdentityAgent=none",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile="+shellQuote(knownHosts),
	)
	return strings.Join(args, " ")
}

// normalizePrivateKey 规范化前端粘贴的 SSH 私钥，保证 OpenSSH 可解析：
// 统一 CRLF/CR 为 LF、去除每行首尾空白（含粘贴缩进）、修剪整体空白并补齐结尾换行。
// OpenSSH 对格式要求严格：缺少结尾换行、CRLF、行首空格都会导致 "invalid format" 鉴权失败。
func normalizePrivateKey(raw string) string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	normalized = strings.TrimSpace(strings.Join(lines, "\n"))
	if normalized == "" {
		return ""
	}
	return normalized + "\n"
}

// ensureAskpassScript 幂等生成静态 askpass 脚本（0700）
func ensureAskpassScript(secretDir string) (string, error) {
	if err := os.MkdirAll(secretDir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(secretDir, askpassFilename)
	if err := os.WriteFile(path, []byte(askpassScript), 0700); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0700); err != nil {
		return "", err
	}
	return path, nil
}

// ProjectKnownHostsPath 项目级持久 known_hosts 路径，用于主机指纹校验（指纹变化时报错，防中间人）
func ProjectKnownHostsPath(secretDir string, ownerType uint8, ownerId int) string {
	return filepath.Join(secretDir, "known_hosts", fmt.Sprintf("%d_%d", ownerType, ownerId))
}

// RemoveKnownHosts 删除项目级持久化凭据文件（known_hosts），删除项目时级联调用。
func RemoveKnownHosts(secretDir string, ownerType uint8, ownerId int) {
	_ = os.RemoveAll(ProjectKnownHostsPath(secretDir, ownerType, ownerId))
}

func ensureKnownHostsFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}

// shellQuote 为 GIT_SSH_COMMAND 中的路径做单引号转义（git 通过 shell 拆分该字符串）
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

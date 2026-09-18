package git

import (
	"errors"
	"regexp"
	"strings"
)

// 领域哨兵错误：调用方（backend/service）负责将其映射为平台错误。
var (
	ErrBinaryMissing    = errors.New("git binary is not available")
	ErrInvalidRemoteURL = errors.New("git: invalid remote url")
	ErrNoChanges        = errors.New("git: no changes to commit")
	ErrDirtyTree        = errors.New("git: working tree has uncommitted changes")
	ErrNotConfigured    = errors.New("git: remote is not configured")
	ErrPushRejected     = errors.New("git: push rejected")
	ErrUnrelatedHistory = errors.New("git: unrelated histories")
	ErrConflict         = errors.New("git: rebase/merge conflict")
	ErrAuthFailed       = errors.New("git: authentication failed")
)

// CommandError git 命令执行失败，保留 stderr 供上层脱敏后归类。
type CommandError struct {
	Stderr string
	Err    error
}

func (e *CommandError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "git command failed"
}

func (e *CommandError) Unwrap() error { return e.Err }

var (
	gitUrlUserinfoRe     = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^/@\s]*:[^/@\s]*@`)
	gitTransportHelperRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+::`)
	gitUrlUserinfoAnyRe  = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://)[^/@\s]+@`)
)

// ValidateRemoteURL 校验远程地址：拒绝空值、选项注入、换行、传输助手（ext:: 等）与内嵌凭据。
func ValidateRemoteURL(raw string) error {
	url := strings.TrimSpace(raw)
	if url == "" {
		return ErrInvalidRemoteURL
	}
	if strings.HasPrefix(url, "-") {
		return ErrInvalidRemoteURL
	}
	if strings.ContainsAny(url, "\n\r\x00\t") {
		return ErrInvalidRemoteURL
	}
	if gitTransportHelperRe.MatchString(url) {
		return ErrInvalidRemoteURL
	}
	if gitUrlUserinfoRe.MatchString(url) {
		return ErrInvalidRemoteURL
	}
	return nil
}

// SanitizeMessage 错误消息统一脱敏：剥离 URL userinfo、掩码已知密钥，并限制长度。
func SanitizeMessage(msg string, secrets ...string) string {
	out := gitUrlUserinfoAnyRe.ReplaceAllString(msg, "${1}***@")
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "***")
	}
	if len(out) > 4000 {
		out = out[:4000]
	}
	return out
}

// IsPushRejected 判断 push 是否因非快进被拒
func IsPushRejected(stderr string) bool {
	s := strings.ToLower(stderr)
	return strings.Contains(s, "non-fast-forward") ||
		strings.Contains(s, "fetch first") ||
		strings.Contains(s, "[rejected]") ||
		strings.Contains(s, "remote contains work that you do not have")
}

// IsUnrelatedHistoryError 判断是否为历史无关错误
func IsUnrelatedHistoryError(stderr string) bool {
	return strings.Contains(strings.ToLower(stderr), "unrelated histories")
}

// IsConflict 判断是否产生 rebase/merge 冲突
func IsConflict(stderr string) bool {
	s := strings.ToLower(stderr)
	return strings.Contains(s, "conflict") || strings.Contains(s, "could not apply")
}

// IsAuthError 判断是否为鉴权失败
func IsAuthError(stderr string) bool {
	s := strings.ToLower(stderr)
	patterns := []string{
		"authentication failed",
		"could not read username",
		"could not read password",
		"permission denied (publickey)",
		"permission denied (public key)",
		"invalid username or password",
		"returned error: 401",
		"returned error: 403",
		"access denied",
		"no such identity",
		"host key verification failed",
	}
	for _, pattern := range patterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

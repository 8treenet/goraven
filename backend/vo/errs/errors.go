package errs

import (
	"fmt"

	"goraven/config"
)

type Error struct {
	enMessage string
	zhMessage string
}

func (e *Error) Error() string {
	if config.Get().GetLanguage() == "en" {
		return e.enMessage
	}
	return e.zhMessage
}

type FormatError struct {
	enFormat string
	zhFormat string
	args     []interface{}
}

func (e *FormatError) Error() string {
	if config.Get().GetLanguage() == "en" {
		return fmt.Sprintf(e.enFormat, e.args...)
	}
	return fmt.Sprintf(e.zhFormat, e.args...)
}

func NewFormatError(enFormat, zhFormat string, args ...interface{}) *FormatError {
	return &FormatError{enFormat: enFormat, zhFormat: zhFormat, args: args}
}

var (
	ErrInvalidCredentials = &Error{
		enMessage: "invalid username or password",
		zhMessage: "用户名或密码错误",
	}
	ErrUsernameTooShort = &Error{
		enMessage: "username must be at least 8 characters",
		zhMessage: "用户名至少 8 位",
	}
	ErrUserNotFound = &Error{
		enMessage: "user not found",
		zhMessage: "用户不存在",
	}
	ErrUsernameAlreadyExists = &Error{
		enMessage: "username already exists",
		zhMessage: "用户名已存在",
	}
	ErrInvalidUsername = &Error{
		enMessage: "username must be 8-16 chars, start and end with a letter or digit, and contain only letters, digits, '_' or '-'",
		zhMessage: "用户名需 8-16 位，以字母或数字开头和结尾，仅包含字母、数字、_、-",
	}
	ErrCannotEditSuperAdmin = &Error{
		enMessage: "cannot edit super admin",
		zhMessage: "无法编辑超级管理员",
	}
	ErrCannotResetSuperAdminPassword = &Error{
		enMessage: "cannot reset super admin password",
		zhMessage: "无法重置超级管理员密码",
	}
	ErrCannotDeleteSuperAdmin = &Error{
		enMessage: "cannot delete super admin",
		zhMessage: "无法删除超级管理员",
	}
	ErrModelNotFound = &Error{
		enMessage: "model not found",
		zhMessage: "模型不存在",
	}
	ErrModelTestFailed = &Error{
		enMessage: "model test failed",
		zhMessage: "模型测试失败",
	}
	ErrCannotDeleteDefaultModel = &Error{
		enMessage: "cannot delete the default model",
		zhMessage: "默认模型不可删除",
	}
	ErrCannotDeleteFlashModel = &Error{
		enMessage: "cannot delete the flash model. Please set another model as the flash model first",
		zhMessage: "Flash 模型不可删除，请先将其他模型设置为 Flash 模型",
	}
	ErrCannotDeleteVisualModel = &Error{
		enMessage: "cannot delete the multimodal recognition model. Please set another model as the visual model first",
		zhMessage: "多模态识别模型不可删除，请先将其他模型设置为多模态识别模型",
	}
	ErrCannotDisableDefaultModel = &Error{
		enMessage: "cannot disable the default model",
		zhMessage: "默认模型不可停用",
	}
	ErrProviderNotFound = &Error{
		enMessage: "provider not found",
		zhMessage: "供应商不存在",
	}
	ErrAPIKeyRequired = &Error{
		enMessage: "API key is required",
		zhMessage: "API Key 不能为空",
	}
	ErrBaseURLRequired = &Error{
		enMessage: "base URL is required",
		zhMessage: "Base URL 不能为空",
	}
	ErrMCPNotFound = &Error{
		enMessage: "MCP endpoint not found",
		zhMessage: "MCP 端点不存在",
	}
	ErrMCPNameAlreadyExists = &Error{
		enMessage: "MCP name already exists",
		zhMessage: "MCP 名称已存在",
	}
	ErrMCPTestFailed = &Error{
		enMessage: "MCP connection test failed",
		zhMessage: "MCP 连接测试失败",
	}
	ErrMCPTransportRequired = &Error{
		enMessage: "transport type is required",
		zhMessage: "传输类型不能为空",
	}
	ErrMCPNameRequired = &Error{
		enMessage: "MCP name is required",
		zhMessage: "MCP 名称不能为空",
	}
	ErrMCPNameInvalid = &Error{
		enMessage: "MCP name must start with a lowercase letter and contain only lowercase letters, digits, and hyphens",
		zhMessage: "MCP 名称必须以小写字母开头，仅允许小写字母、数字和连字符",
	}
	ErrMCPDisplayNameRequired = &Error{
		enMessage: "MCP display name is required",
		zhMessage: "MCP 显示名称不能为空",
	}
	ErrMCPHttpURLRequired = &Error{
		enMessage: "HTTP URL is required for SSE/HTTP transport",
		zhMessage: "SSE/HTTP 传输类型需要填写服务地址",
	}
	ErrMCPStdioArgsRequired = &Error{
		enMessage: "stdio args is required for Stdio transport",
		zhMessage: "Stdio 传输类型需要填写启动参数",
	}
	ErrSystemSkillNotFound = &Error{
		enMessage: "system skill not found",
		zhMessage: "系统技能不存在",
	}
	ErrSystemSkillNameAlreadyExists = &Error{
		enMessage: "system skill name already exists",
		zhMessage: "系统技能名称已存在",
	}
	ErrSystemSkillContentRequired = &Error{
		enMessage: "skill content is required",
		zhMessage: "技能内容不能为空",
	}
	ErrSystemSkillInvalidFormat = &Error{
		enMessage: "invalid skill format: missing frontmatter",
		zhMessage: "技能格式无效：缺少 frontmatter",
	}
	ErrSystemSkillInvalidName = &Error{
		enMessage: "skill name must start with 'goraven-'",
		zhMessage: "技能名称必须以 'goraven-' 开头",
	}
	ErrSystemSkillNameRequired = &Error{
		enMessage: "skill name is required in frontmatter",
		zhMessage: "frontmatter 中技能名称不能为空",
	}
	ErrPresetSkillCannotModify = &Error{
		enMessage: "preset system skills cannot be modified",
		zhMessage: "GoRaven 预设技能不可修改",
	}
	ErrPresetSkillCannotDelete = &Error{
		enMessage: "preset system skills cannot be deleted",
		zhMessage: "GoRaven 预设技能不可删除",
	}
	ErrMarketSkillNotFound = &Error{
		enMessage: "market skill not found",
		zhMessage: "市场技能不存在",
	}
	ErrMarketSkillNameAlreadyExists = &Error{
		enMessage: "market skill name already exists",
		zhMessage: "市场技能名称已存在",
	}
	ErrClawHubImportFailed = &Error{
		enMessage: "clawhub import failed",
		zhMessage: "ClawHub 导入失败",
	}
	ErrSkillPublishInvalidZip = &Error{
		enMessage: "invalid zip file",
		zhMessage: "无效的 zip 文件",
	}
	ErrSkillPublishNoSkillMd = &Error{
		enMessage: "SKILL.md not found in zip",
		zhMessage: "zip 中未找到 SKILL.md",
	}
	ErrSkillPublishTooLarge = &Error{
		enMessage: "skill zip exceeds maximum size (50MB)",
		zhMessage: "技能包解压后超过最大限制 (50MB)",
	}
	ErrSkillUploadNotFound = &Error{
		enMessage: "upload task not found",
		zhMessage: "上传任务不存在",
	}
	ErrSkillUploadNotCompleted = &Error{
		enMessage: "upload not completed yet",
		zhMessage: "上传尚未完成",
	}
	ErrSkillCategoryNotFound = &Error{
		enMessage: "skill category not found",
		zhMessage: "技能分类不存在",
	}
	ErrSkillCategoryIsDefault = &Error{
		enMessage: "default category cannot be deleted or edited",
		zhMessage: "默认分类不可删除或编辑",
	}
	ErrSkillCategoryNameRequired = &Error{
		enMessage: "category name is required",
		zhMessage: "分类名称不能为空",
	}
	ErrSkillCategoryRequired = &Error{
		enMessage: "category is required",
		zhMessage: "请选择分类",
	}
	ErrPersonaTemplateNotFound = &Error{
		enMessage: "persona template not found",
		zhMessage: "角色模板不存在",
	}
	ErrPersonaTemplateRoleInfoRequired = &Error{
		enMessage: "role info is required",
		zhMessage: "系统提示词不能为空",
	}
	ErrPersonaTemplateRoleInfoTooLong = &Error{
		enMessage: "system prompt cannot exceed 500 characters",
		zhMessage: "系统提示词不能超过500字",
	}
	ErrPersonaCategoryNotFound = &Error{
		enMessage: "persona category not found",
		zhMessage: "角色分类不存在",
	}
	ErrPersonaCategoryIsDefault = &Error{
		enMessage: "default category cannot be deleted or edited",
		zhMessage: "默认分类不可删除或编辑",
	}
	ErrPersonaNotFound = &Error{
		enMessage: "persona not found",
		zhMessage: "角色不存在",
	}
	ErrPersonaRoleInfoRequired = &Error{
		enMessage: "role info is required",
		zhMessage: "角色设定不能为空",
	}
	ErrPersonaRoleInfoTooLong = &Error{
		enMessage: "persona settings cannot exceed 500 characters",
		zhMessage: "角色设定不能超过500字",
	}
	ErrPersonaNameAlreadyExists = &Error{
		enMessage: "persona name already exists",
		zhMessage: "角色名称已存在",
	}
	ErrPersonaMCPNotFound = &Error{
		enMessage: "MCP endpoint not found or disabled",
		zhMessage: "MCP 端点不存在或已禁用",
	}
	ErrPersonaSkillNotInstalled = &Error{
		enMessage: "skill not installed",
		zhMessage: "技能未安装",
	}
	ErrPersonaModelDisabled = &Error{
		enMessage: "model not found or disabled",
		zhMessage: "模型不存在或已禁用",
	}
	ErrMCPToolNameConflict = &Error{
		enMessage: "MCP tool name conflict: different MCP endpoints expose tools with the same name",
		zhMessage: "MCP 工具名称冲突：不同 MCP 端点暴露了同名工具",
	}
	ErrSkillNameConflict = &Error{
		enMessage: "skill name conflict: selected skills have duplicate names",
		zhMessage: "技能名称冲突：所选技能存在同名",
	}
	ErrPasswordIncorrect = &Error{
		enMessage: "current password is incorrect",
		zhMessage: "当前密码错误",
	}
	ErrCaptchaRequired = &Error{
		enMessage: "captcha is required",
		zhMessage: "请输入验证码",
	}
	ErrCaptchaIncorrect = &Error{
		enMessage: "captcha is incorrect",
		zhMessage: "验证码错误",
	}
	ErrPasswordSameAsCurrent = &Error{
		enMessage: "new password must be different from current password",
		zhMessage: "新密码不能与当前密码相同",
	}
	ErrUserSkillNotFound = &Error{
		enMessage: "user skill not found",
		zhMessage: "用户技能不存在",
	}
	ErrUserSkillAlreadyInstalled = &Error{
		enMessage: "skill already installed",
		zhMessage: "技能已安装",
	}
	ErrUserSkillNotFailed = &Error{
		enMessage: "skill is not in failed status, cannot retry",
		zhMessage: "技能未处于失败状态，无法重试",
	}
	ErrMarketSkillNotAvailable = &Error{
		enMessage: "market skill is not available",
		zhMessage: "市场技能不可用",
	}
	ErrDefaultModelNotSet = &Error{
		enMessage: "no default model configured, please set a default model first",
		zhMessage: "未配置默认模型，请先在模型管理中设置默认模型",
	}
	ErrModelAndDefaultNotFound = &Error{
		enMessage: "specified model not found and no default model is configured",
		zhMessage: "指定模型不存在，且未配置默认模型",
	}
	ErrSessionNotFound = &Error{
		enMessage: "session not found",
		zhMessage: "会话不存在",
	}
	ErrChatContentRequired = &Error{
		enMessage: "content is required",
		zhMessage: "消息内容不能为空",
	}
	ErrChatModelRequired = &Error{
		enMessage: "aiModelId is required for new session",
		zhMessage: "新建会话需要指定模型",
	}
	ErrChatSessionActive = &Error{
		enMessage: "session already has an active runner",
		zhMessage: "会话正在处理中",
	}
	ErrChatRunnerNotFound = &Error{
		enMessage: "no active runner for session",
		zhMessage: "会话没有活跃的运行器",
	}
	ErrChatCompressTaskNotFound = &Error{
		enMessage: "compress task not found or expired",
		zhMessage: "压缩任务不存在或已过期",
	}
	ErrProjectMutualExclusive = &Error{
		enMessage: "project and sharedProjectId are mutually exclusive",
		zhMessage: "个人项目和团队共享项目不能同时设置",
	}
	ErrSharedProjectBusy = &Error{
		enMessage: "the shared project is currently in use by another session, please try again later",
		zhMessage: "该团队共享项目正在被其他会话使用，请稍后重试",
	}
	ErrShareNotFound = &Error{
		enMessage: "share link not found",
		zhMessage: "分享链接不存在",
	}
	ErrShareExpired = &Error{
		enMessage: "share link has expired",
		zhMessage: "分享链接已过期",
	}
	ErrShareAlreadyExists = &Error{
		enMessage: "share link already exists for this session",
		zhMessage: "该会话已有分享链接",
	}
	ErrShareSessionNotFound = &Error{
		enMessage: "session not found",
		zhMessage: "会话不存在",
	}
	ErrVisualModelNotSet = &Error{
		enMessage: "cannot enable Visual Understanding: no multimodal model is configured. Please set a multimodal model in model management first",
		zhMessage: "无法开启多模态识别：未设置多模态识别模型，请先在模型管理中设置多模态识别模型",
	}
	ErrSettingMustBeInteger = &Error{
		enMessage: "must be an integer",
		zhMessage: "必须为整数",
	}
	ErrSettingMustBeNumber = &Error{
		enMessage: "must be a number",
		zhMessage: "必须为数字",
	}
	ErrSettingMustBeValidNumber = &Error{
		enMessage: "must be a valid number",
		zhMessage: "必须为有效数字",
	}
	ErrSettingMustBeBool = &Error{
		enMessage: "must be true or false",
		zhMessage: "必须为 true 或 false",
	}
	ErrSettingInvalidDateFormat = &Error{
		enMessage: "must be in format 2026-05-11",
		zhMessage: "格式应为 2026-05-11",
	}
	ErrSettingInvalidDatetimeFormat = &Error{
		enMessage: "must be in format 2026-05-13 16:22:37",
		zhMessage: "格式应为 2026-05-13 16:22:37",
	}
	ErrSettingValueTooLong = &Error{
		enMessage: "value too long (max 65535)",
		zhMessage: "值过长（最大 65535）",
	}
	ErrTempAccessInvalid = &Error{
		enMessage: "invalid or expired access key",
		zhMessage: "访问凭证无效或已过期",
	}
	ErrTempAccessPathNotAllowed = &Error{
		enMessage: "path is not allowed by the access key",
		zhMessage: "该路径不在访问凭证允许范围内",
	}
	ErrTempAccessFileNotFound = &Error{
		enMessage: "file not found",
		zhMessage: "文件不存在",
	}
	ErrTempAccessNotFile = &Error{
		enMessage: "requested path is not a file",
		zhMessage: "请求的路径不是文件",
	}
	ErrTempAccessTypeInvalid = &Error{
		enMessage: "type must be 'file' or 'dir'",
		zhMessage: "类型必须为 'file' 或 'dir'",
	}
	ErrTempAccessPathInvalid = &Error{
		enMessage: "path is outside user workspace",
		zhMessage: "路径不在用户工作空间内",
	}
	ErrTempAccessNotDir = &Error{
		enMessage: "path is not a directory",
		zhMessage: "路径不是目录",
	}
	ErrSharedProjectNotFound = &Error{
		enMessage: "shared project not found",
		zhMessage: "共享项目不存在",
	}
	ErrSharedProjectPermission = &Error{
		enMessage: "permission denied: only the owner can manage this shared project",
		zhMessage: "权限不足：仅所有者可管理此共享项目",
	}
	ErrSharedProjectAlreadyShared = &Error{
		enMessage: "this project is already shared",
		zhMessage: "此项目已共享",
	}
	ErrSharedProjectInvalidName = &Error{
		enMessage: "invalid project name: must be a direct subdirectory of projects/",
		zhMessage: "项目名无效：必须是 projects/ 下的直接子目录",
	}
	ErrSharedProjectDirNotFound = &Error{
		enMessage: "project directory does not exist, contact the owner",
		zhMessage: "项目目录不存在，请联系所有者",
	}
	ErrDailyTokenLimitExceeded = &Error{
		enMessage: "daily token limit exceeded, please try again tomorrow or contact your administrator",
		zhMessage: "今日 Token 用量已达上限，请明日再试或联系管理员调整额度",
	}
)

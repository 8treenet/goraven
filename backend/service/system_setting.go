package service

import (
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util"
	"math"
	"strconv"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *SystemSettingService {
			return &SystemSettingService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *SystemSettingService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type SystemSettingService struct {
	Worker    freedom.Worker
	Repo      *repository.SystemSettingRepository
	ModelRepo *repository.ProviderRepository
}

type settingMeta struct {
	key           string
	valueType     string
	defaultValue  string
	displayNameZh string
	displayNameEn string
	descriptionZh string
	descriptionEn string
	inputType     string
	min           *float64
	max           *float64
	placeholder   string
	displayOrder  int
	groupName     string
}

type groupMeta struct {
	name          string
	displayNameZh string
	displayNameEn string
	displayOrder  int
}

var settingRegistry = []settingMeta{

	{key: "general.domain", valueType: po.ValueTypeString, defaultValue: "",
		displayNameZh: "系统域名", displayNameEn: "System Domain",
		descriptionZh: "系统对外服务域名，用于生成文件外链和分享链接", descriptionEn: "Public domain for generating file share links",
		inputType: "text", placeholder: "https://goraven.dev", displayOrder: 1, groupName: "general"},

	{key: "clawhub.api_url", valueType: po.ValueTypeString, defaultValue: "https://clawhub.ai",
		displayNameZh: "ClawHub API 地址", displayNameEn: "ClawHub API URL",
		descriptionZh: "ClawHub 服务接口地址", descriptionEn: "ClawHub service API URL",
		inputType: "text", placeholder: "https://clawhub.ai", displayOrder: 1, groupName: "clawhub"},
	{key: "clawhub.token", valueType: po.ValueTypeString, defaultValue: "",
		displayNameZh: "加速 Token", displayNameEn: "Acceleration Token",
		descriptionZh: "ClawHub 加速令牌，留空则不走加速并会被限速", descriptionEn: "ClawHub acceleration token; leave empty to skip acceleration (rate-limited)",
		inputType: "password", placeholder: "clh_xxxxxx", displayOrder: 2, groupName: "clawhub"},

	{key: "agent.max_iterations", valueType: po.ValueTypeInt, defaultValue: "120",
		displayNameZh: "最大迭代步数", displayNameEn: "Max Iterations",
		descriptionZh: "Agent 单次对话最大执行步骤", descriptionEn: "Max steps an agent can execute in one conversation",
		inputType: "number", min: util.PtrFloat64(20), max: util.PtrFloat64(200), displayOrder: 1, groupName: "agent"},

	{key: "agent.main_agent_timeout_minutes", valueType: po.ValueTypeInt, defaultValue: "20",
		displayNameZh: "任务最长执行时间", displayNameEn: "Task Timeout",
		descriptionZh: "Agent 单次任务允许的最长耗时（分钟），超时将自动终止", descriptionEn: "Maximum duration allowed for a single Agent task (minutes). Exceeding this will terminate the task.",
		inputType: "number", min: util.PtrFloat64(5), max: util.PtrFloat64(50), displayOrder: 2, groupName: "agent"},

	{key: "agent.compress_threshold_percent", valueType: po.ValueTypeInt, defaultValue: "80",
		displayNameZh: "压缩阈值百分比", displayNameEn: "Compress Threshold %",
		descriptionZh: "上下文占模型窗口百分比超过此值时触发压缩", descriptionEn: "Trigger compression when context exceeds this % of model window",
		inputType: "slider", min: util.PtrFloat64(40), max: util.PtrFloat64(80), displayOrder: 1, groupName: "agent"},
	{key: "agent.compress_keep_rounds", valueType: po.ValueTypeInt, defaultValue: "4",
		displayNameZh: "压缩保留轮数", displayNameEn: "Compress Keep Rounds",
		descriptionZh: "压缩时保留最近几轮对话不压缩", descriptionEn: "Number of recent rounds to keep uncompressed",
		inputType: "number", min: util.PtrFloat64(1), max: util.PtrFloat64(20), displayOrder: 2, groupName: "agent"},

	{key: "agent.pruning_token_threshold", valueType: po.ValueTypeInt, defaultValue: "96",
		displayNameZh: "剪枝 Token 阈值/K", displayNameEn: "Pruning Token Threshold/K",
		descriptionZh: "总 token 超过此阈值（单位 K）时触发剪枝", descriptionEn: "Trigger pruning when total tokens exceed this threshold (K)",
		inputType: "number", min: util.PtrFloat64(64), max: util.PtrFloat64(1000), displayOrder: 4, groupName: "agent"},
	{key: "agent.pruning_max_tool_result_length", valueType: po.ValueTypeInt, defaultValue: "2000",
		displayNameZh: "剪枝-工具结果最大长度", displayNameEn: "Pruning-Max Tool Result Length",
		descriptionZh: "剪枝时工具返回结果超过此长度时截断", descriptionEn: "Pruning: truncate tool results exceeding this length",
		inputType: "number", min: util.PtrFloat64(1000), max: util.PtrFloat64(20000), displayOrder: 5, groupName: "agent"},
	{key: "agent.pruning_head_truncate_length", valueType: po.ValueTypeInt, defaultValue: "1000",
		displayNameZh: "剪枝-截断保留头部", displayNameEn: "Pruning-Head Truncate Length",
		descriptionZh: "剪枝截断时保留头部的字符数", descriptionEn: "Pruning: characters to keep from the head when truncating",
		inputType: "number", min: util.PtrFloat64(0), max: util.PtrFloat64(10000), displayOrder: 6, groupName: "agent"},
	{key: "agent.pruning_tail_truncate_length", valueType: po.ValueTypeInt, defaultValue: "1000",
		displayNameZh: "剪枝-截断保留尾部", displayNameEn: "Pruning-Tail Truncate Length",
		descriptionZh: "剪枝截断时保留尾部的字符数", descriptionEn: "Pruning: characters to keep from the tail when truncating",
		inputType: "number", min: util.PtrFloat64(0), max: util.PtrFloat64(10000), displayOrder: 7, groupName: "agent"},

	{key: "agent.llm_request_delay_ms", valueType: po.ValueTypeInt, defaultValue: "500",
		displayNameZh: "LLM 请求延迟/ms", displayNameEn: "LLM Request Delay/ms",
		descriptionZh: "两次 LLM 请求之间的延迟，用于避免限流", descriptionEn: "Delay between LLM requests to avoid rate limiting",
		inputType: "number", min: util.PtrFloat64(0), max: util.PtrFloat64(10000), displayOrder: 8, groupName: "agent"},
	{key: "agent.max_retries", valueType: po.ValueTypeInt, defaultValue: "3",
		displayNameZh: "LLM 最大重试次数", displayNameEn: "LLM Max Retries",
		descriptionZh: "LLM 调用失败时的最大重试次数", descriptionEn: "Max retry count for failed LLM calls",
		inputType: "number", min: util.PtrFloat64(0), max: util.PtrFloat64(10), displayOrder: 9, groupName: "agent"},
	{key: "agent.rate_limit_wait_sec", valueType: po.ValueTypeInt, defaultValue: "8",
		displayNameZh: "429 限流等待/秒", displayNameEn: "429 Rate Limit Wait/sec",
		descriptionZh: "遇到 429 限流时的固定等待秒数", descriptionEn: "Fixed wait seconds when encountering 429 rate limiting",
		inputType: "number", min: util.PtrFloat64(1), max: util.PtrFloat64(60), displayOrder: 10, groupName: "agent"},
	{key: "agent.backoff_base_sec", valueType: po.ValueTypeInt, defaultValue: "3",
		displayNameZh: "退避重试基秒", displayNameEn: "Backoff Base Sec",
		descriptionZh: "退避重试的基秒数, 第 N 次重试等待 N×此值秒", descriptionEn: "Base seconds for backoff retry, wait N×base seconds on Nth retry",
		inputType: "number", min: util.PtrFloat64(1), max: util.PtrFloat64(30), displayOrder: 11, groupName: "agent"},

	{key: "sharing.file_expires_hours", valueType: po.ValueTypeInt, defaultValue: "72",
		displayNameZh: "文件外链有效期/小时", displayNameEn: "File Link Expiry/hours",
		descriptionZh: "文件分享外链的有效期", descriptionEn: "Expiration time for file share links",
		inputType: "number", min: util.PtrFloat64(1), max: util.PtrFloat64(720), displayOrder: 1, groupName: "sharing"},

	{key: "knowledge.enable_ocr", valueType: po.ValueTypeBool, defaultValue: "false",
		displayNameZh: "OCR 开关", displayNameEn: "Enable OCR",
		descriptionZh: "是否启用 OCR 解析", descriptionEn: "Enable OCR parsing for documents",
		inputType: "switch", displayOrder: 1, groupName: "knowledge"},

	{key: "tools.webfetch_enabled", valueType: po.ValueTypeBool, defaultValue: "true",
		displayNameZh: "网页读取", displayNameEn: "Web Fetch",
		descriptionZh: "启用后 Agent 可通过 HTTP GET 读取指定网页内容。注意：仅能获取服务端返回的原始 HTML，无法读取前端渲染的动态页面。",
		descriptionEn: "Allow agents to fetch web page content via HTTP GET. Note: only raw server-returned HTML is available; client-side rendered pages cannot be read.",
		inputType:     "switch", displayOrder: 1, groupName: "tools"},
	{key: "tools.visual_enabled", valueType: po.ValueTypeBool, defaultValue: "false",
		displayNameZh: "多模态识别", displayNameEn: "Visual Understanding",
		descriptionZh: "启用图像、视频、音频识别能力。需在模型管理中设置多模态模型，且该模型需支持多模态。",
		descriptionEn: "Enable image, video, and audio recognition. Requires setting a multimodal model (must support multimodal input) in model management.",
		inputType:     "switch", displayOrder: 2, groupName: "tools"},
	{key: "tools.shell_timeout_minutes", valueType: po.ValueTypeInt, defaultValue: "5",
		displayNameZh: "命令执行超时", displayNameEn: "Command Timeout",
		descriptionZh: "执行终端命令允许的最长耗时（分钟），超时将自动终止", descriptionEn: "Maximum duration allowed for a single terminal command (minutes). Exceeding this will terminate the command.",
		inputType: "number", min: util.PtrFloat64(1), max: util.PtrFloat64(30), displayOrder: 3, groupName: "tools"},
}

var groupRegistry = []groupMeta{
	{name: "general", displayNameZh: "基本配置", displayNameEn: "General", displayOrder: 1},
	{name: "tools", displayNameZh: "工具", displayNameEn: "Tools", displayOrder: 6},
	{name: "clawhub", displayNameZh: "ClawHub", displayNameEn: "ClawHub", displayOrder: 2},
	{name: "agent", displayNameZh: "Agent 配置", displayNameEn: "Agent Config", displayOrder: 3},
	{name: "sharing", displayNameZh: "分享配置", displayNameEn: "Sharing", displayOrder: 4},
	{name: "knowledge", displayNameZh: "知识库", displayNameEn: "Knowledge", displayOrder: 5},
}

var metaMap map[string]*settingMeta

func init() {
	metaMap = make(map[string]*settingMeta, len(settingRegistry))
	for i := range settingRegistry {
		metaMap[settingRegistry[i].key] = &settingRegistry[i]
	}
}

func (service *SystemSettingService) GetSettings() ([]vo.AdminSettingGroup, error) {
	rows, err := service.Repo.GetAll()
	if err != nil {
		return nil, err
	}

	valueMap := make(map[string]string, len(rows))
	for _, r := range rows {
		valueMap[r.Key] = r.Value
	}

	lang := config.Get().GetLanguage()

	groupSettings := make(map[string][]vo.AdminSettingItem, len(groupRegistry))
	for _, m := range settingRegistry {
		item := vo.AdminSettingItem{
			Key:          m.key,
			Value:        valueMap[m.key],
			ValueType:    m.valueType,
			DefaultValue: m.defaultValue,
			InputType:    m.inputType,
			Min:          m.min,
			Max:          m.max,
			DisplayOrder: m.displayOrder,
		}

		if m.valueType == po.ValueTypeString {
			item.Placeholder = m.placeholder
		}

		if item.Value == "" {
			item.Value = m.defaultValue
		}

		if lang == "en" {
			item.DisplayName = m.displayNameEn
			item.Description = m.descriptionEn
		} else {
			item.DisplayName = m.displayNameZh
			item.Description = m.descriptionZh
		}

		groupSettings[m.groupName] = append(groupSettings[m.groupName], item)
	}

	groups := make([]vo.AdminSettingGroup, 0, len(groupRegistry))
	for _, g := range groupRegistry {
		items := groupSettings[g.name]
		if len(items) == 0 {
			continue
		}
		group := vo.AdminSettingGroup{
			Name:         g.name,
			DisplayOrder: g.displayOrder,
			Settings:     items,
		}
		if lang == "en" {
			group.DisplayName = g.displayNameEn
		} else {
			group.DisplayName = g.displayNameZh
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func (service *SystemSettingService) UpdateSettings(req *vo.AdminUpdateSettingsReq) (*vo.AdminUpdateSettingsRsp, error) {
	updates := make(map[string]string, len(req.Settings))

	for _, item := range req.Settings {
		meta, ok := metaMap[item.Key]
		if !ok {
			return nil, errs.NewFormatError("unknown setting key: %s", "未知的设置项: %s", item.Key)
		}

		if err := service.validateSettingValue(meta, item.Value); err != nil {
			return nil, errs.NewFormatError("invalid value for %s: %s", "设置项 %s 的值无效: %s", item.Key, err.Error())
		}

		updates[item.Key] = item.Value
	}

	if v, ok := updates["tools.visual_enabled"]; ok {
		enabled, _ := strconv.ParseBool(v)
		if enabled {
			hasVisual, err := service.ModelRepo.HasVisualModel()
			if err != nil {
				return nil, fmt.Errorf("failed to check visual model: %w", err)
			}
			if !hasVisual {
				return nil, errs.ErrVisualModelNotSet
			}
		}
	}

	pruningKeys := [3]string{
		"agent.pruning_max_tool_result_length",
		"agent.pruning_head_truncate_length",
		"agent.pruning_tail_truncate_length",
	}
	pruningAffected := false
	for _, k := range pruningKeys {
		if _, ok := updates[k]; ok {
			pruningAffected = true
			break
		}
	}
	if pruningAffected {
		pruningValues := make(map[string]int, 3)
		for _, k := range pruningKeys {
			if v, ok := updates[k]; ok {
				pruningValues[k], _ = strconv.Atoi(v)
			} else {
				pruningValues[k] = service.Repo.GetInt(k, 2000)
			}
		}
		maxLen := pruningValues["agent.pruning_max_tool_result_length"]
		headLen := pruningValues["agent.pruning_head_truncate_length"]
		tailLen := pruningValues["agent.pruning_tail_truncate_length"]
		if headLen+tailLen > maxLen {
			return nil, errs.NewFormatError(
				"head truncate length (%d) + tail truncate length (%d) must not exceed max tool result length (%d)",
				"截断保留头部 (%d) + 截断保留尾部 (%d) 不能超过工具结果最大长度 (%d)",
				headLen, tailLen, maxLen,
			)
		}
	}

	if err := service.Repo.BatchUpdate(updates); err != nil {
		return nil, err
	}

	return &vo.AdminUpdateSettingsRsp{Updated: len(updates)}, nil
}

func (service *SystemSettingService) validateSettingValue(meta *settingMeta, value string) error {
	switch meta.valueType {
	case po.ValueTypeInt:
		v, err := strconv.Atoi(value)
		if err != nil {
			return errs.ErrSettingMustBeInteger
		}
		if meta.min != nil && float64(v) < *meta.min {
			return errs.NewFormatError("must be >= %v", "必须 >= %v", *meta.min)
		}
		if meta.max != nil && float64(v) > *meta.max {
			return errs.NewFormatError("must be <= %v", "必须 <= %v", *meta.max)
		}

	case po.ValueTypeFloat:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errs.ErrSettingMustBeNumber
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errs.ErrSettingMustBeValidNumber
		}
		if meta.min != nil && v < *meta.min {
			return errs.NewFormatError("must be >= %v", "必须 >= %v", *meta.min)
		}
		if meta.max != nil && v > *meta.max {
			return errs.NewFormatError("must be <= %v", "必须 <= %v", *meta.max)
		}

	case po.ValueTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return errs.ErrSettingMustBeBool
		}

	case po.ValueTypeDate:
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return errs.ErrSettingInvalidDateFormat
		}

	case po.ValueTypeDatetime:
		if _, err := time.Parse("2006-01-02 15:04:05", value); err != nil {
			return errs.ErrSettingInvalidDatetimeFormat
		}

	case po.ValueTypeString:
		if len(value) > 65535 {
			return errs.ErrSettingValueTooLong
		}
	}

	return nil
}

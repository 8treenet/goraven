package seed

import "goraven/backend/po"

var SystemSettings = []po.SystemSetting{

	{Key: "general.domain", Value: "", ValueType: po.ValueTypeString, GroupName: "general"},

	{Key: "tools.webfetch_enabled", Value: "true", ValueType: po.ValueTypeBool, GroupName: "tools"},
	{Key: "tools.visual_enabled", Value: "false", ValueType: po.ValueTypeBool, GroupName: "tools"},
	{Key: "tools.shell_timeout_minutes", Value: "8", ValueType: po.ValueTypeInt, GroupName: "tools"},

	{Key: "clawhub.api_url", Value: "https://clawhub.ai", ValueType: po.ValueTypeString, GroupName: "clawhub"},
	{Key: "clawhub.token", Value: "", ValueType: po.ValueTypeString, GroupName: "clawhub"},

	{Key: "agent.compress_threshold_percent", Value: "80", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.compress_keep_rounds", Value: "4", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.max_iterations", Value: "120", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "agent.pruning_token_threshold", Value: "96", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "agent.pruning_max_tool_result_length", Value: "2000", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.pruning_head_truncate_length", Value: "1000", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.pruning_tail_truncate_length", Value: "1000", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "agent.llm_request_delay_ms", Value: "500", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "agent.max_retries", Value: "3", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.rate_limit_wait_sec", Value: "8", ValueType: po.ValueTypeInt, GroupName: "agent"},
	{Key: "agent.backoff_base_sec", Value: "3", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "agent.main_agent_timeout_minutes", Value: "30", ValueType: po.ValueTypeInt, GroupName: "agent"},

	{Key: "sharing.file_expires_hours", Value: "72", ValueType: po.ValueTypeInt, GroupName: "sharing"},
	{Key: "sharing.link_expires_hours", Value: "168", ValueType: po.ValueTypeInt, GroupName: "sharing"},

	{Key: "knowledge.enable_ocr", Value: "false", ValueType: po.ValueTypeBool, GroupName: "knowledge"},
}

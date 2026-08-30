package vo

// ClawHubSearchReq ClawHub 搜索请求
type ClawHubSearchReq struct {
	Query string `url:"q"`     // 搜索关键词
	Limit int    `url:"limit"` // 返回条数（默认 25，最大 200）
}

// ClawHubExploreReq ClawHub 浏览技能列表请求
type ClawHubExploreReq struct {
	Sort string `url:"sort"` // 排序：newest/updated/downloads/stars/installs/trending
}

// AdminClawHubSkillDetailRsp ClawHub 技能详情响应（预览用）
type AdminClawHubSkillDetailRsp struct {
	Slug        string `json:"slug"`        // 唯一标识
	Name        string `json:"name"`        // 技能名（从 SKILL.md 解析）
	Description string `json:"description"` // 描述（从 SKILL.md 解析）
	Content     string `json:"content"`     // SKILL.md 完整内容
	Version     string `json:"version"`     // 版本号
}

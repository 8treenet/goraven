package vo

type ClawHubSearchReq struct {
	Query string `url:"q"`
	Limit int    `url:"limit"`
}

type ClawHubExploreReq struct {
	Sort string `url:"sort"`
}

type AdminClawHubSkillDetailRsp struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Version     string `json:"version"`
}

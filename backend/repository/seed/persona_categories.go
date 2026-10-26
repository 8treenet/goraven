package seed

// PersonaCategorySeed holds static seed data for persona categories.
type PersonaCategorySeed struct {
	EnName    string
	ZhName    string
	Icon      string
	IsDefault uint8
}

// PersonaCategories contains the default persona category seed data.
var PersonaCategories = []PersonaCategorySeed{
	{EnName: "General", ZhName: "通用", Icon: "bot", IsDefault: 1},
	{EnName: "Programming", ZhName: "编程开发", Icon: "code"},
	{EnName: "Translation", ZhName: "翻译语言", Icon: "languages"},
	{EnName: "Creative", ZhName: "内容创作", Icon: "pen-line"},
	{EnName: "Data Analysis", ZhName: "数据分析", Icon: "bar-chart-2"},
	{EnName: "Education", ZhName: "学习教育", Icon: "graduation-cap"},
	{EnName: "Business Efficiency", ZhName: "商业效率", Icon: "briefcase"},
}

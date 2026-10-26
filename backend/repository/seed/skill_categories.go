package seed

// SkillCategorySeed holds static seed data for skill categories.
type SkillCategorySeed struct {
	EnName    string
	ZhName    string
	IsDefault uint8
}

// SkillCategories contains the default skill category seed data.
var SkillCategories = []SkillCategorySeed{
	{EnName: "General", ZhName: "通用", IsDefault: 1},
	{EnName: "Programming", ZhName: "编程开发"},
	{EnName: "Data & AI", ZhName: "数据与AI"},
	{EnName: "DevOps", ZhName: "运维部署"},
	{EnName: "Design", ZhName: "设计创意"},
	{EnName: "Business", ZhName: "商业效率"},
	{EnName: "Creative", ZhName: "内容创作"},
	{EnName: "Education", ZhName: "学习教育"},
}

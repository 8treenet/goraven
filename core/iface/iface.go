package iface

type SkillAccessor interface {
	IsSkillSelected(name string) bool
	AddToolDailyStats(name string)
}

type MCPFilter interface {
	IsMCPSelected(name string) bool
}

type FileURLGenerator interface {
	GenerateURL(userID string, filePath string) (string, error)
}

type SkillInfo struct {
	Name        string
	Description string
	Content     string
}

type SystemSkillProvider interface {
	SystemSkillList() ([]SkillInfo, error)
}

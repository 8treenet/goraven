package iface

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

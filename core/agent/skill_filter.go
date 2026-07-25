package agent

type SimpleSkillFilter struct {
}

func NewSimpleSkillFilter(skillNames []string, onUse func(name string)) *SimpleSkillFilter {
	return &SimpleSkillFilter{}
}

func (f *SimpleSkillFilter) IsSkillSelected(name string) bool {
	return false
}

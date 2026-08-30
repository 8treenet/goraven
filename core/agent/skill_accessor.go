package agent

type SimpleSkillAccessor struct {
	selectedMap           map[string]bool
	addToolDailyStatsCall func(name string)
	invokedSkills         map[string]bool
}

func NewSimpleSkillFilter(skillNames []string, addToolDailyStatsCall func(name string)) *SimpleSkillAccessor {
	m := make(map[string]bool, len(skillNames))
	for _, name := range skillNames {
		m[name] = true
	}
	return &SimpleSkillAccessor{
		selectedMap:           m,
		addToolDailyStatsCall: addToolDailyStatsCall,
		invokedSkills:         make(map[string]bool),
	}
}

func (f *SimpleSkillAccessor) IsSkillSelected(name string) bool {
	return f.selectedMap[name]
}

func (f *SimpleSkillAccessor) AddToolDailyStats(name string) {
	f.invokedSkills[name] = true
	f.addToolDailyStatsCall(name)
}

func (f *SimpleSkillAccessor) IsSkillInvoked(name string) bool {
	return f.invokedSkills[name]
}

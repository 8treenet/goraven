package agent

type SimpleMCPFilter struct {
	selectedMap map[string]bool
}

func NewSimpleMCPFilter(mcpNames []string) *SimpleMCPFilter {
	m := make(map[string]bool, len(mcpNames))
	for _, name := range mcpNames {
		m[name] = true
	}
	return &SimpleMCPFilter{
		selectedMap: m,
	}
}

func (f *SimpleMCPFilter) IsMCPSelected(name string) bool {
	return f.selectedMap[name]
}

package agent

type SimpleMCPFilter struct {
}

func NewSimpleMCPFilter(mcpNames []string) *SimpleMCPFilter {
	return &SimpleMCPFilter{}
}

func (f *SimpleMCPFilter) IsMCPSelected(name string) bool {
	return false
}

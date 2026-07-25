package repository

type SystemConfig struct {
	GeneralDomain string

	ClawHubAPIURL string
	ClawHubToken  string

	CompressThresholdPercent   int
	CompressKeepRounds         int
	MaxIterations              int
	PruningTokenThreshold      int
	PruningMaxToolResultLength int
	PruningHeadTruncateLength  int
	PruningTailTruncateLength  int

	LLMRequestDelayMs    int
	FileLinkExpiresHours int

	KnowledgeEnableOCR bool

	WebFetchEnabled bool
	VisualEnabled   bool

	ShellTimeoutMinutes int

	ModelMaxRetries       int
	ModelRateLimitWaitSec int
	ModelBackoffBaseSec   int

	MainAgentTimeoutMinutes int
}

func NewDefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		CompressThresholdPercent:   80,
		CompressKeepRounds:         3,
		MaxIterations:              120,
		PruningTokenThreshold:      96,
		LLMRequestDelayMs:          500,
		PruningMaxToolResultLength: 2000,
		PruningHeadTruncateLength:  1000,
		PruningTailTruncateLength:  1000,
		FileLinkExpiresHours:       72,
		KnowledgeEnableOCR:         false,
		WebFetchEnabled:            true,
		VisualEnabled:              false,
		ShellTimeoutMinutes:        5,
		ModelMaxRetries:            3,
		ModelRateLimitWaitSec:      8,
		ModelBackoffBaseSec:        3,
		MainAgentTimeoutMinutes:    20,
	}
}

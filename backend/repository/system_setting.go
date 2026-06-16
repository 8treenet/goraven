package repository

import (
	"raven/backend/po"
	"strconv"
	"sync"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *SystemSettingRepository {
			return &SystemSettingRepository{}
		})
	})
}

var (
	configCache	*SystemConfig
	configCacheTime	time.Time
	configCacheMu	sync.RWMutex
	configCacheTTL	= 20 * time.Minute
)

type SystemSettingRepository struct {
	freedom.Repository
}

func (repo *SystemSettingRepository) LoadConfig() (*SystemConfig, error) {
	configCacheMu.RLock()
	if configCache != nil && time.Since(configCacheTime) < configCacheTTL {
		result := *configCache
		configCacheMu.RUnlock()
		return &result, nil
	}
	configCacheMu.RUnlock()

	rows, err := repo.GetAll()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.Key] = r.Value
	}

	cfg := &SystemConfig{
		GeneralDomain:			getString(m, "general.domain", ""),
		ClawHubAPIURL:			getString(m, "clawhub.api_url", "https://clawhub.ai"),
		ClawHubToken:			getString(m, "clawhub.token", ""),
		CompressThresholdPercent:	getInt(m, "agent.compress_threshold_percent", 80),
		CompressKeepRounds:		getInt(m, "agent.compress_keep_rounds", 4),
		MaxIterations:			getInt(m, "agent.max_iterations", 150),
		PruningTokenThreshold:		getInt(m, "agent.pruning_token_threshold", 96),
		PruningMaxToolResultLength:	getInt(m, "agent.pruning_max_tool_result_length", 2000),
		PruningHeadTruncateLength:	getInt(m, "agent.pruning_head_truncate_length", 1000),
		PruningTailTruncateLength:	getInt(m, "agent.pruning_tail_truncate_length", 1000),
		LLMRequestDelayMs:		getInt(m, "agent.llm_request_delay_ms", 500),
		FileLinkExpiresHours:		getInt(m, "sharing.file_expires_hours", 72),
		KnowledgeEnableOCR:		getBool(m, "knowledge.enable_ocr", false),
		WebFetchEnabled:		getBool(m, "tools.webfetch_enabled", true),
		VisualEnabled:			getBool(m, "tools.visual_enabled", false),
		ShellTimeoutMinutes:		getInt(m, "tools.shell_timeout_minutes", 5),
		ModelMaxRetries:		getInt(m, "agent.max_retries", 3),
		ModelRateLimitWaitSec:		getInt(m, "agent.rate_limit_wait_sec", 8),
		ModelBackoffBaseSec:		getInt(m, "agent.backoff_base_sec", 3),
		MainAgentTimeoutMinutes:	getInt(m, "agent.main_agent_timeout_minutes", 20),
	}

	configCache = cfg
	configCacheTime = time.Now()

	result := *cfg
	return &result, nil
}

func (repo *SystemSettingRepository) invalidateCache() {
	configCacheMu.Lock()
	configCache = nil
	configCacheMu.Unlock()
}

func getString(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return def
}

func getInt(m map[string]string, key string, def int) int {
	v, ok := m[key]
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getBool(m map[string]string, key string, def bool) bool {
	v, ok := m[key]
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func (repo *SystemSettingRepository) GetAll() ([]*po.SystemSetting, error) {
	var settings []*po.SystemSetting
	err := repo.db().Order("group_name asc, config_key asc").Find(&settings).Error
	return settings, err
}

func (repo *SystemSettingRepository) GetByKey(key string) (*po.SystemSetting, error) {
	var setting po.SystemSetting
	err := repo.db().First(&setting, "config_key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (repo *SystemSettingRepository) GetByGroup(group string) ([]*po.SystemSetting, error) {
	var settings []*po.SystemSetting
	err := repo.db().Where("group_name = ?", group).Order("config_key asc").Find(&settings).Error
	return settings, err
}

func (repo *SystemSettingRepository) Update(key string, value string) error {
	err := repo.db().Model(&po.SystemSetting{}).Where("config_key = ?", key).Update("config_value", value).Error
	if err == nil {
		repo.invalidateCache()
	}
	return err
}

func (repo *SystemSettingRepository) BatchUpdate(updates map[string]string) error {
	err := repo.db().Transaction(func(tx *gorm.DB) error {
		for key, value := range updates {
			if err := tx.Model(&po.SystemSetting{}).Where("config_key = ?", key).Update("config_value", value).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		repo.invalidateCache()
	}
	return err
}

func (repo *SystemSettingRepository) GetString(key, defaultValue string) string {
	setting, err := repo.GetByKey(key)
	if err != nil || setting.Value == "" {
		return defaultValue
	}
	return setting.Value
}

func (repo *SystemSettingRepository) GetInt(key string, defaultValue int) int {
	setting, err := repo.GetByKey(key)
	if err != nil {
		return defaultValue
	}
	v, err := strconv.Atoi(setting.Value)
	if err != nil {
		return defaultValue
	}
	return v
}

func (repo *SystemSettingRepository) GetBool(key string, defaultValue bool) bool {
	setting, err := repo.GetByKey(key)
	if err != nil {
		return defaultValue
	}
	v, err := strconv.ParseBool(setting.Value)
	if err != nil {
		return defaultValue
	}
	return v
}

func (repo *SystemSettingRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

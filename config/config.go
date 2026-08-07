package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/8treenet/freedom"
	"gopkg.in/yaml.v3"
)

var Version = "0.4.3"

var StartTime = time.Now()

func Get() *Configuration {
	once.Do(func() {
		cfg = newConfig()
	})
	return cfg
}

var once sync.Once
var cfg *Configuration

var userSpaceCache = struct {
	sync.RWMutex
	entries map[string]time.Time
}{
	entries: make(map[string]time.Time),
}

const userSpaceCacheTTL = 72 * time.Hour

type Configuration struct {
	App        freedom.Configuration
	Other      map[string]interface{} `toml:"other" yaml:"other"`
	Paths      PathsConf              `toml:"paths" yaml:"paths"`
	Tools      ToolsConf              `toml:"tools" yaml:"tools"`
	System     SystemConf             `toml:"system" yaml:"system"`
	Behavior   BehaviorConf           `toml:"behavior" yaml:"behavior"`
	Redis      RedisConf              `toml:"redis" yaml:"redis"`
	DB         DBConf                 `toml:"db" yaml:"db"`
	configPath string
}

type DBConf struct {
	Addr            string `toml:"addr" yaml:"addr"`
	MaxOpenConns    int    `toml:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifeTime int    `toml:"conn_max_life_time" yaml:"conn_max_life_time"`
	ConnMaxIdleTime int    `toml:"conn_max_idle_time" yaml:"conn_max_idle_time"`
}

type PathsConf struct {
	SqliteDB          string `toml:"sqlite_db" yaml:"sqlite_db"`
	ChromaDIR         string `toml:"chroma_dir" yaml:"chroma_dir"`
	UserSpace         string `toml:"user_space" yaml:"user_space"`
	FrontendDir       string `toml:"frontend_dir" yaml:"frontend_dir"`
	Scripts           string `toml:"scripts" yaml:"scripts"`
	SkillsHub         string `toml:"skills_hub" yaml:"skills_hub"`
	SkillShareDir     string `toml:"skill_share_dir" yaml:"skill_share_dir"`
	UploadDir         string `toml:"upload_dir" yaml:"upload_dir"`
	SkillInstalledDir string `toml:"skill_installed_dir" yaml:"skill_installed_dir"`
}

type ToolsConf struct {
	HTTPTimeoutSeconds int `toml:"http_timeout_seconds" yaml:"http_timeout_seconds"`
}

type RedisConf struct {
	Addr         string `toml:"addr" yaml:"addr"`
	Password     string `toml:"password" yaml:"password"`
	DB           int    `toml:"db" yaml:"db"`
	MaxRetries   int    `toml:"max_retries" yaml:"max_retries"`
	PoolSize     int    `toml:"pool_size" yaml:"pool_size"`
	ReadTimeout  int    `toml:"read_timeout" yaml:"read_timeout"`
	WriteTimeout int    `toml:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  int    `toml:"idle_timeout" yaml:"idle_timeout"`
	MaxConnAge   int    `toml:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout  int    `toml:"pool_timeout" yaml:"pool_timeout"`
}

type SystemConf struct {
	Language    string `toml:"language" yaml:"language"`
	Initialized bool   `toml:"initialized" yaml:"initialized"`
	DBType      string `toml:"db_type" yaml:"db_type"`
	CacheType   string `toml:"cache_type" yaml:"cache_type"`
}

type BehaviorConf struct {
	PreviewUser string `toml:"preview_user" yaml:"preview_user"`
	Docker      bool   `toml:"docker" yaml:"docker"`
}

func newConfig() *Configuration {
	result := &Configuration{}
	def := freedom.DefaultConfiguration()
	def.Other["listen_addr"] = ":8000"
	def.Other["service_name"] = "default"
	result.App = def

	file := ParseConfigPath()
	if file == "" {
		file = "config.yaml"
	}
	result.configPath = file

	err := freedom.Configure(&result, file)
	if err == nil {
		result.App.Other = result.Other
	}

	if err != nil {
		freedom.Logger().Fatal(err)
	}

	if result.Tools.HTTPTimeoutSeconds <= 0 {
		result.Tools.HTTPTimeoutSeconds = 60
	}

	if result.DB.Addr == "" {
		result.DB.Addr = result.Paths.SqliteDB
	}
	if result.DB.MaxOpenConns <= 0 {
		result.DB.MaxOpenConns = 4
	}
	if result.DB.MaxIdleConns <= 0 {
		result.DB.MaxIdleConns = 2
	}
	if result.DB.ConnMaxLifeTime <= 0 {
		result.DB.ConnMaxLifeTime = 200
	}
	if result.DB.ConnMaxIdleTime <= 0 {
		result.DB.ConnMaxIdleTime = 170
	}

	os.MkdirAll(result.Paths.ChromaDIR, 0755)
	os.MkdirAll(result.Paths.SkillsHub, 0755)
	os.MkdirAll(result.GetSkillShareDir(), 0755)
	os.MkdirAll(result.GetUploadDir(), 0755)
	os.MkdirAll(result.GetUploadTempDir(), 0755)
	os.MkdirAll(result.GetDownloadTempDir(), 0755)
	os.MkdirAll(result.GetClawHUBCacheDir(), 0755)
	os.MkdirAll(result.GetSkillInstalledDir(), 0755)
	return result
}

func ParseConfigPath() string {
	confdir := os.Getenv("GORAVEN_CONF")
	if confdir != "" {
		return confdir
	}

	if _, err := os.Stat("/goraven/data/config.yaml"); err == nil {
		return "/goraven/data/config.yaml"
	}
	return "config/config.yaml"
}

func (conf *Configuration) GetUserSpace(userName string) string {
	userSpace := fmt.Sprintf("%s/%s", conf.Paths.UserSpace, userName)

	userSpaceCache.RLock()
	lastCreated, exists := userSpaceCache.entries[userName]
	userSpaceCache.RUnlock()

	if exists && time.Since(lastCreated) < userSpaceCacheTTL {
		return userSpace
	}

	userSpaceCache.Lock()
	userSpaceCache.entries[userName] = time.Now()
	userSpaceCache.Unlock()

	os.MkdirAll(userSpace, 0755)
	os.MkdirAll(userSpace+"/documents", 0755)
	os.MkdirAll(userSpace+"/temp", 0755)
	os.MkdirAll(userSpace+"/downloads", 0755)
	os.MkdirAll(userSpace+"/images", 0755)
	os.MkdirAll(userSpace+"/videos", 0755)
	os.MkdirAll(userSpace+"/projects", 0755)
	os.MkdirAll(userSpace+"/skills", 0755)
	profilePath := userSpace + "/.profile"
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		_ = os.WriteFile(profilePath, nil, 0644)
	}

	return userSpace
}

func (conf *Configuration) GetChromaPath(datasetID int) string {
	path := fmt.Sprintf("%s/%d.db", conf.Paths.ChromaDIR, datasetID)
	return path
}

func (conf *Configuration) ModifyConfig(section, key, value string) error {
	data, err := os.ReadFile(conf.configPath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("decode config file: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("invalid yaml document")
	}

	rootMapping := root.Content[0]
	if rootMapping.Kind != yaml.MappingNode {
		return fmt.Errorf("expected root mapping")
	}

	sectionNode := findMappingValue(rootMapping, section)
	if sectionNode == nil {
		sectionNode = &yaml.Node{Kind: yaml.MappingNode}
		rootMapping.Content = append(rootMapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: section},
			sectionNode,
		)
	} else if sectionNode.Kind != yaml.MappingNode {
		return fmt.Errorf("section %q is not a mapping", section)
	}

	keyIdx := findMappingKey(sectionNode, key)
	if keyIdx >= 0 {
		valNode := sectionNode.Content[keyIdx+1]
		valNode.Value = value
	} else {
		sectionNode.Content = append(sectionNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: "!!str"},
		)
	}

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("encode config file: %w", err)
	}

	if err := os.WriteFile(conf.configPath, out, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func findMappingKey(mapping *yaml.Node, key string) int {
	if mapping.Kind != yaml.MappingNode {
		return -1
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return i
		}
	}
	return -1
}

func findMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	idx := findMappingKey(mapping, key)
	if idx < 0 {
		return nil
	}
	return mapping.Content[idx+1]
}

type BuildInfo struct {
	Version   string
	StartTime time.Time
}

func (conf *Configuration) GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		Version:   Version,
		StartTime: StartTime,
	}
}

func (conf *Configuration) GetLanguage() string {
	if conf.System.Language != "" {
		return conf.System.Language
	}
	return "zh"
}

func (conf *Configuration) GetUploadDir() string {
	if conf.Paths.UploadDir != "" {
		return conf.Paths.UploadDir
	}
	return "/goraven/data/upload"
}

func (conf *Configuration) GetUploadTempDir() string {
	return filepath.Join(os.TempDir(), "goraven-upload")
}

func (conf *Configuration) GetDownloadTempDir() string {
	return filepath.Join(os.TempDir(), "goraven-download")
}

func (conf *Configuration) GetClawHUBCacheDir() string {
	return filepath.Join(os.TempDir(), "goraven-clawhub-cache")
}

func (conf *Configuration) GetSkillShareDir() string {
	if conf.Paths.SkillShareDir != "" {
		return conf.Paths.SkillShareDir
	}
	return "/goraven/data/skill_share"
}

func (conf *Configuration) GetSkillInstalledDir() string {
	if conf.Paths.SkillInstalledDir != "" {
		return conf.Paths.SkillInstalledDir
	}
	return "/goraven/skill_installed"
}

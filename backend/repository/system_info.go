package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/config"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *SystemInfoRepository {
			return &SystemInfoRepository{}
		})
	})
}

const systemInfoCacheKey = "admin:systemInfo"
const systemInfoCacheTTL = 5 * time.Minute

// SystemInfoRepository 系统信息仓库
// 提供数据库统计查询和 Redis 缓存
type SystemInfoRepository struct {
	freedom.Repository
}

// GetSystemInfo 获取缓存中的系统信息
func (repo *SystemInfoRepository) GetSystemInfo() (*vo.SystemInfoRsp, bool) {
	cached, err := repo.Redis().Get(context.Background(), systemInfoCacheKey).Bytes()
	if err != nil {
		return nil, false
	}
	var result vo.SystemInfoRsp
	if err := json.Unmarshal(cached, &result); err != nil {
		return nil, false
	}
	return &result, true
}

// SetSystemInfo 写入系统信息到缓存
func (repo *SystemInfoRepository) SetSystemInfo(info *vo.SystemInfoRsp) {
	data, err := json.Marshal(info)
	if err != nil {
		return
	}
	repo.Redis().Set(context.Background(), systemInfoCacheKey, data, systemInfoCacheTTL)
}

// InvalidateCache 清除缓存
func (repo *SystemInfoRepository) InvalidateCache() {
	repo.Redis().Del(context.Background(), systemInfoCacheKey)
}

// GetDBVersion 获取数据库版本
func (repo *SystemInfoRepository) GetDBVersion(dbType string) string {
	var version string
	db := repo.db()
	switch dbType {
	case "mysql":
		db.Raw("SELECT VERSION()").Scan(&version)
	case "pg":
		db.Raw("SELECT version()").Scan(&version)
	default:
		db.Raw("SELECT sqlite_version()").Scan(&version)
	}
	return version
}

// GetDBName 获取数据库名
func (repo *SystemInfoRepository) GetDBName(dbType string) string {
	switch dbType {
	case "mysql":
		var name string
		repo.db().Raw("SELECT DATABASE()").Scan(&name)
		return name
	case "pg":
		var name string
		repo.db().Raw("SELECT current_database()").Scan(&name)
		return name
	default:
		return filepath.Base(config.Get().Paths.SqliteDB)
	}
}

// GetDBDataSize 获取数据库数据占用大小（字节）
func (repo *SystemInfoRepository) GetDBDataSize(dbType string) int64 {
	switch dbType {
	case "mysql":
		var size int64
		repo.db().Raw("SELECT COALESCE(SUM(data_length), 0) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&size)
		return size
	case "pg":
		var size int64
		repo.db().Raw("SELECT pg_database_size(current_database())").Scan(&size)
		return size
	default:
		if fi, err := os.Stat(config.Get().Paths.SqliteDB); err == nil {
			return fi.Size()
		}
		return 0
	}
}

// GetDBPoolStats 数据库连接池统计
func (repo *SystemInfoRepository) GetDBPoolStats() (*vo.DBPoolInfo, error) {
	var sqlDB *sql.DB
	var err error
	if sqlDB, err = repo.db().DB(); err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	info := &vo.DBPoolInfo{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDurationMs:     stats.WaitDuration.Milliseconds(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
	return info, nil
}

// GetEcosystemCounts 生态统计数据查询
func (repo *SystemInfoRepository) GetEcosystemCounts() (*vo.EcosystemInfo, error) {
	db := repo.db()
	info := &vo.EcosystemInfo{}

	var count int64

	if err := db.Model(&po.User{}).Where("deleted = 0").Count(&count).Error; err != nil {
		return nil, err
	}
	info.TotalUsers = count

	db.Model(&po.User{}).Where("status = 1 AND deleted = 0").Count(&info.ActiveUsers)
	db.Model(&po.User{}).Where("role = 1 AND deleted = 0").Count(&info.AdminUsers)

	db.Model(&po.AIModel{}).Where("deleted = 0").Count(&info.TotalModels)
	db.Model(&po.AIModel{}).Where("status = 1 AND deleted = 0").Count(&info.EnabledModels)

	db.Model(&po.MCPEndpoint{}).Where("deleted = 0").Count(&info.TotalMcps)
	db.Model(&po.MCPEndpoint{}).Where("status = 1 AND deleted = 0").Count(&info.EnabledMcps)

	db.Model(&po.SystemSkill{}).Where("deleted = 0").Count(&info.SystemSkills)
	db.Model(&po.SkillMarket{}).Where("status = 1 AND deleted = 0").Count(&info.MarketSkills)

	db.Model(&po.PersonaTemplate{}).Where("deleted = 0").Count(&info.PersonaTemplates)

	db.Model(&po.Session{}).Where("deleted = 0").Count(&info.TotalSessions)
	db.Model(&po.Message{}).Count(&info.TotalMessages)

	db.Model(&po.TeamProject{}).Count(&info.TotalTeamProjects)

	db.Model(&po.ShareLink{}).Where("deleted = 0").Count(&info.TotalShareLinks)
	db.Model(&po.ShareLink{}).Where("deleted = 0 AND (expires_at IS NULL OR expires_at > ?)", time.Now()).Count(&info.ActiveShareLinks)
	db.Model(&po.ShareLink{}).Where("deleted = 0").Select("COALESCE(SUM(view_count), 0)").Row().Scan(&info.TotalShareViews)

	return info, nil
}

// GetMCPHealthList 查询所有启用 MCP 端点的健康状态
func (repo *SystemInfoRepository) GetMCPHealthList() []vo.MCPHealthItem {
	var endpoints []po.MCPEndpoint
	repo.db().Where("status = 1 AND deleted = 0").Find(&endpoints)

	items := make([]vo.MCPHealthItem, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, vo.MCPHealthItem{
			McpId:           ep.McpId,
			Name:            ep.Name,
			DisplayName:     ep.DisplayName,
			Icon:            ep.Icon,
			HealthLatency:   ep.HealthLatency,
			HealthCheckedAt: ep.HealthCheckedAt,
		})
	}
	return items
}

// GetActiveSessionCount 活跃会话数
func (repo *SystemInfoRepository) GetActiveSessionCount() int {
	var count int64
	repo.db().Model(&po.Session{}).Where("status = 1").Count(&count)
	return int(count)
}

// GetCacheInfo 获取缓存信息（类型 + 内存占用）
func (repo *SystemInfoRepository) GetCacheInfo() (cacheType string, cacheMemory string) {
	cacheType = config.Get().System.CacheType

	if cacheType == "redis" {
		info, err := repo.Redis().Info(context.Background(), "memory").Result()
		if err == nil {
			for _, line := range strings.Split(info, "\r\n") {
				if strings.HasPrefix(line, "used_memory_human:") {
					cacheMemory = strings.TrimPrefix(line, "used_memory_human:")
					break
				}
			}
		}
	} else {
		if counter, ok := repo.Redis().(interface{ ItemCount() int }); ok {
			cacheMemory = fmt.Sprintf("%d items", counter.ItemCount())
		}
	}
	return
}

// db .
func (repo *SystemInfoRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

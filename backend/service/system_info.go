package service

import (
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/config"
	"raven/util/disk"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *SystemInfoService {
			return &SystemInfoService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *SystemInfoService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type SystemInfoService struct {
	Worker	freedom.Worker
	SysRepo	*repository.SystemInfoRepository
}

func (service *SystemInfoService) GetSystemInfo(forceRefresh bool) (*vo.SystemInfoRsp, error) {
	if forceRefresh {
		service.SysRepo.InvalidateCache()
	} else {
		if cached, ok := service.SysRepo.GetSystemInfo(); ok {
			return cached, nil
		}
	}

	info := &vo.SystemInfoRsp{
		CollectedAt: time.Now(),
	}

	info.Overview = service.collectOverview()
	info.Database = service.collectDatabase()
	info.Disks = service.collectDisks()
	info.Plugins = service.collectPlugins()
	info.MCPHealth = service.SysRepo.GetMCPHealthList()

	eco, err := service.SysRepo.GetEcosystemCounts()
	if err != nil {
		return nil, err
	}
	info.Ecosystem = *eco

	service.SysRepo.SetSystemInfo(info)
	return info, nil
}

func (service *SystemInfoService) collectOverview() vo.OverviewInfo {
	cfg := config.Get()
	cacheType, cacheMemory := service.SysRepo.GetCacheInfo()

	info := vo.OverviewInfo{
		Version:	cfg.GetBuildInfo().Version,
		Language:	cfg.GetLanguage(),
		CacheType:	cacheType,
		CacheMemory:	cacheMemory,
	}

	info.ChromaDbBytes = disk.DirSize(cfg.Paths.ChromaDIR)

	uploadDir := cfg.GetUploadDir()
	if uploadDir != "" {
		info.UploadBytes = disk.DirSize(uploadDir)
	}

	info.TempBytes = disk.DirSize(cfg.GetUploadTempDir()) + disk.DirSize(cfg.GetDownloadTempDir())

	return info
}

func (service *SystemInfoService) collectDatabase() vo.DatabaseInfo {
	dbType := config.Get().System.DBType

	info := vo.DatabaseInfo{
		Type:		dbType,
		Version:	service.SysRepo.GetDBVersion(dbType),
		Name:		service.SysRepo.GetDBName(dbType),
	}

	info.DataSizeBytes = service.SysRepo.GetDBDataSize(dbType)

	pool, err := service.SysRepo.GetDBPoolStats()
	if err == nil {
		info.Pool = *pool
	}
	return info
}

func (service *SystemInfoService) collectDisks() []vo.DiskInfo {
	disks := disk.GetMountPoints()
	result := make([]vo.DiskInfo, 0, len(disks))
	for _, d := range disks {
		result = append(result, vo.DiskInfo{
			MountPoint:	d.MountPoint,
			FSType:		d.FSType,
			Device:		d.Device,
			TotalBytes:	d.TotalBytes,
			UsedBytes:	d.UsedBytes,
			FreeBytes:	d.FreeBytes,
			UsedPercent:	d.UsedPercent,
		})
	}
	return result
}

func (service *SystemInfoService) collectPlugins() []vo.PluginInfo {
	result := make([]vo.PluginInfo, 0, 0)
	return result
}

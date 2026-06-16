package service

import (
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/config"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *PreferenceService {
			return &PreferenceService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *PreferenceService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

type PreferenceService struct {
	Worker		freedom.Worker
	SysCfgRepo	*repository.SystemSettingRepository
}

func (service *PreferenceService) GetPreference() *vo.PreferenceRsp {
	domain := ""
	if sysconf, err := service.SysCfgRepo.LoadConfig(); err == nil {
		domain = sysconf.GeneralDomain
	}
	return &vo.PreferenceRsp{
		Language:	config.Get().GetLanguage(),
		Domain:		domain,
	}
}

package service

import (
	"goraven/backend/vo"
	"goraven/config"

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
	Worker freedom.Worker
}

func (service *PreferenceService) GetPreference() *vo.PreferenceRsp {
	return &vo.PreferenceRsp{
		Language: config.Get().GetLanguage(),
	}
}

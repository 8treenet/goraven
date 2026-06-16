package repository

import (
	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Default {
			return &Default{}
		})
	})
}

type Default struct {
	freedom.Repository
}

func (repo *Default) GetIP() string {

	repo.Worker().Logger().Info("I'm Repository GetIP")
	return repo.Worker().IrisContext().RemoteAddr()
}

func (repo *Default) GetUA() string {
	repo.Worker().Logger().Info("I'm Repository GetUA")
	return repo.Worker().IrisContext().Request().UserAgent()
}

func (repo *Default) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

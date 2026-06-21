package repository

import (
	"raven/backend/po"
	"raven/backend/vo"
	"raven/util"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *ShareLinkRepository {
			return &ShareLinkRepository{}
		})
	})
}

type ShareLinkRepository struct {
	freedom.Repository
}

func (repo *ShareLinkRepository) CreateShareLink(shareLink *po.ShareLink) error {
	if shareLink.ShareId == "" {
		shareLink.ShareId = util.UUID()
	}
	return repo.db().Create(shareLink).Error
}

func (repo *ShareLinkRepository) GetShareLink(shareId string) (*po.ShareLink, error) {
	var shareLink po.ShareLink
	err := repo.db().First(&shareLink, "share_id = ? AND deleted = 0", shareId).Error
	return &shareLink, err
}

func (repo *ShareLinkRepository) GetSessionShare(sessionId string, userId string) (*po.ShareLink, error) {
	var shareLink po.ShareLink
	err := repo.db().First(&shareLink, "session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Error
	return &shareLink, err
}

func (repo *ShareLinkRepository) DeleteShareLink(sessionId string, userId string) error {
	return repo.db().Model(&po.ShareLink{}).
		Where("session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).
		Updates(map[string]interface{}{
			"deleted": 1,
			"updated": time.Now(),
		}).Error
}

func (repo *ShareLinkRepository) UpdateShareLink(shareId string, title string, shareType string, expiresAt time.Time) (*po.ShareLink, error) {
	if err := repo.db().Model(&po.ShareLink{}).
		Where("share_id = ? AND deleted = 0", shareId).
		Updates(map[string]interface{}{
			"title":      title,
			"share_type": shareType,
			"expires_at": expiresAt,
			"view_count": 0,
			"updated":    time.Now(),
		}).Error; err != nil {
		return nil, err
	}
	var shareLink po.ShareLink
	if err := repo.db().First(&shareLink, "share_id = ? AND deleted = 0", shareId).Error; err != nil {
		return nil, err
	}
	return &shareLink, nil
}

func (repo *ShareLinkRepository) ListUserShareLinks(userId string, req *vo.UserShareListReq) ([]po.ShareLink, *PageResult, error) {
	query := repo.db().Model(&po.ShareLink{}).Where("user_id = ? AND deleted = 0", userId)
	var shareLinks []po.ShareLink
	pr, err := Paginate(query.Order("created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &shareLinks)
	if err != nil {
		return nil, nil, err
	}
	return shareLinks, pr, nil
}

func (repo *ShareLinkRepository) IncrementViewCount(shareId string) error {
	return repo.db().Model(&po.ShareLink{}).
		Where("share_id = ? AND deleted = 0", shareId).
		Updates(map[string]interface{}{
			"view_count": gorm.Expr("view_count + 1"),
			"updated":    time.Now(),
		}).Error
}

func (repo *ShareLinkRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

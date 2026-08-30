package repository

import (
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/util"
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

// ShareLinkRepository 对话分享链接仓储
type ShareLinkRepository struct {
	freedom.Repository
}

// CreateShareLink 创建分享链接
func (repo *ShareLinkRepository) CreateShareLink(shareLink *po.ShareLink) error {
	if shareLink.ShareId == "" {
		shareLink.ShareId = util.UUID()
	}
	return repo.db().Create(shareLink).Error
}

// GetShareLink 根据 shareId 获取分享链接（无鉴权，公开访问用）
func (repo *ShareLinkRepository) GetShareLink(shareId string) (*po.ShareLink, error) {
	var shareLink po.ShareLink
	err := repo.db().First(&shareLink, "share_id = ? AND deleted = 0", shareId).Error
	return &shareLink, err
}

// GetSessionShare 获取会话的分享链接（需校验 userId）
func (repo *ShareLinkRepository) GetSessionShare(sessionId string, userId string) (*po.ShareLink, error) {
	var shareLink po.ShareLink
	err := repo.db().First(&shareLink, "session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Error
	return &shareLink, err
}

// UpdateShareLink 更新分享链接的标题、类型和过期时间，重置浏览次数
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

// DeleteShareLink 删除分享链接（软删除，校验 userId）
func (repo *ShareLinkRepository) DeleteShareLink(sessionId string, userId string) error {
	return repo.db().Model(&po.ShareLink{}).
		Where("session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).
		Updates(map[string]interface{}{
			"deleted": 1,
			"updated": time.Now(),
		}).Error
}

// ListUserShareLinks 获取用户分享链接分页列表
func (repo *ShareLinkRepository) ListUserShareLinks(userId string, req *vo.UserShareListReq) ([]po.ShareLink, *PageResult, error) {
	query := repo.db().Model(&po.ShareLink{}).Where("user_id = ? AND deleted = 0", userId)
	var shareLinks []po.ShareLink
	pr, err := Paginate(query.Order("created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &shareLinks)
	if err != nil {
		return nil, nil, err
	}
	return shareLinks, pr, nil
}

// IncrementViewCount 增加浏览次数
func (repo *ShareLinkRepository) IncrementViewCount(shareId string) error {
	return repo.db().Model(&po.ShareLink{}).
		Where("share_id = ? AND deleted = 0", shareId).
		Updates(map[string]interface{}{
			"view_count": gorm.Expr("view_count + 1"),
			"updated":   time.Now(),
		}).Error
}

func (repo *ShareLinkRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

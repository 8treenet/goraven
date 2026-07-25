package repository

import (
	"context"
	"encoding/json"
	"goraven/backend/po"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

const (
	tempAkKeyPrefix = "hfs:tempak:"

	TempAkTTL = 15 * time.Minute
)

type TempAccessCache struct {
	UserName string `json:"userName"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *HFSRepository {
			return &HFSRepository{}
		})
	})
}

type HFSRepository struct {
	freedom.Repository
}

func (repo *HFSRepository) CreateFileLink(link *po.FileLink) error {
	return repo.db().Create(link).Error
}

func (repo *HFSRepository) GetFileLinkByLinkId(linkId string) (*po.FileLink, error) {
	var link po.FileLink
	err := repo.db().First(&link, "link_id = ?", linkId).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (repo *HFSRepository) GetFileLinkByPath(userId, filePath string) (*po.FileLink, error) {
	var link po.FileLink
	err := repo.db().Where("user_id = ? AND file_path = ?", userId, filePath).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (repo *HFSRepository) CreateUpload(upload *po.ChunkUpload) error {
	return repo.db().Create(upload).Error
}

func (repo *HFSRepository) GetUploadByUploadId(uploadId string) (*po.ChunkUpload, error) {
	var upload po.ChunkUpload
	err := repo.db().Where("upload_id = ? AND deleted = ?", uploadId, 0).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

func (repo *HFSRepository) UpdateUploadStatus(uploadId string, status uint8) error {
	return repo.db().Model(&po.ChunkUpload{}).
		Where("upload_id = ?", uploadId).
		Update("status", status).Error
}

func (repo *HFSRepository) SoftDeleteUpload(uploadId string) error {
	return repo.db().Model(&po.ChunkUpload{}).
		Where("upload_id = ?", uploadId).
		Update("deleted", 1).Error
}

func (repo *HFSRepository) SoftDeleteExpiredUploads(days int) error {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	return repo.db().Model(&po.ChunkUpload{}).
		Where("deleted = ? AND created < ?", 0, cutoff).
		Update("deleted", 1).Error
}

func (repo *HFSRepository) MarkUploadUsed(uploadId string) error {
	return repo.db().Model(&po.ChunkUpload{}).
		Where("upload_id = ?", uploadId).
		Updates(map[string]interface{}{
			"status": po.UploadStatusUsed,
		}).Error
}

func (repo *HFSRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}

func (repo *HFSRepository) SetTempAccess(ak string, data *TempAccessCache) error {
	key := tempAkKeyPrefix + ak
	val, _ := json.Marshal(data)
	return repo.Redis().Set(context.Background(), key, val, TempAkTTL).Err()
}

func (repo *HFSRepository) GetTempAccess(ak string) (*TempAccessCache, error) {
	key := tempAkKeyPrefix + ak
	data, err := repo.Redis().Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, nil
	}
	var cache TempAccessCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func (repo *HFSRepository) DeleteTempAccess(ak string) {
	repo.Redis().Del(context.Background(), tempAkKeyPrefix+ak)
}

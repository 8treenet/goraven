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
	// TempAkTTL 临时访问凭证有效期
	TempAkTTL = 15 * time.Minute
)

// 临时凭证空间类型
const (
	// TempSpaceUser 用户空间凭证，Path 为用户工作区相对路径
	TempSpaceUser = "user"
	// TempSpaceShared 共享空间（团队项目）凭证，Path 为 ak 空间路径 /projects/<项目名>/...
	TempSpaceShared = "shared"
)

// TempAccessCache 临时访问凭证缓存数据
type TempAccessCache struct {
	UserName string `json:"userName"`
	Path     string `json:"path"`  // 凭证绑定路径，格式由 Space 决定
	Type     string `json:"type"`  // "file" 或 "dir"
	Space    string `json:"space"` // TempSpaceUser 用户空间 / TempSpaceShared 共享空间
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

// CreateFileLink 创建文件外链记录
func (repo *HFSRepository) CreateFileLink(link *po.FileLink) error {
	return repo.db().Create(link).Error
}

// GetFileLinkByLinkId 通过外链ID查询记录
func (repo *HFSRepository) GetFileLinkByLinkId(linkId string) (*po.FileLink, error) {
	var link po.FileLink
	err := repo.db().First(&link, "link_id = ?", linkId).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetFileLinkByPath 通过用户ID和文件路径查询记录
func (repo *HFSRepository) GetFileLinkByPath(userId, filePath string) (*po.FileLink, error) {
	var link po.FileLink
	err := repo.db().Where("user_id = ? AND file_path = ?", userId, filePath).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// CreateUpload 创建分片上传任务
func (repo *HFSRepository) CreateUpload(upload *po.ChunkUpload) error {
	return repo.db().Create(upload).Error
}

// GetUploadByUploadId 通过uploadId查询上传任务
func (repo *HFSRepository) GetUploadByUploadId(uploadId string) (*po.ChunkUpload, error) {
	var upload po.ChunkUpload
	err := repo.db().Where("upload_id = ? AND deleted = ?", uploadId, 0).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

// UpdateUploadStatus 更新上传任务状态
func (repo *HFSRepository) UpdateUploadStatus(uploadId string, status uint8) error {
	return repo.db().Model(&po.ChunkUpload{}).
		Where("upload_id = ?", uploadId).
		Update("status", status).Error
}

// SoftDeleteUpload 软删除上传任务
func (repo *HFSRepository) SoftDeleteUpload(uploadId string) error {
	return repo.db().Model(&po.ChunkUpload{}).
		Where("upload_id = ?", uploadId).
		Update("deleted", 1).Error
}

// SoftDeleteExpiredUploads 软删除超过指定天数的上传任务
func (repo *HFSRepository) SoftDeleteExpiredUploads(days int) error {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	return repo.db().Model(&po.ChunkUpload{}).
		Where("deleted = ? AND created < ?", 0, cutoff).
		Update("deleted", 1).Error
}

// MarkUploadUsed 标记上传任务已使用
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

// SetTempAccess 写入临时访问凭证，TempAkTTL 后过期
func (repo *HFSRepository) SetTempAccess(ak string, data *TempAccessCache) error {
	key := tempAkKeyPrefix + ak
	val, _ := json.Marshal(data)
	return repo.Redis().Set(context.Background(), key, val, TempAkTTL).Err()
}

// GetTempAccess 读取临时访问凭证，不存在或过期返回 nil, nil
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

// DeleteTempAccess 删除临时访问凭证
func (repo *HFSRepository) DeleteTempAccess(ak string) {
	repo.Redis().Del(context.Background(), tempAkKeyPrefix+ak)
}

package vo

import "time"

type FileManagerListReq struct {
	Dir   string `url:"dir"`
	Sort  string `url:"sort"`
	Order string `url:"order"`
}

type FileManagerListItem struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"isDir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	IsDefault bool      `json:"isDefault,omitempty"`
	IsShared  bool      `json:"isShared,omitempty"`
	SharedId  int       `json:"sharedId,omitempty"`
}

type FileManagerListRsp struct {
	Items []FileManagerListItem `json:"items"`
}

type FileManagerUploadReq struct {
	UploadId string `json:"uploadId" validate:"required"`
	Dir      string `json:"dir"`
}

type FileManagerUploadRsp struct {
	Path string `json:"path"`
}

type FileManagerMkdirReq struct {
	Path string `json:"path" validate:"required"`
}

type FileManagerRenameReq struct {
	OldPath string `json:"oldPath" validate:"required"`
	NewPath string `json:"newPath" validate:"required"`
}

type FileManagerDeleteReq struct {
	Paths []string `json:"paths" validate:"required"`
}

type FileManagerCompressReq struct {
	Paths      []string `json:"paths" validate:"required"`
	OutputName string   `json:"outputName" validate:"required"`
}

type FileManagerCompressRsp struct {
	ZipPath string `json:"zipPath"`
}

type FileManagerDecompressReq struct {
	Path     string `json:"path" validate:"required"`
	ToSubDir bool   `json:"toSubDir"`
}

type FileManagerProfileEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FileManagerProfileListRsp struct {
	Items []FileManagerProfileEntry `json:"items"`
}

type FileManagerProfileCreateReq struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type FileManagerProfileUpdateReq struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

type FileManagerProfileDeleteReq struct {
	Key string `json:"key" validate:"required"`
}

type FileManagerUsageRsp struct {
	TotalSize int64 `json:"totalSize"`
	UsedSize  int64 `json:"usedSize"`
	FileCount int   `json:"fileCount"`
}

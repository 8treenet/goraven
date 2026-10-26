package vo

import "time"

// FileManagerListReq 文件列表请求
type FileManagerListReq struct {
	Dir   string `url:"dir"`   // 目录路径，为空时列出根目录
	Sort  string `url:"sort"`  // 排序字段：name/size/time，默认 name
	Order string `url:"order"` // 升降序：asc/desc，默认 asc
}

// FileManagerListItem 文件列表项
type FileManagerListItem struct {
	Name      string    `json:"name"`                // 文件名
	Path      string    `json:"path"`                // 绝对路径（相对于用户工作空间根目录）
	IsDir     bool      `json:"isDir"`               // 是否为目录
	Size      int64     `json:"size"`                // 文件大小（字节），目录为 0
	ModTime   time.Time `json:"modTime"`             // 修改时间
	IsDefault bool      `json:"isDefault,omitempty"` // 是否为系统初始化创建的受保护条目（仅根目录下生效）
}

// FileManagerListRsp 文件列表响应
type FileManagerListRsp struct {
	Items []FileManagerListItem `json:"items"` // 文件列表
}

// FileManagerUploadReq 提交文件到用户空间请求
type FileManagerUploadReq struct {
	UploadId string `json:"uploadId" validate:"required"` // HFS 分片上传合并后的 uploadId
	Dir      string `json:"dir"`                          // 目标目录，为空时放入根目录
}

// FileManagerUploadRsp 提交文件到用户空间响应
type FileManagerUploadRsp struct {
	Path string `json:"path"` // 文件在用户空间中的相对路径
}

// FileManagerMkdirReq 创建目录请求
type FileManagerMkdirReq struct {
	Path string `json:"path" validate:"required"` // 目录路径
}

// FileManagerRenameReq 重命名请求
type FileManagerRenameReq struct {
	OldPath string `json:"oldPath" validate:"required"` // 原路径
	NewPath string `json:"newPath" validate:"required"` // 新路径
}

// FileManagerDeleteReq 删除文件/目录请求
type FileManagerDeleteReq struct {
	Paths []string `json:"paths" validate:"required"` // 要删除的路径列表
}

// FileManagerCompressReq 压缩请求
type FileManagerCompressReq struct {
	Paths      []string `json:"paths" validate:"required"`      // 要压缩的路径列表
	OutputName string   `json:"outputName" validate:"required"` // 压缩包名称
}

// FileManagerCompressRsp 压缩响应
type FileManagerCompressRsp struct {
	ZipPath string `json:"zipPath"` // 生成的 zip 文件相对路径
}

// FileManagerDecompressReq 解压请求
type FileManagerDecompressReq struct {
	Path     string `json:"path" validate:"required"` // zip 文件路径
	ToSubDir bool   `json:"toSubDir"`                 // 是否解压到同名子目录
}

// FileManagerProfileEntry 环境变量条目
type FileManagerProfileEntry struct {
	Key   string `json:"key"`   // 环境变量名
	Value string `json:"value"` // 环境变量值
}

// FileManagerProfileListRsp 环境变量列表响应
type FileManagerProfileListRsp struct {
	Items []FileManagerProfileEntry `json:"items"`
}

// FileManagerProfileCreateReq 新增环境变量请求
type FileManagerProfileCreateReq struct {
	Key   string `json:"key" validate:"required"`   // 环境变量名
	Value string `json:"value"`                      // 环境变量值
}

// FileManagerProfileUpdateReq 更新环境变量请求
type FileManagerProfileUpdateReq struct {
	Key   string `json:"key" validate:"required"`   // 环境变量名
	Value string `json:"value"`                      // 环境变量新值
}

// FileManagerProfileDeleteReq 删除环境变量请求
type FileManagerProfileDeleteReq struct {
	Key string `json:"key" validate:"required"` // 环境变量名
}

// FileManagerUsageRsp 磁盘使用统计响应
type FileManagerUsageRsp struct {
	TotalSize int64 `json:"totalSize"` // 总大小（字节）
	UsedSize  int64 `json:"usedSize"`  // 已使用大小（字节）
	FileCount int   `json:"fileCount"` // 文件数量
}

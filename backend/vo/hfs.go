package vo

// ChunkUploadCreateReq 创建上传任务请求
type ChunkUploadCreateReq struct {
	FileName    string `json:"fileName" validate:"required"`
	FileSize    int64  `json:"fileSize" validate:"required"`
	ChunkSize   int    `json:"chunkSize" validate:"required"`
	TotalChunks int    `json:"totalChunks" validate:"required"`
}

// ChunkUploadCreateRsp 创建上传任务响应
type ChunkUploadCreateRsp struct {
	UploadId string `json:"uploadId"`
	TempDir  string `json:"tempDir"`
}

// ChunkUploadReq 上传分片请求
type ChunkUploadReq struct {
	UploadId   string `form:"uploadId" validate:"required"`
	ChunkIndex int    `form:"chunkIndex"`
	ChunkMd5   string `form:"chunkMd5"`
}

// ChunkMergeReq 合并上传请求
type ChunkMergeReq struct {
	UploadId string `json:"uploadId" validate:"required"`
}

// ChunkMergeRsp 合并上传响应
type ChunkMergeRsp struct {
	UploadId string `json:"uploadId"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

// AssetsReq 提交静态资源请求
type AssetsReq struct {
	UploadId string `json:"uploadId" validate:"required"`
}

// AssetsRsp 提交静态资源响应
type AssetsRsp struct {
	Path string `json:"path"`
}

// TempAccessReq 申请临时访问凭证请求
type TempAccessReq struct {
	Path string `json:"path" validate:"required"` // 用户空间相对路径
	Type string `json:"type" validate:"required"` // "file" 或 "dir"
}

// TempAccessRsp 申请临时访问凭证响应
type TempAccessRsp struct {
	Ak        string `json:"ak"`
	ExpiresAt int64  `json:"expiresAt"` // 过期时间（unix 秒）
}

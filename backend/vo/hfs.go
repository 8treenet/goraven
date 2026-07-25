package vo

type ChunkUploadCreateReq struct {
	FileName    string `json:"fileName" validate:"required"`
	FileSize    int64  `json:"fileSize" validate:"required"`
	ChunkSize   int    `json:"chunkSize" validate:"required"`
	TotalChunks int    `json:"totalChunks" validate:"required"`
}

type ChunkUploadCreateRsp struct {
	UploadId string `json:"uploadId"`
	TempDir  string `json:"tempDir"`
}

type ChunkUploadReq struct {
	UploadId   string `form:"uploadId" validate:"required"`
	ChunkIndex int    `form:"chunkIndex"`
	ChunkMd5   string `form:"chunkMd5"`
}

type ChunkMergeReq struct {
	UploadId string `json:"uploadId" validate:"required"`
}

type ChunkMergeRsp struct {
	UploadId string `json:"uploadId"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

type AssetsReq struct {
	UploadId string `json:"uploadId" validate:"required"`
}

type AssetsRsp struct {
	Path string `json:"path"`
}

type TempAccessReq struct {
	Path string `json:"path" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type TempAccessRsp struct {
	Ak        string `json:"ak"`
	ExpiresAt int64  `json:"expiresAt"`
}

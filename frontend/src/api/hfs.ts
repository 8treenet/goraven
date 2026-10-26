import http from './http'

/** POST /api/hfs/upload/create */
export function createUpload(data: {
  fileName: string
  fileSize: number
  chunkSize: number
  totalChunks: number
}) {
  return http.post<{ uploadId: string; tempDir: string }>('/hfs/upload/create', data)
}

/** PUT /api/hfs/upload/chunk */
export function uploadChunk(uploadId: string, chunkIndex: number, blob: Blob, chunkMd5?: string) {
  const form = new FormData()
  form.append('file', blob)

  const params = new URLSearchParams()
  params.set('uploadId', uploadId)
  params.set('chunkIndex', String(chunkIndex))
  if (chunkMd5) params.set('chunkMd5', chunkMd5)

  return http.put<{ status: string }>(`/hfs/upload/chunk?${params.toString()}`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** POST /api/hfs/upload/merge */
export function mergeUpload(uploadId: string) {
  return http.post<{
    uploadId: string
    filePath: string
    fileName: string
    fileSize: number
  }>('/hfs/upload/merge', { uploadId })
}

/** POST /api/hfs/assets — 将合并后的临时文件转为静态资源 */
export function commitAssets(uploadId: string) {
  return http.post<{ path: string }>('/hfs/assets', { uploadId })
}

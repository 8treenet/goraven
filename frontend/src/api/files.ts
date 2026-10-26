import http from './http'
import type { FileItem, StorageUsage, ProfileListResponse } from './types'

export interface FileListResponse {
  items: FileItem[]
}

/** GET /api/fileManager/list */
export function listFiles(dir?: string, sort?: string, order?: string) {
  return http.get<FileListResponse>('/fileManager/list', { params: { dir, sort, order } })
}

/** POST /api/fileManager/mkdir */
export function mkdir(path: string) {
  return http.post('/fileManager/mkdir', { path })
}

/** PUT /api/fileManager/rename */
export function rename(oldPath: string, newPath: string) {
  return http.put('/fileManager/rename', { oldPath, newPath })
}

/** DELETE /api/fileManager/delete */
export function deleteFiles(paths: string[]) {
  return http.delete('/fileManager/delete', { data: { paths } })
}

/** POST /api/fileManager/compress */
export function compress(data: { paths: string[]; outputName: string }) {
  return http.post<{ zipPath: string }>('/fileManager/compress', data)
}

/** POST /api/fileManager/decompress */
export function decompress(data: { path: string; toSubDir?: boolean }) {
  return http.post('/fileManager/decompress', data)
}

/** GET /api/fileManager/usage */
export function getUsage() {
  return http.get<StorageUsage>('/fileManager/usage')
}

/** POST /api/fileManager/upload — 将 HFS 分片合并后的文件移入用户空间 */
export function commitUpload(uploadId: string, dir?: string) {
  return http.post<{ path: string }>('/fileManager/upload', { uploadId, dir })
}

/** GET /api/hfs/private/<path> — 构建文件下载 URL（需配合 fetch + Bearer token 使用） */
export function getDownloadUrl(path: string): string {
  const segments = path.replace(/^\/+/, '').split('/').filter(Boolean)
  return `/api/hfs/private/${segments.map(encodeURIComponent).join('/')}`
}

/** GET /api/hfs/file/<abs-path> — 按文件系统绝对路径构建下载 URL（需配合 fetch + Bearer token 使用）
 *  用于聊天内 <goraven-file> 跨用户共享场景，任意已登录用户均可访问 user_space 下的文件。 */
export function getFileUrl(absPath: string): string {
  const segments = absPath.replace(/^\/+/, '').split('/').filter(Boolean)
  return `/api/hfs/file/${segments.map(encodeURIComponent).join('/')}`
}

/** POST /api/hfs/access — 申请临时访问凭证（15 分钟有效） */
export function createTempAccess(path: string, type: 'file' | 'dir') {
  return http.post<{ ak: string; expiresAt: number }>('/hfs/access', { path, type })
}

/** 构建临时凭证下载 URL（无需 Bearer token，ak 在 URL 路径中） */
export function getAkDownloadUrl(ak: string, path: string): string {
  const segments = path.replace(/^\/+/, '').split('/').filter(Boolean)
  return `/api/hfs/ak/${ak}/${segments.map(encodeURIComponent).join('/')}`
}

/** GET /api/fileManager/profile */
export function getProfile() {
  return http.get<ProfileListResponse>('/fileManager/profile')
}

/** POST /api/fileManager/profile */
export function createProfile(key: string, value: string) {
  return http.post('/fileManager/profile', { key, value })
}

/** PUT /api/fileManager/profile */
export function updateProfile(key: string, value: string) {
  return http.put('/fileManager/profile', { key, value })
}

/** DELETE /api/fileManager/profile */
export function deleteProfile(key: string) {
  return http.delete('/fileManager/profile', { data: { key } })
}

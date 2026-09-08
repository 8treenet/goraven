import http from './http'
import type {
  MyProjectItem,
  MyProjectListRsp,
  MyProjectCreateRsp,
  FileItem,
  StorageUsage,
} from './types'

export interface MyFileListResponse {
  items: FileItem[]
}

/* ---------- 项目管理 ---------- */

/** GET /api/myProject/list */
export function listMyProjects() {
  return http.get<MyProjectListRsp>('/myProject/list')
}

/** GET /api/myProject/:id */
export function getMyProject(id: number) {
  return http.get<MyProjectItem>(`/myProject/${id}`)
}

/** POST /api/myProject/create */
export function createMyProject(projectName: string, description: string) {
  return http.post<MyProjectCreateRsp>('/myProject/create', { projectName, description })
}

/** PUT /api/myProject/:id — 更新简介和/或改名（projectName 为空表示不改名） */
export function updateMyProject(id: number, data: { projectName?: string; description?: string }) {
  return http.put(`/myProject/${id}`, data)
}

/** DELETE /api/myProject/:id */
export function deleteMyProject(id: number) {
  return http.delete(`/myProject/${id}`)
}

/* ---------- 文件操作 ---------- */

/** GET /api/myProject/:id/list */
export function listMyFiles(id: number, dir?: string, sort?: string, order?: string) {
  return http.get<MyFileListResponse>(`/myProject/${id}/list`, { params: { dir, sort, order } })
}

/** POST /api/myProject/:id/upload */
export function commitMyUpload(id: number, uploadId: string, dir?: string) {
  return http.post<{ path: string }>(`/myProject/${id}/upload`, { uploadId, dir })
}

/** POST /api/myProject/:id/mkdir */
export function myMkdir(id: number, path: string) {
  return http.post(`/myProject/${id}/mkdir`, { path })
}

/** PUT /api/myProject/:id/rename */
export function myRename(id: number, oldPath: string, newPath: string) {
  return http.put(`/myProject/${id}/rename`, { oldPath, newPath })
}

/** DELETE /api/myProject/:id/delete */
export function myDeleteFiles(id: number, paths: string[]) {
  return http.delete(`/myProject/${id}/delete`, { data: { paths } })
}

/** POST /api/myProject/:id/compress */
export function myCompress(id: number, data: { paths: string[]; outputName: string }) {
  return http.post<{ zipPath: string }>(`/myProject/${id}/compress`, data)
}

/** POST /api/myProject/:id/decompress */
export function myDecompress(id: number, data: { path: string; toSubDir?: boolean }) {
  return http.post(`/myProject/${id}/decompress`, data)
}

/** GET /api/myProject/:id/usage */
export function getMyUsage(id: number) {
  return http.get<StorageUsage>(`/myProject/${id}/usage`)
}

/* ---------- 下载与预览 ---------- */

/** GET /api/myProject/:id/download/:path — 构建文件下载 URL（需配合 fetch + Bearer token 使用） */
export function getMyDownloadUrl(id: number, path: string): string {
  const segments = path.replace(/^\/+/, '').split('/').filter(Boolean)
  return `/api/myProject/${id}/download/${segments.map(encodeURIComponent).join('/')}`
}

/** POST /api/myProject/:id/access — 申请临时访问凭证（15 分钟有效，用于 iframe/office 预览） */
export function createMyTempAccess(id: number, path: string, type: 'file' | 'dir') {
  return http.post<{ ak: string; expiresAt: number }>(`/myProject/${id}/access`, { path, type })
}
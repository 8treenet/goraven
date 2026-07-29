import http from './http'
import type {
  TeamProjectItem,
  TeamProjectListRsp,
  TeamProjectCreateRsp,
  FileItem,
  StorageUsage,
} from './types'

export interface TeamFileListResponse {
  items: FileItem[]
}

/* ---------- 项目管理 ---------- */

/** GET /api/teamProject/list */
export function listTeamProjects() {
  return http.get<TeamProjectListRsp>('/teamProject/list')
}

/** GET /api/teamProject/:id */
export function getTeamProject(id: number) {
  return http.get<TeamProjectItem>(`/teamProject/${id}`)
}

/** POST /api/teamProject/create */
export function createTeamProject(projectName: string, description: string) {
  return http.post<TeamProjectCreateRsp>('/teamProject/create', { projectName, description })
}

/** DELETE /api/teamProject/:id */
export function deleteTeamProject(id: number) {
  return http.delete(`/teamProject/${id}`)
}

/** PUT /api/teamProject/:id */
export function updateProjectDescription(id: number, description: string) {
  return http.put(`/teamProject/${id}`, { description })
}

/* ---------- 文件操作 ---------- */

/** GET /api/teamProject/:id/list */
export function listTeamFiles(id: number, dir?: string, sort?: string, order?: string) {
  return http.get<TeamFileListResponse>(`/teamProject/${id}/list`, { params: { dir, sort, order } })
}

/** POST /api/teamProject/:id/upload */
export function commitTeamUpload(id: number, uploadId: string, dir?: string) {
  return http.post<{ path: string }>(`/teamProject/${id}/upload`, { uploadId, dir })
}

/** POST /api/teamProject/:id/mkdir */
export function teamMkdir(id: number, path: string) {
  return http.post(`/teamProject/${id}/mkdir`, { path })
}

/** PUT /api/teamProject/:id/rename */
export function teamRename(id: number, oldPath: string, newPath: string) {
  return http.put(`/teamProject/${id}/rename`, { oldPath, newPath })
}

/** DELETE /api/teamProject/:id/delete */
export function teamDeleteFiles(id: number, paths: string[]) {
  return http.delete(`/teamProject/${id}/delete`, { data: { paths } })
}

/** POST /api/teamProject/:id/compress */
export function teamCompress(id: number, data: { paths: string[]; outputName: string }) {
  return http.post<{ zipPath: string }>(`/teamProject/${id}/compress`, data)
}

/** POST /api/teamProject/:id/decompress */
export function teamDecompress(id: number, data: { path: string; toSubDir?: boolean }) {
  return http.post(`/teamProject/${id}/decompress`, data)
}

/** GET /api/teamProject/:id/usage */
export function getTeamUsage(id: number) {
  return http.get<StorageUsage>(`/teamProject/${id}/usage`)
}

/* ---------- 下载与预览 ---------- */

/** GET /api/teamProject/:id/download/:path — 构建文件下载 URL（需配合 fetch + Bearer token 使用） */
export function getTeamDownloadUrl(id: number, path: string): string {
  const segments = path.replace(/^\/+/, '').split('/').filter(Boolean)
  return `/api/teamProject/${id}/download/${segments.map(encodeURIComponent).join('/')}`
}

/** POST /api/teamProject/:id/access — 申请临时访问凭证（15 分钟有效，用于 iframe/office 预览） */
export function createTeamTempAccess(id: number, path: string, type: 'file' | 'dir') {
  return http.post<{ ak: string; expiresAt: number }>(`/teamProject/${id}/access`, { path, type })
}

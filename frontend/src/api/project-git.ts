import http from './http'
import type {
  GitStatus,
  GitTestPayload,
  GitTestRsp,
  GitCommitRsp,
  GitLogRsp,
  GitDiffRsp,
} from './types'

/** 项目 Git API：个人项目与团队项目路由对称，按 basePath 生成一套 */
export interface ProjectGitApi {
  getStatus: (projectId: number) => Promise<GitStatus>
  testRemoteDirect: (payload: GitTestPayload) => Promise<GitTestRsp>
  retryClone: (projectId: number) => Promise<unknown>
  resolveUnrelated: (projectId: number, action: 'merge' | 'local_only') => Promise<unknown>
  commit: (projectId: number, message: string, push: boolean) => Promise<GitCommitRsp>
  push: (projectId: number) => Promise<unknown>
  pull: (projectId: number) => Promise<unknown>
  log: (projectId: number, limit: number, offset: number) => Promise<GitLogRsp>
  diff: (projectId: number, path: string, commit?: string) => Promise<GitDiffRsp>
}

export function createGitApi(base: string): ProjectGitApi {
  return {
    getStatus: (projectId) => http.get<GitStatus>(`${base}/${projectId}/git`),
    testRemoteDirect: (payload) => http.post<GitTestRsp>(`${base}/git/test`, payload),
    retryClone: (projectId) => http.post(`${base}/${projectId}/git/clone-retry`),
    resolveUnrelated: (projectId, action) => http.post(`${base}/${projectId}/git/resolve-unrelated`, { action }),
    commit: (projectId, message, push) => http.post<GitCommitRsp>(`${base}/${projectId}/git/commit`, { message, push }),
    push: (projectId) => http.post(`${base}/${projectId}/git/push`),
    pull: (projectId) => http.post(`${base}/${projectId}/git/pull`),
    log: (projectId, limit, offset) => http.get<GitLogRsp>(`${base}/${projectId}/git/log`, { params: { limit, offset } }),
    diff: (projectId, path, commit = '') => http.get<GitDiffRsp>(`${base}/${projectId}/git/diff`, { params: { path, commit } }),
  }
}

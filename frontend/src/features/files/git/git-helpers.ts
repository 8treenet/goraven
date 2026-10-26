import { t as translate } from '@/i18n'
import type { GitChange } from '@/api/types'

/** 后端 CloneState：0非克隆 1克隆中 2成功 3失败 */
export const GIT_CLONE_STATE = { NONE: 0, RUNNING: 1, SUCCESS: 2, FAILED: 3 } as const
/** 认证方式：0公开 1SSH 2HTTPS */
export const GIT_AUTH = { NONE: 0, SSH: 1, HTTPS: 2 } as const

export const GIT_PAGE_SIZE = 20

/** GitInfoPanel 与 ShareDialog 克隆表单共用的认证方式选项 */
export const AUTH_OPTIONS = [
  { value: GIT_AUTH.NONE, labelKey: 'files.gitAuthNone' },
  { value: GIT_AUTH.SSH, labelKey: 'files.gitAuthSSH' },
  { value: GIT_AUTH.HTTPS, labelKey: 'files.gitAuthHTTPS' },
] as const

/** 变更角标颜色 */
export function changeBadgeClass(status: string): string {
  switch (status) {
    case 'A':
    case 'U':
      return 'text-success'
    case 'D':
      return 'text-destructive'
    default: // M/R/C
      return 'text-warning'
  }
}

/** 变更角标 title（本地化说明） */
export function changeBadgeLabel(status: string): string {
  switch (status) {
    case 'A':
      return translate('files.gitChangeA')
    case 'D':
      return translate('files.gitChangeD')
    case 'U':
      return translate('files.gitChangeU')
    case 'R':
      return translate('files.gitChangeR')
    case 'C':
      return translate('files.gitChangeC')
    default:
      return translate('files.gitChangeM')
  }
}

/** path → change 映射，供文件行角标查询 */
export function buildChangeMap(changes: GitChange[]): Map<string, GitChange> {
  return new Map(changes.map((c) => [c.path, c]))
}

/**
 * 判断错误是否为「历史无关」（后端消息跟随系统语言，双语匹配）。
 * zh: "远程与本地历史无关…" en: "...histories are unrelated..."
 */
export function isUnrelatedHistoryError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err)
  return msg.includes('histories are unrelated') || msg.includes('远程与本地历史无关')
}

/** 从远程 URL 推导项目名：git@github.com:org/repo.git → repo */
export function deriveProjectNameFromUrl(url: string): string {
  const trimmed = url.trim().replace(/\/+$/, '').replace(/\.git$/i, '')
  return trimmed.split(/[:/]/).pop() ?? ''
}

/** 认证类型是否需要凭据 */
export function authNeedsCredential(authType: number): boolean {
  return authType === GIT_AUTH.SSH || authType === GIT_AUTH.HTTPS
}

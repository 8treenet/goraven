import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { ArrowDownToLine, ArrowUpFromLine, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import type { GitChange, GitStatus } from '@/api/types'
import type { ProjectGitApi } from '@/api/project-git'
import { formatSize } from '../file-helpers'
import { changeBadgeClass, changeBadgeLabel, isUnrelatedHistoryError } from './git-helpers'

interface GitChangesPanelProps {
  gitApi: ProjectGitApi
  projectId: number
  status: GitStatus
  /** 提交/推送/拉取完成后回调（父级刷新状态与文件列表） */
  onMutated: (opts?: { pulled?: boolean }) => void
  /** 「历史无关」错误回调：由抽屉弹选择对话框 */
  onUnrelated: () => void
}

export function GitChangesPanel({ gitApi, projectId, status, onMutated, onUnrelated }: GitChangesPanelProps) {
  const t = useT()
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState<'commit' | 'commitPush' | 'push' | 'pull' | null>(null)

  const dirty = status.changes.length > 0
  const hasRemote = status.remote.url.trim().length > 0
  const canCommit = status.changes.length > 0 && message.trim().length > 0 && busy === null
  const canPull = hasRemote && !dirty && busy === null

  const handleError = useCallback((err: Error) => {
    if (isUnrelatedHistoryError(err)) {
      onUnrelated()
      return
    }
    toast.error(err.message)
  }, [onUnrelated])

  const handleCommit = useCallback((push: boolean) => {
    const kind = push ? 'commitPush' : 'commit'
    setBusy(kind)
    gitApi.commit(projectId, message.trim(), push)
      .then((rsp) => {
        toast.success(translate('files.gitCommitSuccess').replace('{hash}', rsp.hash.slice(0, 7)))
        if (rsp.pushed) toast.success(translate('files.gitPushSuccess'))
        setMessage('')
        onMutated()
      })
      .catch(handleError)
      .finally(() => setBusy(null))
  }, [gitApi, projectId, message, handleError, onMutated])

  const handlePush = useCallback(() => {
    setBusy('push')
    gitApi.push(projectId)
      .then(() => {
        toast.success(translate('files.gitPushSuccess'))
        onMutated()
      })
      .catch(handleError)
      .finally(() => setBusy(null))
  }, [gitApi, projectId, handleError, onMutated])

  const handlePull = useCallback(() => {
    setBusy('pull')
    gitApi.pull(projectId)
      .then(() => {
        toast.success(translate('files.gitPullSuccess'))
        onMutated({ pulled: true })
      })
      .catch(handleError)
      .finally(() => setBusy(null))
  }, [gitApi, projectId, handleError, onMutated])

  return (
    <div className="border-t border-border px-5 py-4">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-xs font-medium text-text-2">
          {t('files.gitChanges').replace('{n}', String(status.changes.length))}
        </h3>
        <div className="flex items-center gap-1">
          <button
            onClick={handlePull}
            disabled={!canPull}
            title={!hasRemote ? t('files.gitNoRemoteHint') : canPull ? t('files.gitPull') : t('files.gitPullDisabledDirty')}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {busy === 'pull' ? <Loader2 className="size-3.5 animate-spin" /> : <ArrowDownToLine className="size-3.5" />}
            {t('files.gitPull')}
          </button>
          <button
            onClick={handlePush}
            disabled={busy !== null || !hasRemote}
            title={hasRemote ? t('files.gitPush') : t('files.gitNoRemoteHint')}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {busy === 'push' ? <Loader2 className="size-3.5 animate-spin" /> : <ArrowUpFromLine className="size-3.5" />}
            {t('files.gitPush')}
          </button>
        </div>
      </div>

      {status.skipped.length > 0 && (
        <p className="mb-2 rounded-md bg-warning/10 px-2.5 py-1.5 text-xs text-warning">
          {t('files.gitSkippedLarge').replace('{n}', String(status.skipped.length))}
        </p>
      )}

      {status.changes.length === 0 ? (
        <p className="rounded-md border border-border px-3 py-4 text-center text-xs text-text-3">{t('files.gitNoChanges')}</p>
      ) : (
        <>
          <div className="max-h-52 space-y-0.5 overflow-auto">
            {status.changes.map((change: GitChange) => (
              <div
                key={change.path}
                className="flex items-center gap-2 rounded px-1 py-1 transition-colors hover:bg-bg-hover"
              >
                <span
                  title={changeBadgeLabel(change.status)}
                  className={cn('w-4 shrink-0 text-center font-mono text-xs font-semibold', changeBadgeClass(change.status))}
                >
                  {change.status}
                </span>
                <span className="min-w-0 flex-1 truncate font-mono text-xs text-text-1">{change.path}</span>
                {change.size !== undefined && (
                  <span className="shrink-0 text-[11px] text-text-3 tabular-nums">{formatSize(change.size)}</span>
                )}
              </div>
            ))}
          </div>
          <div className="mt-3 space-y-2">
            <input
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              maxLength={500}
              spellCheck={false}
              placeholder={t('files.gitCommitMessagePlaceholder')}
              className="w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
            />
            <div className="flex gap-2">
              <button
                onClick={() => handleCommit(false)}
                disabled={!canCommit}
                className="flex flex-1 items-center justify-center gap-1.5 rounded-md bg-bg-layer-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-hover disabled:cursor-not-allowed disabled:opacity-40"
              >
                {busy === 'commit' && <Loader2 className="size-3.5 animate-spin" />}
                {t('files.gitCommit')}
              </button>
              <button
                onClick={() => handleCommit(true)}
                disabled={!canCommit || !hasRemote}
                title={hasRemote ? undefined : t('files.gitNoRemoteHint')}
                className="flex flex-1 items-center justify-center gap-1.5 rounded-md bg-highlight py-1.5 text-xs text-highlight-fg transition-colors hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {busy === 'commitPush' && <Loader2 className="size-3.5 animate-spin" />}
                {t('files.gitCommitAndPush')}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

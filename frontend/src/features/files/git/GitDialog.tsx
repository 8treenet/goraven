import { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { AlertCircle, GitBranch, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useT } from '@/i18n'
import type { GitStatus } from '@/api/types'
import type { ProjectGitApi } from '@/api/project-git'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { GIT_CLONE_STATE } from './git-helpers'
import { GitInfoPanel } from './GitInfoPanel'
import { GitChangesPanel } from './GitChangesPanel'
import { GitHistoryPanel } from './GitHistoryPanel'

type GitTab = 'changes' | 'history' | 'config'

interface GitDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectName: string
  projectId: number
  gitApi: ProjectGitApi
  canConfig: boolean
  /** 提交/推送/拉取等成功后通知父级（刷新文件行角标等） */
  onMutated?: () => void
  /** 拉取成功后回调（父级刷新文件列表） */
  onAfterPull?: () => void
}

export function GitDialog(props: GitDialogProps) {
  const { open, onOpenChange, projectName, projectId, gitApi, canConfig, onMutated, onAfterPull } = props
  const t = useT()
  const [status, setStatus] = useState<GitStatus | null>(null)
  const [statusError, setStatusError] = useState<string | null>(null)
  const [tab, setTab] = useState<GitTab>('changes')
  const [historyMounted, setHistoryMounted] = useState(false)
  const [historyKey, setHistoryKey] = useState(0)
  const [retrying, setRetrying] = useState(false)
  const [unrelatedOpen, setUnrelatedOpen] = useState(false)

  const refresh = useCallback(() => {
    gitApi.getStatus(projectId)
      .then((rsp) => {
        setStatus(rsp)
        setStatusError(null)
      })
      .catch((err: Error) => setStatusError(err.message))
  }, [gitApi, projectId])

  // 打开或切换项目时重置并拉取
  useEffect(() => {
    if (!open) return
    setTab('changes')
    setHistoryMounted(false)
    setUnrelatedOpen(false)
    setStatus(null)
    setStatusError(null)
    refresh()
  }, [open, projectId, refresh])

  // 克隆进行中每 3 秒轮询直至终态
  useEffect(() => {
    if (!open || status?.cloneState !== GIT_CLONE_STATE.RUNNING) return
    const timer = setInterval(refresh, 3000)
    return () => clearInterval(timer)
  }, [open, status?.cloneState, refresh])

  const handleMutated = useCallback((opts?: { pulled?: boolean }) => {
    refresh()
    setHistoryKey((k) => k + 1)
    onMutated?.()
    if (opts?.pulled) onAfterPull?.()
  }, [refresh, onMutated, onAfterPull])

  const handleResolve = useCallback((action: 'merge' | 'local_only') => {
    setUnrelatedOpen(false)
    gitApi.resolveUnrelated(projectId, action)
      .then(() => {
        if (action === 'local_only') {
          toast.success(t('files.gitNotConnected'))
        }
        refresh()
        setHistoryKey((k) => k + 1)
      })
      .catch((err: Error) => toast.error(err.message))
  }, [gitApi, projectId, refresh, t])

  const handleRetryClone = useCallback(() => {
    setRetrying(true)
    gitApi.retryClone(projectId)
      .then(() => refresh())
      .catch((err: Error) => toast.error(err.message))
      .finally(() => setRetrying(false))
  }, [gitApi, projectId, refresh])

  const selectTab = useCallback((next: GitTab) => {
    setTab(next)
    if (next === 'history') setHistoryMounted(true)
  }, [])

  const cloning = status?.cloneState === GIT_CLONE_STATE.RUNNING
  const cloneFailed = status?.cloneState === GIT_CLONE_STATE.FAILED
  const showTabs = !!status?.enabled && !cloning

  const tabs = useMemo(() => ([
    { key: 'changes' as const, label: t('files.gitTabChanges') },
    { key: 'history' as const, label: t('files.gitTabHistory') },
    { key: 'config' as const, label: t('files.gitTabConfig') },
  ]), [t])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {/* 关闭即卸载内容：Radix 退出动画会让子面板在关闭后仍存活片刻，
          期间父组件置空的 projectId(=0) 会触发子面板重发请求，故不做退出动画 */}
      {open && (
        <DialogContent className="flex max-h-[85dvh] max-w-2xl flex-col p-0 max-sm:left-0 max-sm:top-0 max-sm:h-dvh max-sm:max-h-none max-sm:w-screen max-sm:max-w-none max-sm:translate-x-0 max-sm:translate-y-0 max-sm:rounded-none">
        <DialogHeader className="mb-0 shrink-0 px-5 pb-3 pr-10 pt-5 max-sm:pt-[max(1.25rem,env(safe-area-inset-top))]">
          <DialogTitle className="flex items-center gap-2">
            <GitBranch className="size-4 text-highlight" />
            <span className="truncate">{projectName}</span>
            <span className="text-xs font-normal text-text-3">Git</span>
          </DialogTitle>
          <DialogDescription className="flex items-center gap-2">
            {status && (
              <>
                {status.initialized && status.branch ? (
                  <>
                    <span>{t('files.gitBranchLabel')}: <span className="font-mono text-text-2">{status.branch}</span></span>
                    {status.ahead > 0 && <span className="text-warning">{t('files.gitAhead').replace('{n}', String(status.ahead))}</span>}
                    {status.behind > 0 && <span className="text-warning">{t('files.gitBehind').replace('{n}', String(status.behind))}</span>}
                  </>
                ) : status.enabled ? (
                  <span>{t('files.gitNotConnected')}</span>
                ) : (
                  <span>{t('files.gitDisabled')}</span>
                )}
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        {showTabs && (
          <div className="mx-5 flex shrink-0 gap-1 rounded-md border border-border p-0.5">
            {tabs.map((item) => (
              <button
                key={item.key}
                onClick={() => selectTab(item.key)}
                className={cn(
                  'flex-1 rounded px-2 py-1.5 text-xs transition-colors max-sm:py-2',
                  tab === item.key ? 'bg-bg-layer-3 font-medium text-text-1' : 'text-text-3 hover:text-text-1',
                )}
              >
                {item.label}
              </button>
            ))}
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto pb-2 max-sm:pb-[max(0.5rem,env(safe-area-inset-bottom))]">
          {statusError && (
            <div className="flex flex-col items-center gap-3 px-5 py-12">
              <AlertCircle className="size-8 text-destructive" />
              <p className="text-center text-sm text-text-2">{statusError}</p>
            </div>
          )}

          {!statusError && !status && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="size-5 animate-spin text-text-3" />
            </div>
          )}

          {!statusError && status && !status.enabled && (
            <p className="px-5 py-6 text-xs text-text-3">{t('files.gitEnableHint')}</p>
          )}

          {!statusError && status && status.enabled && (
            <>
              {cloning && (
                <div className="mx-5 mt-4 flex items-center gap-2 rounded-md bg-interactive-soft px-3 py-2.5 text-xs text-interactive">
                  <Loader2 className="size-3.5 animate-spin" />
                  {t('files.gitCloneRunning')}
                </div>
              )}
              {cloning && <GitInfoPanel status={status} />}

              {!cloning && (
                <>
                  {cloneFailed && (
                    <div className="mx-5 mt-4 rounded-md bg-destructive/10 px-3 py-2.5 text-xs text-destructive">
                      <p className="flex items-center gap-1.5">
                        <AlertCircle className="size-3.5" />
                        {t('files.gitCloneFailed')}
                      </p>
                      {status.cloneMessage && <p className="mt-1 break-all opacity-80">{status.cloneMessage}</p>}
                      {canConfig && (
                        <button
                          onClick={handleRetryClone}
                          disabled={retrying}
                          className="mt-2 flex items-center gap-1.5 rounded-md bg-bg-layer-3 px-2.5 py-1 text-xs text-text-1 transition-colors hover:bg-bg-hover disabled:opacity-50"
                        >
                          {retrying ? <Loader2 className="size-3 animate-spin" /> : <RefreshCw className="size-3" />}
                          {t('files.gitCloneRetry')}
                        </button>
                      )}
                    </div>
                  )}

                  <div className={tab === 'changes' ? '' : 'hidden'}>
                    <GitChangesPanel
                      gitApi={gitApi}
                      projectId={projectId}
                      status={status}
                      onMutated={handleMutated}
                      onUnrelated={() => setUnrelatedOpen(true)}
                    />
                  </div>

                  {historyMounted && (
                    <div className={tab === 'history' ? '' : 'hidden'}>
                      <GitHistoryPanel gitApi={gitApi} projectId={projectId} refreshKey={historyKey} />
                    </div>
                  )}

                  <div className={tab === 'config' ? '' : 'hidden'}>
                    <GitInfoPanel status={status} />
                  </div>
                </>
              )}
            </>
          )}
        </div>

        {unrelatedOpen && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-md bg-black/50 p-5 max-sm:rounded-none">
            <div className="w-full max-w-xs rounded-md border border-border bg-bg-layer-1 p-4 shadow-pop">
              <h4 className="text-sm font-semibold text-text-1">{t('files.gitUnrelatedTitle')}</h4>
              <p className="mt-1.5 text-xs text-text-3">{t('files.gitUnrelatedDesc')}</p>
              <div className="mt-3 flex flex-col gap-2">
                <button
                  onClick={() => handleResolve('merge')}
                  className="rounded-md bg-highlight py-1.5 text-xs text-highlight-fg transition-colors hover:opacity-90"
                >
                  {t('files.gitUnrelatedMerge')}
                </button>
                <button
                  onClick={() => handleResolve('local_only')}
                  className="rounded-md bg-bg-layer-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-hover"
                >
                  {t('files.gitUnrelatedLocalOnly')}
                </button>
                <button onClick={() => setUnrelatedOpen(false)} className="py-1 text-xs text-text-3 hover:text-text-1">
                  {t('common.cancel')}
                </button>
              </div>
            </div>
          </div>
        )}
        </DialogContent>
      )}
    </Dialog>
  )
}

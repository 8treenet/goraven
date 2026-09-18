import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { ChevronDown, ChevronRight, History, Loader2 } from 'lucide-react'
import { useT, t as translate } from '@/i18n'
import type { GitCommitInfo } from '@/api/types'
import type { ProjectGitApi } from '@/api/project-git'
import { formatTime } from '../file-helpers'
import { GIT_PAGE_SIZE } from './git-helpers'
import { GitDiffView } from './GitDiffView'

interface GitHistoryPanelProps {
  gitApi: ProjectGitApi
  projectId: number
  /** 自增以重载第一页（提交/推送后） */
  refreshKey: number
}

export function GitHistoryPanel({ gitApi, projectId, refreshKey }: GitHistoryPanelProps) {
  const t = useT()
  const [items, setItems] = useState<GitCommitInfo[]>([])
  const [loaded, setLoaded] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [diff, setDiff] = useState('')
  const [diffLoading, setDiffLoading] = useState(false)

  const loadFirstPage = useCallback(() => {
    gitApi.log(projectId, GIT_PAGE_SIZE, 0)
      .then((rsp) => {
        setItems(rsp.items || [])
        setLoaded(true)
      })
      .catch((err: Error) => {
        setLoaded(true)
        toast.error(err.message)
      })
  }, [gitApi, projectId])

  useEffect(() => {
    loadFirstPage()
  }, [loadFirstPage, refreshKey])

  const hasMore = loaded && items.length > 0 && items.length % GIT_PAGE_SIZE === 0

  const loadMore = useCallback(() => {
    setLoadingMore(true)
    gitApi.log(projectId, GIT_PAGE_SIZE, items.length)
      .then((rsp) => {
        setItems((prev) => [...prev, ...(rsp.items || [])])
      })
      .catch((err: Error) => {
        toast.error(err.message)
      })
      .finally(() => setLoadingMore(false))
  }, [gitApi, projectId, items.length])

  const toggleExpand = useCallback((commit: GitCommitInfo) => {
    if (expanded === commit.hash) {
      setExpanded(null)
      return
    }
    setExpanded(commit.hash)
    setDiff('')
    setDiffLoading(true)
    gitApi.diff(projectId, '', commit.hash)
      .then((rsp) => setDiff(rsp.diff))
      .catch((err: Error) => {
        setExpanded(null)
        toast.error(err.message)
      })
      .finally(() => setDiffLoading(false))
  }, [expanded, gitApi, projectId])

  return (
    <div className="border-t border-border px-5 py-4">
      <h3 className="mb-2 flex items-center gap-1.5 text-xs font-medium text-text-2">
        <History className="size-3.5" />
        {t('files.gitHistory')}
      </h3>
      {loaded && items.length === 0 && (
        <p className="py-3 text-center text-xs text-text-3">{t('files.gitNoHistory')}</p>
      )}
      <div className="space-y-1">
        {items.map((commit) => {
          const isOpen = expanded === commit.hash
          return (
            <div key={commit.hash} className="rounded-md border border-border">
              <button
                onClick={() => toggleExpand(commit)}
                className="flex w-full items-center gap-2 px-2.5 py-2 text-left transition-colors hover:bg-bg-hover"
              >
                {isOpen ? <ChevronDown className="size-3.5 shrink-0 text-text-3" /> : <ChevronRight className="size-3.5 shrink-0 text-text-3" />}
                <span className="shrink-0 font-mono text-xs text-highlight">{commit.shortHash}</span>
                <span className="min-w-0 flex-1 truncate text-xs text-text-1">{commit.message}</span>
                <span className="hidden shrink-0 text-[11px] text-text-3 sm:inline">{commit.author}</span>
                <span className="shrink-0 text-[11px] text-text-3 tabular-nums">{formatTime(commit.time)}</span>
              </button>
              {isOpen && (
                <div className="border-t border-border p-2">
                  {diffLoading ? (
                    <div className="flex items-center justify-center py-4">
                      <Loader2 className="size-4 animate-spin text-text-3" />
                    </div>
                  ) : (
                    <GitDiffView diff={diff} emptyHint={translate('files.gitDiffEmpty')} />
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>
      {hasMore && (
        <button
          onClick={loadMore}
          disabled={loadingMore}
          className="mt-2 flex w-full items-center justify-center gap-1.5 rounded-md border border-border py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-hover disabled:opacity-50"
        >
          {loadingMore && <Loader2 className="size-3 animate-spin" />}
          {t('files.gitLoadMore')}
        </button>
      )}
    </div>
  )
}

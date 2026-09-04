import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Play, Ellipsis, Trash2, AlertCircle, RefreshCw, HelpCircle, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { useT, t as translate } from '@/i18n'
import { cn } from '@/lib/utils'
import { useSidebarStore } from '@/stores/sidebar-store'
import { AutomationStatus } from '@/api/types'
import type { AutomationTaskItem } from '@/api/types'
import { getAutomationTasks, updateTaskStatus, deleteTask, executeTask } from '@/api/automation'
import { StatusBadge, scheduleLabel, nextRunText, Pagination } from './shared'

const PAGE_SIZE = 10

type PageState = 'loading' | 'data' | 'empty' | 'error'
type TabKey = 'all' | 'enabled' | 'disabled' | 'done'

const TABS: { key: TabKey; labelKey: 'automation.tabAll' | 'automation.tabEnabled' | 'automation.tabDisabled' | 'automation.tabDone' }[] = [
  { key: 'all', labelKey: 'automation.tabAll' },
  { key: 'enabled', labelKey: 'automation.tabEnabled' },
  { key: 'disabled', labelKey: 'automation.tabDisabled' },
  { key: 'done', labelKey: 'automation.tabDone' },
]

/** 各页签对应的后端状态过滤值，all 不筛选 */
const TAB_STATUS: Record<TabKey, number | undefined> = {
  all: undefined,
  enabled: AutomationStatus.Enabled,
  disabled: AutomationStatus.Disabled,
  done: AutomationStatus.Done,
}

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)
  const [state, setState] = useState<PageState>('loading')
  const [tasks, setTasks] = useState<AutomationTaskItem[]>([])
  const [tab, setTab] = useState<TabKey>('all')
  const [page, setPage] = useState(1)
  const [totalPage, setTotalPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [menuOpenId, setMenuOpenId] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AutomationTaskItem | null>(null)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [hintOpen, setHintOpen] = useState(false)

  const fetchTasks = useCallback((targetTab: TabKey, targetPage: number, silent = false) => {
    if (!silent) setState('loading')
    getAutomationTasks({ page: targetPage, pageSize: PAGE_SIZE, status: TAB_STATUS[targetTab] })
      .then((rsp) => {
        setTasks(rsp.list)
        setPage(rsp.page)
        setTotalPage(rsp.totalPage)
        setTotalCount(rsp.totalCount)
        setState(rsp.list.length > 0 ? 'data' : 'empty')
      })
      .catch(() => setState('error'))
  }, [])

  useEffect(() => {
    fetchTasks(tab, page)
  }, [tab, page, fetchTasks])

  const patchTask = useCallback((id: number, patch: Partial<AutomationTaskItem>) => {
    setTasks((prev) => prev.map((x) => (x.id === id ? { ...x, ...patch } : x)))
  }, [])

  const handleTabChange = useCallback((key: TabKey) => {
    setTab(key)
    setPage(1)
  }, [])

  const handleToggle = useCallback(async (task: AutomationTaskItem, checked: boolean) => {
    if (task.status === AutomationStatus.Done) {
      toast.error(translate('automation.doneCannotToggle'))
      return
    }
    const target = checked ? AutomationStatus.Enabled : AutomationStatus.Disabled
    setBusyId(task.id)
    try {
      await updateTaskStatus(task.id, target)
      patchTask(task.id, { status: target })
      toast.success(target === AutomationStatus.Enabled ? translate('automation.taskEnabled') : translate('automation.taskDisabled'))
      if (target === AutomationStatus.Enabled) fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusyId(null)
    }
  }, [patchTask, fetchTasks, tab, page])

  const handleExecute = useCallback(async (task: AutomationTaskItem) => {
    setBusyId(task.id)
    try {
      await executeTask(task.id)
      toast.success(translate('automation.executionStarted'))
      fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusyId(null)
    }
  }, [fetchTasks, tab, page])

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return
    try {
      await deleteTask(deleteTarget.id)
      setDeleteTarget(null)
      setMenuOpenId(null)
      toast.success(translate('automation.deleted'))
      fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.deleteFailed'))
    }
  }, [deleteTarget, fetchTasks, tab, page])

  const handleRowClick = useCallback((id: number) => {
    navigate(`/automation/${id}`)
  }, [navigate])

  return (
    <div>
      {/* Sticky toolbar */}
      <div className="sticky top-0 z-10 flex h-10 items-center gap-2 border-b border-border-custom bg-bg-base px-4">
        <button
          onClick={openMobile}
          aria-label={t('sidebar.menu')}
          className="shrink-0 text-text-1 transition-colors md:hidden"
        >
          <Menu className="size-4" />
        </button>
        <div className="flex items-center gap-1.5">
          <h1 className="text-[18px] font-semibold text-text-1">{t('automation.title')}</h1>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label={t('automation.createHint')}
                onClick={() => setHintOpen(true)}
                className="flex items-center text-amber-500 transition-colors hover:text-amber-400"
              >
                <HelpCircle className="size-3.5" />
              </button>
            </TooltipTrigger>
            <TooltipContent className="max-w-72">{t('automation.createHint')}</TooltipContent>
          </Tooltip>
        </div>
      </div>

      {/* Status tabs */}
      <div className="flex gap-1 border-b border-border-custom px-4 pt-2">
        {TABS.map(({ key, labelKey }) => (
          <button
            key={key}
            onClick={() => handleTabChange(key)}
            className={cn(
              'relative px-2.5 py-1.5 text-[13px] transition-colors md:px-3',
              tab === key
                ? 'font-medium text-text-1 after:absolute after:bottom-[-1px] after:left-0 after:right-0 after:h-[2px] after:rounded-full after:bg-highlight'
                : 'text-text-3 hover:text-text-2',
            )}
          >
            {t(labelKey)}
          </button>
        ))}
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}
      {state === 'error' && <ErrorView onRetry={() => fetchTasks(tab, page)} />}
      {state === 'empty' && (tab === 'all' ? <EmptyView /> : (
        <p className="px-4 py-10 text-center text-sm text-text-3">{t('automation.emptyTasks')}</p>
      ))}
      {state === 'data' && (
        <div>
          {tasks.map((task) => {
            const done = task.status === AutomationStatus.Done
            const next = nextRunText(task, t)
            const meta: string[] = [scheduleLabel(task, t)]
            if (task.aiModelName) meta.push(`${t('automation.model')} · ${task.aiModelName}`)
            if (task.personaName) meta.push(`${t('automation.persona')} · ${task.personaName}`)
            if (task.mcpNames.length > 0) meta.push(t('automation.mcpCount').replace('{n}', String(task.mcpNames.length)))
            if (task.skillNames.length > 0) meta.push(t('automation.skillCount').replace('{n}', String(task.skillNames.length)))

            return (
              <div
                key={task.id}
                className="flex flex-wrap items-center gap-x-3 gap-y-2 border-b border-border-custom px-4 py-3 transition-colors hover:bg-bg-hover md:flex-nowrap md:gap-3"
              >
                <button
                  className="flex w-full min-w-0 flex-col items-start gap-1.5 text-left md:w-auto md:flex-1"
                  onClick={() => handleRowClick(task.id)}
                >
                  <span className="flex w-full min-w-0 items-center gap-2">
                    <span className="min-w-0 flex-1 truncate text-sm font-semibold text-text-1">{task.title}</span>
                    <StatusBadge status={task.status} t={t} className="md:hidden" />
                  </span>
                  <span className="flex flex-wrap items-center gap-1.5 text-[11px] text-text-3">
                    {meta.map((m) => (
                      <span key={m} className="rounded border border-border-custom px-1.5 py-0.5 text-text-2">
                        {m}
                      </span>
                    ))}
                  </span>
                </button>

                <div className="flex min-w-0 flex-1 flex-col items-start gap-1.5 md:flex-none md:shrink-0 md:items-end">
                  <span className="text-[11px] text-text-3">
                    {next.label}：{next.value}
                  </span>
                  <StatusBadge status={task.status} t={t} className="hidden md:inline-flex" />
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <button
                    title={t('automation.executeNow')}
                    disabled={done || busyId === task.id}
                    onClick={(e) => { e.stopPropagation(); handleExecute(task) }}
                    className={cn(
                      'flex h-7 w-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1',
                      (done || busyId === task.id) && 'cursor-not-allowed opacity-35 hover:bg-transparent hover:text-text-3',
                    )}
                  >
                    <Play className="size-3.5" />
                  </button>

                  <Switch
                    size="sm"
                    checked={task.status === AutomationStatus.Enabled}
                    disabled={done || busyId === task.id}
                    onCheckedChange={(checked) => handleToggle(task, checked)}
                  />

                  <div className="relative">
                    <button
                      onClick={(e) => { e.stopPropagation(); setMenuOpenId(menuOpenId === task.id ? null : task.id) }}
                      className="flex h-7 w-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
                    >
                      <Ellipsis className="size-4" />
                    </button>
                    {menuOpenId === task.id && (
                      <>
                        <div className="fixed inset-0 z-10" onClick={() => setMenuOpenId(null)} />
                        <div className="absolute right-0 top-full z-20 mt-1 w-32 rounded-md border border-border-custom bg-bg-layer-2 py-1 shadow-pop">
                          <button
                            className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                            onClick={(e) => { e.stopPropagation(); setMenuOpenId(null); setDeleteTarget(task) }}
                          >
                            <Trash2 className="size-3" />
                            {t('automation.delete')}
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
          <Pagination
            page={page}
            totalPages={totalPage}
            totalCount={totalCount}
            onPageChange={setPage}
          />
        </div>
      )}

      {/* Hint dialog */}
      <Dialog open={hintOpen} onOpenChange={setHintOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <HelpCircle className="size-4 text-amber-500" />
              {t('automation.title')}
            </DialogTitle>
            <DialogDescription>{t('automation.createHint')}</DialogDescription>
          </DialogHeader>
          <div className="mt-4 flex justify-end">
            <Button variant="outline" size="default" onClick={() => setHintOpen(false)}>
              {t('common.close')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('automation.confirmDeleteTitle')}</DialogTitle>
            <DialogDescription>
              {t('automation.confirmDeleteDesc').replace('{name}', deleteTarget?.title ?? '')}
            </DialogDescription>
          </DialogHeader>
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" size="default" onClick={() => setDeleteTarget(null)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" size="default" onClick={handleDeleteConfirm}>
              {t('automation.delete')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

/* ============================================
   Sub-views
   ============================================ */

function Skeleton() {
  return (
    <div>
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="flex items-center gap-3 border-b border-border-custom px-4 py-3">
          <div className="flex-1 space-y-2">
            <span className="block h-4 w-28 rounded bg-bg-layer-3 animate-pulse" />
            <span className="block h-3.5 w-64 rounded bg-bg-layer-3 animate-pulse" />
          </div>
          <span className="h-5 w-14 rounded-full bg-bg-layer-3 animate-pulse" />
          <span className="h-7 w-7 rounded-md bg-bg-layer-3 animate-pulse" />
        </div>
      ))}
    </div>
  )
}

function EmptyView() {
  const t = useT()
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <AlertCircle className="size-8 text-text-muted" />
      <p className="mt-3 text-sm text-text-2">{t('automation.emptyTasks')}</p>
      <p className="mt-1 text-xs text-text-3">{t('automation.emptyTasksHint')}</p>
    </div>
  )
}

function ErrorView({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <AlertCircle className="size-8 text-text-muted" />
      <p className="mt-3 text-sm text-text-2">{t('common.loadFailed')}</p>
      <Button variant="outline" size="default" className="mt-4" onClick={onRetry}>
        <RefreshCw className="size-4" />
        {t('common.retry')}
      </Button>
    </div>
  )
}
